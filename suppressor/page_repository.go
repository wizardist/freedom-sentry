package suppressor

import (
	"log"
	"strings"
	"time"
)

type SuppressedPageRepository interface {
	GetAll() ([]string, error)
}

func NewPageRepository(revRepo RevisionRepository, listName string) (SuppressedPageRepository, chan bool) {
	repo := &cachingSuppressedPageRepoImpl{
		repo: &suppressedPageRepoImpl{
			revRepo:  revRepo,
			listName: listName,
		},
	}

	go func() {
		for {
			select {
			case <-repo.purgeChan:
				repo.timestamp = time.Time{}
			}
		}
	}()

	return repo, repo.purgeChan
}

type suppressedPageRepoImpl struct {
	revRepo  RevisionRepository
	listName string
}

func (p suppressedPageRepoImpl) GetAll() ([]string, error) {
	suppressedPagesStr, err := p.revRepo.GetLatestPageContent(p.listName)
	if err != nil {
		log.Println("failed to retrieve the list of suppressed pages")
		return nil, err
	}

	suppressedPages := rawPageListToSlice(suppressedPagesStr)
	return suppressedPages, nil
}

func rawPageListToSlice(suppressedPagesStr string) []string {
	return strings.Split(suppressedPagesStr, "\n")
}

type cachingSuppressedPageRepoImpl struct {
	list      []string
	timestamp time.Time
	repo      SuppressedPageRepository

	purgeChan chan bool
}

func (c *cachingSuppressedPageRepoImpl) GetAll() ([]string, error) {
	if !c.timestamp.IsZero() || time.Now().Sub(c.timestamp) < 24*time.Hour {
		return c.list, nil
	}

	list, err := c.repo.GetAll()
	if err != nil {
		c.timestamp = time.Time{}
		return nil, err
	}

	c.list = list
	c.timestamp = time.Now()

	return list, nil
}
