package output

import (
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/errs"
)

const (
	MaxPage_GetInstruments           = Default_MaxPage
	MaxResultsPerPage_GetInstruments = Default_MaxResultsPerPage

	MaxPage_GetGrades           = Default_MaxPage
	MaxResultsPerPage_GetGrades = Default_MaxResultsPerPage

	MaxPage_GetCourses           = Default_MaxPage
	MaxResultsPerPage_GetCourses = Default_MaxResultsPerPage
)

// ============================== INSTRUMENT ==============================

type GetInstrumentsRequest struct {
	PaginationRequest
}
type GetInstrumentsResponse struct {
	Data    GetInstrumentsResult `json:"data"`
	Message string               `json:"message,omitempty"`
}
type GetInstrumentsResult struct {
	Results []teaching.Instrument `json:"results"`
	PaginationResponse
}

func (r GetInstrumentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetInstruments, MaxResultsPerPage_GetInstruments); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetInstrumentRequest struct {
	ID teaching.InstrumentID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetInstrumentResponse struct {
	Data    teaching.Instrument `json:"data"`
	Message string              `json:"message,omitempty"`
}

func (r GetInstrumentRequest) Validate() errs.ValidationError {
	return nil
}

type InsertInstrumentsRequest struct {
	Data []InsertInstrumentsRequestParam `json:"data"`
}
type InsertInstrumentsRequestParam struct {
	Name string `json:"name"`
}
type InsertInstrumentsResponse struct {
	Data    UpsertInstrumentResult `json:"data"`
	Message string                 `json:"message,omitempty"`
}

func (r InsertInstrumentsRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateInstrumentsRequest struct {
	Data []UpdateInstrumentsRequestParam `json:"data"`
}
type UpdateInstrumentsRequestParam struct {
	ID teaching.InstrumentID `json:"id"`
	InsertInstrumentsRequestParam
}
type UpdateInstrumentsResponse struct {
	Data    UpsertInstrumentResult `json:"data"`
	Message string                 `json:"message,omitempty"`
}

func (r UpdateInstrumentsRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertInstrumentResult struct {
	Results []teaching.Instrument `json:"results"`
}

// ============================== GRADE ==============================

type GetGradesRequest struct {
	PaginationRequest
}
type GetGradesResponse struct {
	Data    GetGradesResult `json:"data"`
	Message string          `json:"message,omitempty"`
}
type GetGradesResult struct {
	Results []teaching.Grade `json:"results"`
	PaginationResponse
}

func (r GetGradesRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetGrades, MaxResultsPerPage_GetGrades); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetGradeRequest struct {
	ID teaching.GradeID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetGradeResponse struct {
	Data    teaching.Grade `json:"data"`
	Message string         `json:"message,omitempty"`
}

func (r GetGradeRequest) Validate() errs.ValidationError {
	return nil
}

type InsertGradesRequest struct {
	Data []InsertGradesRequestParam `json:"data"`
}
type InsertGradesRequestParam struct {
	Name string `json:"name"`
}
type InsertGradesResponse struct {
	Data    UpsertGradeResult `json:"data"`
	Message string            `json:"message,omitempty"`
}

func (r InsertGradesRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateGradesRequest struct {
	Data []UpdateGradesRequestParam `json:"data"`
}
type UpdateGradesRequestParam struct {
	ID teaching.GradeID `json:"id"`
	InsertGradesRequestParam
}
type UpdateGradesResponse struct {
	Data    UpsertGradeResult `json:"data"`
	Message string            `json:"message,omitempty"`
}

func (r UpdateGradesRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertGradeResult struct {
	Results []teaching.Grade `json:"results"`
}

// ============================== COURSE ==============================

type GetCoursesRequest struct {
	PaginationRequest
}
type GetCoursesResponse struct {
	Data    GetCoursesResult `json:"data"`
	Message string           `json:"message,omitempty"`
}
type GetCoursesResult struct {
	Results []teaching.Course `json:"results"`
	PaginationResponse
}

func (r GetCoursesRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetCourses, MaxResultsPerPage_GetCourses); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetCourseRequest struct {
	ID teaching.CourseID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetCourseResponse struct {
	Data    teaching.Course `json:"data"`
	Message string          `json:"message,omitempty"`
}

func (r GetCourseRequest) Validate() errs.ValidationError {
	return nil
}

type InsertCoursesRequest struct {
	Data []InsertCoursesRequestParam `json:"data"`
}
type InsertCoursesRequestParam struct {
	InstrumentID          teaching.InstrumentID `json:"instrumentId"`
	GradeID               teaching.GradeID      `json:"gradeId"`
	DefaultFee            int64                 `json:"defaultFee"`
	DefaultDurationMinute int32                 `json:"defaultDurationMinute"`
}
type InsertCoursesResponse struct {
	Data    UpsertCourseResult `json:"data"`
	Message string             `json:"message,omitempty"`
}

func (r InsertCoursesRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateCoursesRequest struct {
	Data []UpdateCoursesRequestParam `json:"data"`
}
type UpdateCoursesRequestParam struct {
	ID                    teaching.CourseID `json:"id"`
	DefaultFee            int64             `json:"defaultFee"`
	DefaultDurationMinute int32             `json:"defaultDurationMinute"`
}
type UpdateCoursesResponse struct {
	Data    UpsertCourseResult `json:"data"`
	Message string             `json:"message,omitempty"`
}

func (r UpdateCoursesRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertCourseResult struct {
	Results []teaching.Course `json:"results"`
}
