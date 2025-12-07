package suppressor

import (
	"log/slog"
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

	repo.purgeChan = make(chan bool)

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
		slog.Error("failed to retrieve suppression list", "list_name", p.listName, "error", err)
		return nil, err
	}

	lines := strings.Split(suppressedPagesStr, "\n")
	suppressedPages := rawPageListToSlice(suppressedPagesStr)
	slog.Debug("parsed suppression list",
		"list_name", p.listName,
		"total_lines", len(lines),
		"valid_pages", len(suppressedPages))
	return suppressedPages, nil
}

func rawPageListToSlice(suppressedPagesStr string) []string {
	lines := strings.Split(suppressedPagesStr, "\n")

	list := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		list = append(list, trimmed)
	}

	return list
}

type cachingSuppressedPageRepoImpl struct {
	list      []string
	timestamp time.Time
	repo      SuppressedPageRepository

	purgeChan chan bool
}

func (c *cachingSuppressedPageRepoImpl) GetAll() ([]string, error) {
	if !c.timestamp.IsZero() || time.Now().Sub(c.timestamp) < 24*time.Hour {
		slog.Debug("using cached suppression list", "count", len(c.list))
		return c.list, nil
	}

	list, err := c.repo.GetAll()
	if err != nil {
		c.timestamp = time.Time{}
		return nil, err
	}

	c.list = list
	c.timestamp = time.Now()
	slog.Info("refreshed suppression list cache", "count", len(list))

	return list, nil
}
