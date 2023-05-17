package output

import (
	"fmt"

	"sonamusica-backend/errs"
)

const (
	Default_MaxPage           = 50
	Default_MaxResultsPerPage = 100
)

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
	} else if maxResultsPerPage > 0 && r.Page > maxResultsPerPage {
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
