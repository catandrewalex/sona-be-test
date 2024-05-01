package output

import (
	"fmt"
	"time"

	"sonamusica-backend/app-service/util"
	"sonamusica-backend/errs"
)

const (
	Default_MaxPage           = 50
	Default_MaxResultsPerPage = 10000
)

type YearMonthFilterType string

const (
	YearMonthFilterType_Standard YearMonthFilterType = "STANDARD"
	YearMonthFilterType_Salary   YearMonthFilterType = "SALARY"
)

type ErrorResponse struct {
	Errors  map[string]string `json:"errors"`
	Message string            `json:"message,omitempty"`
}

type PaginationRequest struct {
	Page           int `json:"page"`
	ResultsPerPage int `json:"resultsPerPage"`
}

func (r PaginationRequest) Validate(maxPage int, maxResultsPerPage int) errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Page <= 0 {
		errorDetail["page"] = "page must be greater than 0"
	} else if maxPage > 0 && r.Page > maxPage {
		errorDetail["page"] = fmt.Sprintf("page must be less than %d", maxPage)
	}
	if maxResultsPerPage > 0 && r.Page > maxResultsPerPage {
		errorDetail["resultsPerPage"] = "resultsPerPage must be greater than 0"
	} else if maxResultsPerPage > 0 && r.ResultsPerPage > maxResultsPerPage {
		errorDetail["resultsPerPage"] = fmt.Sprintf("resultsPerPage must be less than %d", maxPage)
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type PaginationResponse struct {
	TotalPages   int `json:"totalPages"`
	TotalResults int `json:"totalResults"`
	CurrentPage  int `json:"currentPage"`
}

func NewPaginationResponse(paginationResult util.PaginationResult) PaginationResponse {
	return PaginationResponse{
		TotalPages:   paginationResult.TotalPages,
		TotalResults: paginationResult.TotalResults,
		CurrentPage:  paginationResult.CurrentPage,
	}
}

type TimeFilter struct {
	StartDatetime time.Time `json:"startDatetime,omitempty"`
	EndDatetime   time.Time `json:"endDatetime,omitempty"`
}

func (r TimeFilter) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if !r.EndDatetime.IsZero() && r.StartDatetime.After(r.EndDatetime) {
		errorDetail["startDatetime"] = "startDatetime must be less than endDatetime"
		errorDetail["endDatetime"] = "endDatetime must be greater than startDatetime"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type YearMonthFilter struct {
	Year  int `json:"year,omitempty"`
	Month int `json:"month,omitempty"`
}

func (d YearMonthFilter) Validate() errs.ValidationError {
	// we allow empty YearMonthFilter
	if d.Year == 0 && d.Month == 0 {
		return nil
	}

	errorDetail := make(errs.ValidationErrorDetail, 0)
	if d.Year < 1960 || d.Year > 2100 {
		errorDetail["year"] = "year must be in between 1960-2100"
	}
	if d.Month < 1 || d.Month > 12 {
		errorDetail["month"] = "month must be in between 1-12"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

func (d YearMonthFilter) ToTimeFilter(filterType YearMonthFilterType) TimeFilter {
	if d.Year == 0 && d.Month == 0 {
		return TimeFilter{}
	}

	timeFilter := TimeFilter{}
	switch filterType {
	case YearMonthFilterType_Standard:
		timeFilter = TimeFilter{
			StartDatetime: time.Date(d.Year, time.Month(d.Month), 1, 0, 0, 0, 0, time.UTC),
			EndDatetime:   time.Date(d.Year, time.Month(d.Month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1),
		}
	case YearMonthFilterType_Salary:
		timeFilter = TimeFilter{
			StartDatetime: time.Date(d.Year, time.Month(d.Month), 26, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0),
			EndDatetime:   time.Date(d.Year, time.Month(d.Month), 25, 0, 0, 0, 0, time.UTC),
		}
	}
	return timeFilter
}
