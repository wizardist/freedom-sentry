package suppressor

import (
	"freedom-sentry/mediawiki"
	"golang.org/x/exp/slices"
	"log"
	"sync"
	"time"
)

type batchingSuppressor struct {
	period     time.Duration
	size       int
	suppressor RevisionSuppressor

	initOnce sync.Once // Constraint to initialize everything below safely
	buffer   []mediawiki.Revision
	lock     sync.Mutex

	drainRequest      chan bool
	forceDrainRequest chan bool
}

func (b *batchingSuppressor) SuppressRevisions(revs []mediawiki.Revision) error {
	b.initOnce.Do(b.init)

	if len(revs) == 0 {
		return nil
	}

	withLock(&b.lock, func() {
		slices.Grow(b.buffer, len(revs))
		for _, rev := range revs {
			b.buffer = append(b.buffer, rev)
		}
	})

	b.drainRequest <- true

	return nil
}

func (b *batchingSuppressor) drainBuffer() error {
	var batch []mediawiki.Revision

	if len(b.buffer) >= b.size {
		batch, b.buffer = b.buffer[:b.size], b.buffer[b.size:]
		return b.suppressor.SuppressRevisions(batch)
	}

	return nil
}

func (b *batchingSuppressor) forceDrainBuffer() error {
	var batch []mediawiki.Revision

	delimiter := b.size
	if len(b.buffer) < delimiter {
		delimiter = len(b.buffer)
	}

	batch, b.buffer = b.buffer[:delimiter], b.buffer[delimiter:]

	if len(batch) == 0 {
		return nil
	}

	return b.suppressor.SuppressRevisions(batch)
}

func (b *batchingSuppressor) init() {
	if b.drainRequest != nil {
		return
	}

	b.drainRequest = make(chan bool)
	b.forceDrainRequest = make(chan bool)

	if b.period > 0 {
		// Production code, tests use period = 0 and ping forceDrainRequest manually
		go func() {
			forceDrainScheduler := time.NewTicker(b.period)

			for {
				select {
				case <-forceDrainScheduler.C:
					b.forceDrainRequest <- true
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case <-b.forceDrainRequest:
				withLockErr(&b.lock, b.forceDrainBuffer)
			case <-b.drainRequest:
				withLockErr(&b.lock, b.drainBuffer)
			}
		}
	}()
}

func withLock(lk sync.Locker, fn func()) {
	lk.Lock()
	defer lk.Unlock()

	fn()
}

func withLockErr(lk sync.Locker, fn func() error) {
	lk.Lock()
	defer lk.Unlock()

	err := fn()
	if err != nil {
		log.Println(err)
	}
}
