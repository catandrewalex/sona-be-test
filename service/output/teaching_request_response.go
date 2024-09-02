package output

import (
	"fmt"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/errs"

	"time"
)

type GetUserTeachingInfoRequest struct{}
type GetUserTeachingInfoResponse struct {
	Data    teaching.UserTeachingInfo `json:"data"`
	Message string                    `json:"message,omitempty"`
}

func (r GetUserTeachingInfoRequest) Validate() errs.ValidationError {
	return nil
}

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
	BalanceBonus        int32                      `json:"balanceBonus,omitempty"`
	CourseFeeValue      int32                      `json:"courseFeeValue,omitempty"`
	TransportFeeValue   int32                      `json:"transportFeeValue,omitempty"`
	PenaltyFeeValue     int32                      `json:"penaltyFeeValue,omitempty"`
	DiscountFeeValue    int32                      `json:"discountFeeValue,omitempty"`
}
type SubmitEnrollmentPaymentResponse struct {
	Message string `json:"message,omitempty"`
}

func (r SubmitEnrollmentPaymentRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if r.BalanceTopUp < 0 {
		errorDetail["balanceTopUp"] = "balanceTopUp must be >= 0"
	}
	if r.CourseFeeValue < 0 {
		errorDetail["courseFeeValue"] = "courseFeeValue must be >= 0"
	}
	if r.TransportFeeValue < 0 {
		errorDetail["transportFeeValue"] = "transportFeeValue must be >= 0"
	}
	if r.PenaltyFeeValue < 0 {
		errorDetail["penaltyFeeValue"] = "penaltyFeeValue must be >= 0"
	}
	if r.DiscountFeeValue < 0 {
		errorDetail["discountFeeValue"] = "discountFeeValue must be >= 0"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type EditEnrollmentPaymentRequest struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"enrollmentPaymentId"`
	PaymentDate         time.Time                  `json:"paymentDate"`
	BalanceBonus        int32                      `json:"balanceBonus,omitempty"`
	DiscountFeeValue    int32                      `json:"discountFeeValue,omitempty"`
}
type EditEnrollmentPaymentResponse struct {
	Data    entity.EnrollmentPayment `json:"data"`
	Message string                   `json:"message,omitempty"`
}

func (r EditEnrollmentPaymentRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

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

// ============================== CLASS & ATTENDANCE ==============================

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

type GetAttendancesByClassIDRequest struct {
	ClassID   entity.ClassID   `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	StudentID entity.StudentID `json:"studentId,omitempty"`
	PaginationRequest
	YearMonthFilter
}
type GetAttendancesByClassIDResponse struct {
	Data    GetAttendancesByClassIDResult `json:"data"`
	Message string                        `json:"message,omitempty"`
}

type GetAttendancesByClassIDResult struct {
	Results []entity.Attendance `json:"results"`
	PaginationResponse
}

func (r GetAttendancesByClassIDRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if validationErr := r.YearMonthFilter.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type AddAttendanceRequest struct {
	ClassID               entity.ClassID   `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	TeacherID             entity.TeacherID `json:"teacherId"`
	Date                  time.Time        `json:"date"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	UsedStudentTokenQuota float64          `json:"usedStudentTokenQuota,omitempty"`
	Duration              int32            `json:"duration,omitempty"`
	Note                  string           `json:"note,omitempty"`
}
type AddAttendanceResponse struct {
	Data    UpsertAttendanceResult `json:"data"`
	Message string                 `json:"message,omitempty"`
}

func (r AddAttendanceRequest) Validate() errs.ValidationError {
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

type EditAttendanceRequest struct {
	AttendanceID          entity.AttendanceID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	TeacherID             entity.TeacherID    `json:"teacherId"`
	Date                  time.Time           `json:"date"` // in RFC3339 format: "2023-12-30T14:58:10+07:00"
	UsedStudentTokenQuota float64             `json:"usedStudentTokenQuota,omitempty"`
	Duration              int32               `json:"duration,omitempty"`
	Note                  string              `json:"note,omitempty"`
}
type EditAttendanceResponse struct {
	Data    UpsertAttendanceResult `json:"data"`
	Message string                 `json:"message,omitempty"`
}

func (r EditAttendanceRequest) Validate() errs.ValidationError {
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

type RemoveAttendanceRequest struct {
	AttendanceID entity.AttendanceID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type RemoveAttendanceResponse struct {
	Message string `json:"message,omitempty"`
}

func (r RemoveAttendanceRequest) Validate() errs.ValidationError {
	return nil
}

// ============================== TEACHER_PAYMENT ==============================

type GetUnpaidTeachersRequest struct {
	YearMonthFilter
	PaginationRequest
}
type GetUnpaidTeachersResponse struct {
	Data GetUnpaidTeachersResult `json:"data"`
}
type GetUnpaidTeachersResult struct {
	Results []teaching.TeacherForPayment `json:"results"`
	PaginationResponse
}

func (r GetUnpaidTeachersRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if validationErr := r.YearMonthFilter.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type GetPaidTeachersRequest struct {
	YearMonthFilter
	PaginationRequest
}
type GetPaidTeachersResponse struct {
	Data GetPaidTeachersResult `json:"data"`
}
type GetPaidTeachersResult struct {
	Results []teaching.TeacherForPayment `json:"results"`
	PaginationResponse
}

func (r GetPaidTeachersRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if validationErr := r.YearMonthFilter.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type GetTeacherPaymentInvoiceItemsRequest struct {
	TeacherID entity.TeacherID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	YearMonthFilter
}
type GetTeacherPaymentInvoiceItemsResponse struct {
	Data GetTeacherPaymentInvoiceItemsResult `json:"data"`
}
type GetTeacherPaymentInvoiceItemsResult struct {
	Results []teaching.TeacherPaymentInvoiceItem `json:"results"`
}

func (r GetTeacherPaymentInvoiceItemsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if validationErr := r.YearMonthFilter.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type GetTeacherPaymentsAsInvoiceItemsRequest struct {
	TeacherID entity.TeacherID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	YearMonthFilter
}
type GetTeacherPaymentsAsInvoiceItemsResponse struct {
	Data GetTeacherPaymentsAsInvoiceItemsResult `json:"data"`
}
type GetTeacherPaymentsAsInvoiceItemsResult struct {
	Results []teaching.TeacherPaymentInvoiceItem `json:"results"`
}

func (r GetTeacherPaymentsAsInvoiceItemsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	if validationErr := r.YearMonthFilter.Validate(); validationErr != nil {
		for key, value := range validationErr.GetErrorDetail() {
			errorDetail[key] = value
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type SubmitTeacherPaymentsRequest struct {
	Data []SubmitTeacherPaymentsRequestParam `json:"data"`
}
type SubmitTeacherPaymentsRequestParam struct {
	AttendanceID          entity.AttendanceID `json:"attendanceId"`
	PaidCourseFeeValue    int32               `json:"paidCourseFeeValue,omitempty"`
	PaidTransportFeeValue int32               `json:"paidTransportFeeValue,omitempty"`
}
type SubmitTeacherPaymentsResponse struct {
	Message string `json:"message,omitempty"`
}

func (r SubmitTeacherPaymentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	for i, datum := range r.Data {
		if datum.PaidCourseFeeValue < 0 {
			errorDetail[fmt.Sprintf("data.%d.paidCourseFeeValue", i)] = "paidCourseFeeValue must be >= 0"
		}
		if datum.PaidTransportFeeValue < 0 {
			errorDetail[fmt.Sprintf("data.%d.paidTransportFeeValue", i)] = "paidTransportFeeValue must be >= 0"
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type ModifyTeacherPaymentsRequest struct {
	Data []ModifyTeacherPaymentsRequestParam `json:"data"`
}
type ModifyTeacherPaymentsRequestParam struct {
	TeacherPaymentID      entity.TeacherPaymentID `json:"teacherPaymentId"`
	PaidCourseFeeValue    int32                   `json:"paidCourseFeeValue,omitempty"`
	PaidTransportFeeValue int32                   `json:"paidTransportFeeValue,omitempty"`
	IsDeleted             bool                    `json:"isDeleted,omitempty"`
}
type ModifyTeacherPaymentsResponse struct {
	Data    ModifyTeacherPaymentsResult `json:"data"`
	Message string                      `json:"message,omitempty"`
}

type ModifyTeacherPaymentsResult struct {
	Results []entity.TeacherPayment `json:"results"`
}

func (r ModifyTeacherPaymentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	for i, datum := range r.Data {
		if datum.PaidCourseFeeValue < 0 {
			errorDetail[fmt.Sprintf("data.%d.paidCourseFeeValue", i)] = "paidCourseFeeValue must be >= 0"
		}
		if datum.PaidTransportFeeValue < 0 {
			errorDetail[fmt.Sprintf("data.%d.paidTransportFeeValue", i)] = "paidTransportFeeValue must be >= 0"
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type EditTeacherPaymentsRequest struct {
	Data []EditTeacherPaymentsRequestParam `json:"data"`
}
type EditTeacherPaymentsRequestParam struct {
	TeacherPaymentID      entity.TeacherPaymentID `json:"teacherPaymentId"`
	PaidCourseFeeValue    int32                   `json:"paidCourseFeeValue,omitempty"`
	PaidTransportFeeValue int32                   `json:"paidTransportFeeValue,omitempty"`
}
type EditTeacherPaymentsResponse struct {
	Data    EditTeacherPaymentsResult `json:"data"`
	Message string                    `json:"message,omitempty"`
}

type EditTeacherPaymentsResult struct {
	Results []entity.TeacherPayment `json:"results"`
}

func (r EditTeacherPaymentsRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)

	for i, datum := range r.Data {
		if datum.PaidCourseFeeValue < 0 {
			errorDetail[fmt.Sprintf("data.%d.paidCourseFeeValue", i)] = "paidCourseFeeValue must be >= 0"
		}
		if datum.PaidTransportFeeValue < 0 {
			errorDetail[fmt.Sprintf("data.%d.paidTransportFeeValue", i)] = "paidTransportFeeValue must be >= 0"
		}
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type RemoveTeacherPaymentsRequest struct {
	Data []RemoveTeacherPaymentsRequestParam `json:"data"`
}
type RemoveTeacherPaymentsRequestParam struct {
	TeacherPaymentID entity.TeacherPaymentID `json:"teacherPaymentId"`
}
type RemoveTeacherPaymentsResponse struct {
	Message string `json:"message,omitempty"`
}

func (r RemoveTeacherPaymentsRequest) Validate() errs.ValidationError {
	return nil
}
