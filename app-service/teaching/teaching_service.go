package teaching

import (
	"context"
	"time"

	"sonamusica-backend/app-service/entity"
)

const (
	Default_OneCourseCycle  = 4
	Default_BalanceTopUp    = 4
	Default_PenaltyFeeValue = 20000
)

type StudentEnrollmentInvoice struct {
	BalanceTopUp      int32 `json:"balanceTopUp"`
	PenaltyFeeValue   int32 `json:"penaltyFeeValue"`
	CourseFeeValue    int32 `json:"courseFeeValue"`
	TransportFeeValue int32 `json:"transportFeeValue"`
}

type TeachingService interface {
	CalculateStudentEnrollmentInvoice(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID) (StudentEnrollmentInvoice, error)
	SubmitStudentEnrollmentPayment(ctx context.Context, spec SubmitStudentEnrollmentPaymentSpec) error
	EditStudentEnrollmentPaymentBalance(ctx context.Context, spec EditStudentEnrollmentPaymentBalanceSpec) error
	RemoveStudentEnrollmentPayment(ctx context.Context, enrollmentPaymentID entity.EnrollmentPaymentID) error

	AddPresence(ctx context.Context, spec AddPresenceSpec) error
}

type SubmitStudentEnrollmentPaymentSpec struct {
	StudentEnrollmentID entity.StudentEnrollmentID
	PaymentDate         time.Time

	BalanceTopUp      int32
	PenaltyFeeValue   int32
	CourseFeeValue    int32
	TransportFeeValue int32
}
type EditStudentEnrollmentPaymentBalanceSpec struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID
	PaymentDate         time.Time
	BalanceTopUp        int32
}

type AddPresenceSpec struct {
	entity.InsertPresenceSpec
}
