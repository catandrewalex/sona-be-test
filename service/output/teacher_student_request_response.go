package output

import (
	"sonamusica-backend/app-service/identity"
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
	ID teaching.TeacherID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetTeacherResponse struct {
	Data    teaching.Teacher `json:"data"`
	Message string           `json:"message,omitempty"`
}

func (r GetTeacherRequest) Validate() errs.ValidationError {
	return nil
}

type InsertTeachersRequest struct {
	UserIDs []identity.UserID `json:"userIDs"`
}
type InsertTeachersResponse struct {
	Data    []teaching.TeacherID `json:"data"`
	Message string               `json:"message,omitempty"`
}

func (r InsertTeachersRequest) Validate() errs.ValidationError {
	return nil
}

type InsertTeachersWithNewUsersRequest struct {
	Params []InsertUserRequestParam `json:"params"`
}
type InsertTeachersWithNewUsersResponse struct {
	InsertTeachersResponse
}

func (r InsertTeachersWithNewUsersRequest) Validate() errs.ValidationError {
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
	ID teaching.StudentID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetStudentResponse struct {
	Data    teaching.Student `json:"data"`
	Message string           `json:"message,omitempty"`
}

func (r GetStudentRequest) Validate() errs.ValidationError {
	return nil
}
