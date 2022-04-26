package suppressor

import (
	"log"
	"strings"
)

type PageRepository interface {
	GetAllSuppressed() ([]string, error)
}

func NewPageRepository(revRepo RevisionRepository, listName string) PageRepository {
	return &pageRepoImpl{
		revRepo:  revRepo,
		listName: listName,
	}
}

type pageRepoImpl struct {
	revRepo  RevisionRepository
	listName string
}

func (p pageRepoImpl) GetAllSuppressed() ([]string, error) {
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
