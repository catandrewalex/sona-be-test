package output

import (
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/errs"
)

const (
	MaxPage_GetTeachers           = Default_MaxPage
	MaxResultsPerPage_GetTeachers = Default_MaxResultsPerPage

	MaxPage_GetStudents           = Default_MaxPage
	MaxResultsPerPage_GetStudents = Default_MaxResultsPerPage
)

type GetTeachersRequest struct {
	PaginationRequest
}
type GetTeachersResponse struct {
	Data    GetTeachersResult `json:"data"`
	Message string            `json:"message,omitempty"`
}
type GetTeachersResult struct {
	Results []teaching.Teacher `json:"results"`
	PaginationResponse
}

func (r GetTeachersRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetTeachers, MaxResultsPerPage_GetTeachers); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetStudentsRequest struct {
	PaginationRequest
}
type GetStudentsResponse struct {
	Data    GetStudentsResult `json:"data"`
	Message string            `json:"message,omitempty"`
}
type GetStudentsResult struct {
	Results []teaching.Student `json:"results"`
	PaginationResponse
}

func (r GetStudentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetTeachers, MaxResultsPerPage_GetTeachers); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}
