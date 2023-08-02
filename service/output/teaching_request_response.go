package output

import (
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/errs"

	"time"
)

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
	StudentEnrollmentID entity.StudentEnrollmentID `json:"studentEnrollmentID"`
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

type EditEnrollmentPaymentBalanceRequest struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"enrollmentPaymentID"`
	PaymentDate         time.Time                  `json:"paymentDate"`
	BalanceTopUp        int32                      `json:"balanceTopUp"`
}
type EditEnrollmentPaymentBalanceResponse struct {
	Message string `json:"message,omitempty"`
}

func (r EditEnrollmentPaymentBalanceRequest) Validate() errs.ValidationError {
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
	EnrollmentPaymentID entity.EnrollmentPaymentID `json:"enrollmentPaymentID"`
}
type RemoveEnrollmentPaymentResponse struct {
	Message string `json:"message,omitempty"`
}

func (r RemoveEnrollmentPaymentRequest) Validate() errs.ValidationError {
	return nil
}