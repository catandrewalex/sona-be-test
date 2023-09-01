package output

import (
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/errs"

	"time"
)

// ============================== ENROLLMENT_PAYMENT ==============================

type SearchEnrollmentPaymentsRequest struct {
	TimeFilter
}
type SearchEnrollmentPaymentsResponse struct {
	Data    SearchEnrollmentPaymentsResult `json:"data"`
	Message string                         `json:"message,omitempty"`
}

type SearchEnrollmentPaymentsResult struct {
	Results []entity.EnrollmentPayment `json:"results"`
}

func (r SearchEnrollmentPaymentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

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

type GetEnrollmentPaymentInvoiceRequest struct {
	StudentEnrollmentID entity.StudentEnrollmentID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetEnrollmentPaymentInvoiceResponse struct {
	Data    teaching.StudentEnrollmentInvoice `json:"data"`
	Message string                            `json:"message,omitempty"`
}

func (r GetEnrollmentPaymentInvoiceRequest) Validate() errs.ValidationError {
	return nil
}

type SubmitEnrollmentPaymentRequest struct {
	StudentEnrollmentID entity.StudentEnrollmentID `json:"studentEnrollmentId"`
	PaymentDate         time.Time                  `json:"paymentDate"`
	BalanceTopUp        int32                      `json:"balanceTopUp"`
	PenaltyFeeValue     int32                      `json:"penaltyFeeValue,omitempty"`
	CourseFeeValue      int32                      `json:"courseFeeValue,omitempty"`
	TransportFeeValue   int32                      `json:"transportFeeValue,omitempty"`
}
type SubmitEnrollmentPaymentResponse struct {
	Message string `json:"message,omitempty"`
}

func (r SubmitEnrollmentPaymentRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if r.BalanceTopUp < 0 {
		errorDetail["balanceTopUp"] = "balanceTopUp must be >= 0"
	}
	if r.PenaltyFeeValue < 0 {
		errorDetail["penaltyFeeValue"] = "penaltyFeeValue must be >= 0"
	}
	if r.CourseFeeValue < 0 {
		errorDetail["courseFeeValue"] = "courseFeeValue must be >= 0"
	}
	if r.TransportFeeValue < 0 {
		errorDetail["transportFeeValue"] = "transportFeeValue must be >= 0"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type EditEnrollmentPaymentRequest struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"enrollmentPaymentId"`
	PaymentDate         time.Time                  `json:"paymentDate"`
	BalanceTopUp        int32                      `json:"balanceTopUp"`
}
type EditEnrollmentPaymentResponse struct {
	Data    entity.EnrollmentPayment `json:"data"`
	Message string                   `json:"message,omitempty"`
}

func (r EditEnrollmentPaymentRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if r.BalanceTopUp < 0 {
		errorDetail["balanceTopUp"] = "balanceTopUp must be >= 0"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type RemoveEnrollmentPaymentRequest struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"enrollmentPaymentId"`
}
type RemoveEnrollmentPaymentResponse struct {
	Message string `json:"message,omitempty"`
}

func (r RemoveEnrollmentPaymentRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== CLASS & PRESENCE ==============================

type SearchClassRequest struct {
	TeacherID entity.TeacherID `json:"teacherId,omitempty"`
	StudentID entity.StudentID `json:"studentId,omitempty"`
	CourseID  entity.CourseID  `json:"courseId,omitempty"`
}
type SearchClassResponse struct {
	Data    SearchClassResult `json:"data"`
	Message string            `json:"message,omitempty"`
}

type SearchClassResult struct {
	Results []entity.Class `json:"results"`
}

func (r SearchClassRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if r.TeacherID == entity.TeacherID_None && r.StudentID == entity.StudentID_None && r.CourseID == entity.CourseID_None {
		errorDetail["searchFilter"] = "either teacherId, studentId, courseId filter must be filled"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetPresencesByClassIDRequest struct {
	ClassID   entity.ClassID   `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	StudentID entity.StudentID `json:"studentId,omitempty"`
	PaginationRequest
	TimeFilter
}
type GetPresencesByClassIDResponse struct {
	Data    GetPresencesByClassIDResult `json:"data"`
	Message string                      `json:"message,omitempty"`
}

type GetPresencesByClassIDResult struct {
	Results []entity.Presence `json:"results"`
	PaginationResponse
}

func (r GetPresencesByClassIDRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

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

type AddPresenceRequest struct {
	ClassID               entity.ClassID   `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	TeacherID             entity.TeacherID `json:"teacherId"`
	Date                  time.Time        `json:"date"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	UsedStudentTokenQuota float64          `json:"usedStudentTokenQuota,omitempty"`
	Duration              int32            `json:"duration,omitempty"`
	Note                  string           `json:"note,omitempty"`
}
type AddPresenceResponse struct {
	Data    UpsertPresenceResult `json:"data"`
	Message string               `json:"message,omitempty"`
}

func (r AddPresenceRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if r.UsedStudentTokenQuota < 0 {
		errorDetail["usedStudentTokenQuota"] = "usedStudentTokenQuota must be >= 0"
	}
	if r.Duration < 0 {
		errorDetail["duration"] = "duration must be >= 0"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type EditPresenceRequest struct {
	PresenceID            entity.PresenceID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	TeacherID             entity.TeacherID  `json:"teacherId"`
	Date                  time.Time         `json:"date"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	UsedStudentTokenQuota float64           `json:"usedStudentTokenQuota,omitempty"`
	Duration              int32             `json:"duration,omitempty"`
	Note                  string            `json:"note,omitempty"`
}
type EditPresenceResponse struct {
	Data    UpsertPresenceResult `json:"data"`
	Message string               `json:"message,omitempty"`
}

func (r EditPresenceRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if r.UsedStudentTokenQuota < 0 {
		errorDetail["usedStudentTokenQuota"] = "usedStudentTokenQuota must be >= 0"
	}
	if r.Duration < 0 {
		errorDetail["duration"] = "duration must be >= 0"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type RemovePresenceRequest struct {
	PresenceID entity.PresenceID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type RemovePresenceResponse struct {
	Message string `json:"message,omitempty"`
}

func (r RemovePresenceRequest) Validate() errs.ValidationError {
	return nil
}
