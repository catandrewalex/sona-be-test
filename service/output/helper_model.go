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
