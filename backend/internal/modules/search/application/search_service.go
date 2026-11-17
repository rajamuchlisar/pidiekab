package application

import (
	"context"

	"pidiekab/internal/modules/search/domain"
	"pidiekab/internal/modules/search/infrastructure/postgres"
)

type SearchService struct {
	repo *postgres.SearchRepository
}

func NewSearchService(repo *postgres.SearchRepository) *SearchService {
	return &SearchService{repo: repo}
}

type SearchInput struct {
	Locale       string
	Query        string
	CategorySlug *string
	Page         int
	PerPage      int
}

func (s *SearchService) Search(ctx context.Context, input SearchInput) (*domain.SearchResponse, error) {
	locale, err := domain.ValidateLocale(input.Locale)
	if err != nil {
		return nil, err
	}

	query, err := domain.ValidateQuery(input.Query)
	if err != nil {
		return nil, err
	}

	page := input.Page
	if page < 1 {
		page = 1
	}

	perPage := input.PerPage
	if perPage < 1 || perPage > 50 {
		perPage = 12
	}

	params := postgres.SearchParams{
		Locale:       locale,
		Query:        query,
		CategorySlug: input.CategorySlug,
		Page:         page,
		PerPage:      perPage,
	}

	items, total, err := s.repo.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	return &domain.SearchResponse{
		Data: items,
		Meta: domain.SearchMeta{
			Locale:     locale,
			Query:      query,
			Category:   input.CategorySlug,
			Page:       page,
			PerPage:    perPage,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

