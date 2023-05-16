package util

type PaginationSpec struct {
	Page           int
	ResultsPerPage int
}

const (
	default_Page           = 1
	default_ResultsPerPage = 1000
)

func (s *PaginationSpec) SetDefaultOnInvalidValues() {
	if s.Page <= 0 {
		s.Page = default_Page
	}
	if s.ResultsPerPage <= 0 {
		s.ResultsPerPage = default_ResultsPerPage
	}
}

func (s PaginationSpec) GetLimitAndOffset() (int, int) {
	limit := s.ResultsPerPage
	offset := (s.Page - 1) * s.ResultsPerPage
	return limit, offset
}

type PaginationResult struct {
	TotalPages   int
	TotalResults int
	CurrentPage  int
}

func NewPaginationResult(totalResults int, resultsPerPage int, currentPage int) *PaginationResult {
	totalPages := totalResults / resultsPerPage
	remnants := totalResults % resultsPerPage
	if remnants > 0 {
		totalPages += 1
	}
	return &PaginationResult{
		TotalPages:   totalPages,
		TotalResults: totalResults,
		CurrentPage:  currentPage,
	}
}
