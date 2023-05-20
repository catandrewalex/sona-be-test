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

type GetTeacherRequest struct {
	ID teaching.TeacherID `json:"id"`
}
type GetTeacherResponse struct {
	Data    teaching.Teacher `json:"data"`
	Message string           `json:"message,omitempty"`
}

func (r GetTeacherRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.ID == teaching.TeacherID_None {
		errorDetail["id"] = "id cannot be empty"
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

type GetStudentRequest struct {
	ID teaching.StudentID `json:"id"`
}
type GetStudentResponse struct {
	Data    teaching.Student `json:"data"`
	Message string           `json:"message,omitempty"`
}

func (r GetStudentRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.ID == teaching.StudentID_None {
		errorDetail["id"] = "id cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}
