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

// ============================== TEACHER ==============================

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
	Data []InsertTeachersRequestParam `json:"data"`
}
type InsertTeachersRequestParam struct {
	UserID identity.UserID `json:"userID"`
}
type InsertTeachersResponse struct {
	Data    InsertTeacherResult `json:"data"`
	Message string              `json:"message,omitempty"`
}
type InsertTeacherResult struct {
	Results []teaching.Teacher `json:"results"`
}

func (r InsertTeachersRequest) Validate() errs.ValidationError {
	return nil
}

type InsertTeachersWithNewUsersRequest struct {
	Data []InsertUserRequestParam `json:"data"`
}
type InsertTeachersWithNewUsersResponse struct {
	InsertTeachersResponse
}

func (r InsertTeachersWithNewUsersRequest) Validate() errs.ValidationError {
	return nil
}

type DeleteTeachersRequest struct {
	Data []DeleteTeachersRequestParam `json:"data"`
}
type DeleteTeachersRequestParam struct {
	TeacherID teaching.TeacherID `json:"teacherID"`
}
type DeleteTeachersResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteTeachersRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== STUDENT ==============================

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

type InsertStudentsRequest struct {
	Data []InsertStudentsRequestParam `json:"data"`
}
type InsertStudentsRequestParam struct {
	UserID identity.UserID `json:"userID"`
}
type InsertStudentsResponse struct {
	Data    InsertStudentResult `json:"data"`
	Message string              `json:"message,omitempty"`
}
type InsertStudentResult struct {
	Results []teaching.Student `json:"results"`
}

func (r InsertStudentsRequest) Validate() errs.ValidationError {
	return nil
}

type InsertStudentsWithNewUsersRequest struct {
	Data []InsertUserRequestParam `json:"data"`
}
type InsertStudentsWithNewUsersResponse struct {
	InsertStudentsResponse
}

func (r InsertStudentsWithNewUsersRequest) Validate() errs.ValidationError {
	return nil
}

type DeleteStudentsRequest struct {
	Data []DeleteStudentsRequestParam `json:"data"`
}
type DeleteStudentsRequestParam struct {
	StudentID teaching.StudentID `json:"studentID"`
}
type DeleteStudentsResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteStudentsRequest) Validate() errs.ValidationError {
	return nil
}
