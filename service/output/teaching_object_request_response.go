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

	MaxPage_GetClasses           = Default_MaxPage
	MaxResultsPerPage_GetClasses = Default_MaxResultsPerPage

	MaxPage_GetTeacherSpecialFees           = Default_MaxPage
	MaxResultsPerPage_GetTeacherSpecialFees = Default_MaxResultsPerPage
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
	InstrumentID teaching.InstrumentID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
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
	InstrumentID teaching.InstrumentID `json:"instrumentId"`
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

type DeleteInstrumentsRequest struct {
	Data []DeleteInstrumentsRequestParam `json:"data"`
}
type DeleteInstrumentsRequestParam struct {
	InstrumentID teaching.InstrumentID `json:"instrumentId"`
}
type DeleteInstrumentsResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteInstrumentsRequest) Validate() errs.ValidationError {
	return nil
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
	GradeID teaching.GradeID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
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
	GradeID teaching.GradeID `json:"gradeId"`
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

type DeleteGradesRequest struct {
	Data []DeleteGradesRequestParam `json:"data"`
}
type DeleteGradesRequestParam struct {
	GradeID teaching.GradeID `json:"gradeId"`
}
type DeleteGradesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteGradesRequest) Validate() errs.ValidationError {
	return nil
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
	CourseID teaching.CourseID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
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
	CourseID              teaching.CourseID `json:"courseId"`
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

type DeleteCoursesRequest struct {
	Data []DeleteCoursesRequestParam `json:"data"`
}
type DeleteCoursesRequestParam struct {
	CourseID teaching.CourseID `json:"courseId"`
}
type DeleteCoursesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteCoursesRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== CLASS ==============================

type GetClassesRequest struct {
	PaginationRequest
	IncludeDeactivated bool `json:"includeDeactivated,omitempty"`
}
type GetClassesResponse struct {
	Data    GetClassesResult `json:"data"`
	Message string           `json:"message,omitempty"`
}
type GetClassesResult struct {
	Results []teaching.Class `json:"results"`
	PaginationResponse
}

func (r GetClassesRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetClasses, MaxResultsPerPage_GetClasses); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetClassRequest struct {
	ClassID teaching.ClassID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetClassResponse struct {
	Data    teaching.Class `json:"data"`
	Message string         `json:"message,omitempty"`
}

func (r GetClassRequest) Validate() errs.ValidationError {
	return nil
}

type InsertClassesRequest struct {
	Data []InsertClassesRequestParam `json:"data"`
}
type InsertClassesRequestParam struct {
	TeacherID    teaching.TeacherID   `json:"teacherId"`
	StudentIDs   []teaching.StudentID `json:"studentIds"`
	CourseID     teaching.CourseID    `json:"courseId"`
	TransportFee int64                `json:"transportFee,omitempty"`
}
type InsertClassesResponse struct {
	Data    UpsertClassResult `json:"data"`
	Message string            `json:"message,omitempty"`
}

func (r InsertClassesRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateClassesRequest struct {
	Data []UpdateClassesRequestParam `json:"data"`
}
type UpdateClassesRequestParam struct {
	ClassID       teaching.ClassID     `json:"classId"`
	TeacherID     teaching.TeacherID   `json:"teacherId"`
	StudentIDs    []teaching.StudentID `json:"StudentIds"`
	TransportFee  int64                `json:"transportFee,omitempty"`
	IsDeactivated bool                 `json:"isDeactivated,omitempty"`
}
type UpdateClassesResponse struct {
	Data    UpsertClassResult `json:"data"`
	Message string            `json:"message,omitempty"`
}

func (r UpdateClassesRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertClassResult struct {
	Results []teaching.Class `json:"results"`
}

type DeleteClassesRequest struct {
	Data []DeleteClassesRequestParam `json:"data"`
}
type DeleteClassesRequestParam struct {
	ClassID teaching.ClassID `json:"classId"`
}
type DeleteClassesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteClassesRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== TEACHER_SPECIAL_FEE ==============================

type GetTeacherSpecialFeesRequest struct {
	PaginationRequest
}
type GetTeacherSpecialFeesResponse struct {
	Data    GetTeacherSpecialFeesResult `json:"data"`
	Message string                      `json:"message,omitempty"`
}
type GetTeacherSpecialFeesResult struct {
	Results []teaching.TeacherSpecialFee `json:"results"`
	PaginationResponse
}

func (r GetTeacherSpecialFeesRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetTeacherSpecialFees, MaxResultsPerPage_GetTeacherSpecialFees); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetTeacherSpecialFeeRequest struct {
	TeacherSpecialFeeID teaching.TeacherSpecialFeeID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetTeacherSpecialFeeResponse struct {
	Data    teaching.TeacherSpecialFee `json:"data"`
	Message string                     `json:"message,omitempty"`
}

func (r GetTeacherSpecialFeeRequest) Validate() errs.ValidationError {
	return nil
}

type InsertTeacherSpecialFeesRequest struct {
	Data []InsertTeacherSpecialFeesRequestParam `json:"data"`
}
type InsertTeacherSpecialFeesRequestParam struct {
	TeacherID teaching.TeacherID `json:"teacherId"`
	CourseID  teaching.CourseID  `json:"courseId"`
	Fee       int64              `json:"fee"`
}
type InsertTeacherSpecialFeesResponse struct {
	Data    UpsertTeacherSpecialFeeResult `json:"data"`
	Message string                        `json:"message,omitempty"`
}

func (r InsertTeacherSpecialFeesRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateTeacherSpecialFeesRequest struct {
	Data []UpdateTeacherSpecialFeesRequestParam `json:"data"`
}
type UpdateTeacherSpecialFeesRequestParam struct {
	TeacherSpecialFeeID teaching.TeacherSpecialFeeID `json:"teacherSpecialFeeId"`
	Fee                 int64                        `json:"fee"`
}
type UpdateTeacherSpecialFeesResponse struct {
	Data    UpsertTeacherSpecialFeeResult `json:"data"`
	Message string                        `json:"message,omitempty"`
}

func (r UpdateTeacherSpecialFeesRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertTeacherSpecialFeeResult struct {
	Results []teaching.TeacherSpecialFee `json:"results"`
}

type DeleteTeacherSpecialFeesRequest struct {
	Data []DeleteTeacherSpecialFeesRequestParam `json:"data"`
}
type DeleteTeacherSpecialFeesRequestParam struct {
	TeacherSpecialFeeID teaching.TeacherSpecialFeeID `json:"teacherSpecialFeeId"`
}
type DeleteTeacherSpecialFeesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteTeacherSpecialFeesRequest) Validate() errs.ValidationError {
	return nil
}
