package output

import (
	"fmt"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/errs"
	"time"
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

	MaxPage_GetStudentEnrollments           = Default_MaxPage
	MaxResultsPerPage_GetStudentEnrollments = Default_MaxResultsPerPage

	MaxPage_GetTeacherSpecialFees           = Default_MaxPage
	MaxResultsPerPage_GetTeacherSpecialFees = Default_MaxResultsPerPage

	MaxPage_GetEnrollmentPayments           = Default_MaxPage
	MaxResultsPerPage_GetEnrollmentPayments = Default_MaxResultsPerPage

	MaxPage_GetStudentLearningTokens           = Default_MaxPage
	MaxResultsPerPage_GetStudentLearningTokens = Default_MaxResultsPerPage

	MaxPage_GetPresences           = Default_MaxPage
	MaxResultsPerPage_GetPresences = Default_MaxResultsPerPage
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
	Results []entity.Instrument `json:"results"`
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
	InstrumentID entity.InstrumentID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetInstrumentResponse struct {
	Data    entity.Instrument `json:"data"`
	Message string            `json:"message,omitempty"`
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
	InstrumentID entity.InstrumentID `json:"instrumentId"`
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
	Results []entity.Instrument `json:"results"`
}

type DeleteInstrumentsRequest struct {
	Data []DeleteInstrumentsRequestParam `json:"data"`
}
type DeleteInstrumentsRequestParam struct {
	InstrumentID entity.InstrumentID `json:"instrumentId"`
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
	Results []entity.Grade `json:"results"`
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
	GradeID entity.GradeID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetGradeResponse struct {
	Data    entity.Grade `json:"data"`
	Message string       `json:"message,omitempty"`
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
	GradeID entity.GradeID `json:"gradeId"`
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
	Results []entity.Grade `json:"results"`
}

type DeleteGradesRequest struct {
	Data []DeleteGradesRequestParam `json:"data"`
}
type DeleteGradesRequestParam struct {
	GradeID entity.GradeID `json:"gradeId"`
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
	Results []entity.Course `json:"results"`
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
	CourseID entity.CourseID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetCourseResponse struct {
	Data    entity.Course `json:"data"`
	Message string        `json:"message,omitempty"`
}

func (r GetCourseRequest) Validate() errs.ValidationError {
	return nil
}

type InsertCoursesRequest struct {
	Data []InsertCoursesRequestParam `json:"data"`
}
type InsertCoursesRequestParam struct {
	InstrumentID          entity.InstrumentID `json:"instrumentId"`
	GradeID               entity.GradeID      `json:"gradeId"`
	DefaultFee            int32               `json:"defaultFee"`
	DefaultDurationMinute int32               `json:"defaultDurationMinute"`
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
	CourseID              entity.CourseID `json:"courseId"`
	DefaultFee            int32           `json:"defaultFee"`
	DefaultDurationMinute int32           `json:"defaultDurationMinute"`
}
type UpdateCoursesResponse struct {
	Data    UpsertCourseResult `json:"data"`
	Message string             `json:"message,omitempty"`
}

func (r UpdateCoursesRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertCourseResult struct {
	Results []entity.Course `json:"results"`
}

type DeleteCoursesRequest struct {
	Data []DeleteCoursesRequestParam `json:"data"`
}
type DeleteCoursesRequestParam struct {
	CourseID entity.CourseID `json:"courseId"`
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
	Results []entity.Class `json:"results"`
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
	ClassID entity.ClassID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetClassResponse struct {
	Data    entity.Class `json:"data"`
	Message string       `json:"message,omitempty"`
}

func (r GetClassRequest) Validate() errs.ValidationError {
	return nil
}

type InsertClassesRequest struct {
	Data []InsertClassesRequestParam `json:"data"`
}
type InsertClassesRequestParam struct {
	TeacherID    entity.TeacherID   `json:"teacherId"`
	StudentIDs   []entity.StudentID `json:"studentIds"`
	CourseID     entity.CourseID    `json:"courseId"`
	TransportFee int32              `json:"transportFee,omitempty"`
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
	ClassID       entity.ClassID     `json:"classId"`
	TeacherID     entity.TeacherID   `json:"teacherId"`
	StudentIDs    []entity.StudentID `json:"StudentIds"`
	TransportFee  int32              `json:"transportFee,omitempty"`
	IsDeactivated bool               `json:"isDeactivated,omitempty"`
}
type UpdateClassesResponse struct {
	Data    UpsertClassResult `json:"data"`
	Message string            `json:"message,omitempty"`
}

func (r UpdateClassesRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertClassResult struct {
	Results []entity.Class `json:"results"`
}

type DeleteClassesRequest struct {
	Data []DeleteClassesRequestParam `json:"data"`
}
type DeleteClassesRequestParam struct {
	ClassID entity.ClassID `json:"classId"`
}
type DeleteClassesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteClassesRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== STUDENT_ENROLLMENT ==============================

type GetStudentEnrollmentsRequest struct {
	PaginationRequest
}
type GetStudentEnrollmentsResponse struct {
	Data    GetStudentEnrollmentsResult `json:"data"`
	Message string                      `json:"message,omitempty"`
}
type GetStudentEnrollmentsResult struct {
	Results []entity.StudentEnrollment `json:"results"`
	PaginationResponse
}

func (r GetStudentEnrollmentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetStudentEnrollments, MaxResultsPerPage_GetStudentEnrollments); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
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
	Results []entity.TeacherSpecialFee `json:"results"`
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
	TeacherSpecialFeeID entity.TeacherSpecialFeeID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetTeacherSpecialFeeResponse struct {
	Data    entity.TeacherSpecialFee `json:"data"`
	Message string                   `json:"message,omitempty"`
}

func (r GetTeacherSpecialFeeRequest) Validate() errs.ValidationError {
	return nil
}

type InsertTeacherSpecialFeesRequest struct {
	Data []InsertTeacherSpecialFeesRequestParam `json:"data"`
}
type InsertTeacherSpecialFeesRequestParam struct {
	TeacherID entity.TeacherID `json:"teacherId"`
	CourseID  entity.CourseID  `json:"courseId"`
	Fee       int32            `json:"fee"`
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
	TeacherSpecialFeeID entity.TeacherSpecialFeeID `json:"teacherSpecialFeeId"`
	Fee                 int32                      `json:"fee"`
}
type UpdateTeacherSpecialFeesResponse struct {
	Data    UpsertTeacherSpecialFeeResult `json:"data"`
	Message string                        `json:"message,omitempty"`
}

func (r UpdateTeacherSpecialFeesRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertTeacherSpecialFeeResult struct {
	Results []entity.TeacherSpecialFee `json:"results"`
}

type DeleteTeacherSpecialFeesRequest struct {
	Data []DeleteTeacherSpecialFeesRequestParam `json:"data"`
}
type DeleteTeacherSpecialFeesRequestParam struct {
	TeacherSpecialFeeID entity.TeacherSpecialFeeID `json:"teacherSpecialFeeId"`
}
type DeleteTeacherSpecialFeesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteTeacherSpecialFeesRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== ENROLLMENT_PAYMENT ==============================

type GetEnrollmentPaymentsRequest struct {
	PaginationRequest
}
type GetEnrollmentPaymentsResponse struct {
	Data    GetEnrollmentPaymentsResult `json:"data"`
	Message string                      `json:"message,omitempty"`
}
type GetEnrollmentPaymentsResult struct {
	Results []entity.EnrollmentPayment `json:"results"`
	PaginationResponse
}

func (r GetEnrollmentPaymentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetEnrollmentPayments, MaxResultsPerPage_GetEnrollmentPayments); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetEnrollmentPaymentRequest struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetEnrollmentPaymentResponse struct {
	Data    entity.EnrollmentPayment `json:"data"`
	Message string                   `json:"message,omitempty"`
}

func (r GetEnrollmentPaymentRequest) Validate() errs.ValidationError {
	return nil
}

type InsertEnrollmentPaymentsRequest struct {
	Data []InsertEnrollmentPaymentsRequestParam `json:"data"`
}
type InsertEnrollmentPaymentsRequestParam struct {
	StudentEnrollmentID entity.StudentEnrollmentID `json:"studentEnrollmentID"`
	PaymentDate         time.Time                  `json:"paymentDate"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	BalanceTopUp        int32                      `json:"balanceTopUp"`
	Value               int32                      `json:"value,omitempty"`
	ValuePenalty        int32                      `json:"valuePenalty,omitempty"`
}
type InsertEnrollmentPaymentsResponse struct {
	Data    UpsertEnrollmentPaymentResult `json:"data"`
	Message string                        `json:"message,omitempty"`
}

func (r InsertEnrollmentPaymentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	for i, datum := range r.Data {
		if datum.BalanceTopUp < 0 {
			errorDetail[fmt.Sprintf("data.%d.balanceTopUp", i)] = fmt.Sprintf("balanceTopUp must be >= 0")
		}
		if datum.Value < 0 {
			errorDetail[fmt.Sprintf("data.%d.value", i)] = fmt.Sprintf("value must be >= 0")
		}
		if datum.ValuePenalty < 0 {
			errorDetail[fmt.Sprintf("data.%d.valuePenalty", i)] = fmt.Sprintf("valuePenalty must be >= 0")
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type UpdateEnrollmentPaymentsRequest struct {
	Data []UpdateEnrollmentPaymentsRequestParam `json:"data"`
}
type UpdateEnrollmentPaymentsRequestParam struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"enrollmentPaymentID"`
	PaymentDate         time.Time                  `json:"paymentDate"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	BalanceTopUp        int32                      `json:"balanceTopUp,omitempty"`
	Value               int32                      `json:"value,omitempty"`
	ValuePenalty        int32                      `json:"valuePenalty,omitempty"`
}
type UpdateEnrollmentPaymentsResponse struct {
	Data    UpsertEnrollmentPaymentResult `json:"data"`
	Message string                        `json:"message,omitempty"`
}

func (r UpdateEnrollmentPaymentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	for i, datum := range r.Data {
		if datum.BalanceTopUp < 0 {
			errorDetail[fmt.Sprintf("data.%d.balanceTopUp", i)] = fmt.Sprintf("balanceTopUp must be >= 0")
		}
		if datum.Value < 0 {
			errorDetail[fmt.Sprintf("data.%d.value", i)] = fmt.Sprintf("value must be >= 0")
		}
		if datum.ValuePenalty < 0 {
			errorDetail[fmt.Sprintf("data.%d.valuePenalty", i)] = fmt.Sprintf("valuePenalty must be >= 0")
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type UpsertEnrollmentPaymentResult struct {
	Results []entity.EnrollmentPayment `json:"results"`
}

type DeleteEnrollmentPaymentsRequest struct {
	Data []DeleteEnrollmentPaymentsRequestParam `json:"data"`
}
type DeleteEnrollmentPaymentsRequestParam struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"enrollmentPaymentID"`
}
type DeleteEnrollmentPaymentsResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteEnrollmentPaymentsRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== STUDENT_LEARNING_TOKEN ==============================

type GetStudentLearningTokensRequest struct {
	PaginationRequest
}
type GetStudentLearningTokensResponse struct {
	Data    GetStudentLearningTokensResult `json:"data"`
	Message string                         `json:"message,omitempty"`
}
type GetStudentLearningTokensResult struct {
	Results []entity.StudentLearningToken `json:"results"`
	PaginationResponse
}

func (r GetStudentLearningTokensRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetStudentLearningTokens, MaxResultsPerPage_GetStudentLearningTokens); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetStudentLearningTokenRequest struct {
	StudentLearningTokenID entity.StudentLearningTokenID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetStudentLearningTokenResponse struct {
	Data    entity.StudentLearningToken `json:"data"`
	Message string                      `json:"message,omitempty"`
}

func (r GetStudentLearningTokenRequest) Validate() errs.ValidationError {
	return nil
}

type InsertStudentLearningTokensRequest struct {
	Data []InsertStudentLearningTokensRequestParam `json:"data"`
}
type InsertStudentLearningTokensRequestParam struct {
	StudentEnrollmentID entity.StudentEnrollmentID `json:"studentEnrollmentID"`
	Quota               int32                      `json:"quota"`
	CourseFeeValue      int32                      `json:"courseFeeValue,omitempty"`
	TransportFeeValue   int32                      `json:"transportFeeValue,omitempty"`
}
type InsertStudentLearningTokensResponse struct {
	Data    UpsertStudentLearningTokenResult `json:"data"`
	Message string                           `json:"message,omitempty"`
}

func (r InsertStudentLearningTokensRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateStudentLearningTokensRequest struct {
	Data []UpdateStudentLearningTokensRequestParam `json:"data"`
}
type UpdateStudentLearningTokensRequestParam struct {
	StudentLearningTokenID entity.StudentLearningTokenID `json:"studentLearningTokenID"`
	Quota                  int32                         `json:"quota"`
	CourseFeeValue         int32                         `json:"courseFeeValue,omitempty"`
	TransportFeeValue      int32                         `json:"transportFeeValue,omitempty"`
}
type UpdateStudentLearningTokensResponse struct {
	Data    UpsertStudentLearningTokenResult `json:"data"`
	Message string                           `json:"message,omitempty"`
}

func (r UpdateStudentLearningTokensRequest) Validate() errs.ValidationError {
	return nil
}

type UpsertStudentLearningTokenResult struct {
	Results []entity.StudentLearningToken `json:"results"`
}

type DeleteStudentLearningTokensRequest struct {
	Data []DeleteStudentLearningTokensRequestParam `json:"data"`
}
type DeleteStudentLearningTokensRequestParam struct {
	StudentLearningTokenID entity.StudentLearningTokenID `json:"studentLearningTokenID"`
}
type DeleteStudentLearningTokensResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeleteStudentLearningTokensRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== PRESENCE ==============================

type GetPresencesRequest struct {
	PaginationRequest
	TimeFilter
}
type GetPresencesResponse struct {
	Data    GetPresencesResult `json:"data"`
	Message string             `json:"message,omitempty"`
}
type GetPresencesResult struct {
	Results []entity.Presence `json:"results"`
	PaginationResponse
}

func (r GetPresencesRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetPresences, MaxResultsPerPage_GetPresences); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if validationErr := r.TimeFilter.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetPresenceRequest struct {
	PresenceID entity.PresenceID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetPresenceResponse struct {
	Data    entity.Presence `json:"data"`
	Message string          `json:"message,omitempty"`
}

func (r GetPresenceRequest) Validate() errs.ValidationError {
	return nil
}

type InsertPresencesRequest struct {
	Data []InsertPresencesRequestParam `json:"data"`
}
type InsertPresencesRequestParam struct {
	ClassID                entity.ClassID                `json:"classID"`
	TeacherID              entity.TeacherID              `json:"teacherID"`
	StudentID              entity.StudentID              `json:"studentID"`
	StudentLearningTokenID entity.StudentLearningTokenID `json:"studentLearningTokenID"`
	Date                   time.Time                     `json:"date"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	UsedStudentTokenQuota  float64                       `json:"usedStudentTokenQuota,omitempty"`
	Duration               int32                         `json:"duration,omitempty"`
}
type InsertPresencesResponse struct {
	Data    UpsertPresenceResult `json:"data"`
	Message string               `json:"message,omitempty"`
}

func (r InsertPresencesRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	for i, datum := range r.Data {
		if datum.UsedStudentTokenQuota < 0 {
			errorDetail[fmt.Sprintf("data.%d.usedStudentTokenQuota", i)] = fmt.Sprintf("usedStudentTokenQuota must be >= 0")
		}
		if datum.Duration < 0 {
			errorDetail[fmt.Sprintf("data.%d.duration", i)] = fmt.Sprintf("duration must be >= 0")
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type UpdatePresencesRequest struct {
	Data []UpdatePresencesRequestParam `json:"data"`
}
type UpdatePresencesRequestParam struct {
	PresenceID             entity.PresenceID             `json:"presenceID"`
	ClassID                entity.ClassID                `json:"classID"`
	TeacherID              entity.TeacherID              `json:"teacherID"`
	StudentID              entity.StudentID              `json:"studentID"`
	StudentLearningTokenID entity.StudentLearningTokenID `json:"studentLearningTokenID"`
	Date                   time.Time                     `json:"date"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	UsedStudentTokenQuota  float64                       `json:"usedStudentTokenQuota"`
	Duration               int32                         `json:"duration"`
}
type UpdatePresencesResponse struct {
	Data    UpsertPresenceResult `json:"data"`
	Message string               `json:"message,omitempty"`
}

func (r UpdatePresencesRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	for i, datum := range r.Data {
		if datum.UsedStudentTokenQuota < 0 {
			errorDetail[fmt.Sprintf("data.%d.usedStudentTokenQuota", i)] = fmt.Sprintf("usedStudentTokenQuota must be >= 0")
		}
		if datum.Duration < 0 {
			errorDetail[fmt.Sprintf("data.%d.duration", i)] = fmt.Sprintf("duration must be >= 0")
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type UpsertPresenceResult struct {
	Results []entity.Presence `json:"results"`
}

type DeletePresencesRequest struct {
	Data []DeletePresencesRequestParam `json:"data"`
}
type DeletePresencesRequestParam struct {
	PresenceID entity.PresenceID `json:"presenceID"`
}
type DeletePresencesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r DeletePresencesRequest) Validate() errs.ValidationError {
	return nil
}
