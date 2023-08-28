package teaching

import (
	"context"
	"time"

	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/util"
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
	SearchEnrollmentPayment(ctx context.Context, timeFilter util.TimeSpec) ([]entity.EnrollmentPayment, error)
	CalculateStudentEnrollmentInvoice(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID) (StudentEnrollmentInvoice, error)
	SubmitEnrollmentPayment(ctx context.Context, spec SubmitStudentEnrollmentPaymentSpec) error
	EditEnrollmentPayment(ctx context.Context, spec EditStudentEnrollmentPaymentSpec) (entity.EnrollmentPaymentID, error)
	RemoveEnrollmentPayment(ctx context.Context, enrollmentPaymentID entity.EnrollmentPaymentID) error

	SearchClass(ctx context.Context, spec SearchClassSpec) ([]entity.Class, error)

	GetPresencesByClassID(ctx context.Context, spec GetPresencesByClassIDSpec) (GetPresencesByClassIDResult, error)
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
type EditStudentEnrollmentPaymentSpec struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID
	PaymentDate         time.Time
	BalanceTopUp        int32
}

type SearchClassSpec struct {
	TeacherID entity.TeacherID
	StudentID entity.StudentID
	CourseID  entity.CourseID
}

type GetPresencesByClassIDSpec struct {
	ClassID   entity.ClassID
	StudentID entity.StudentID
	util.PaginationSpec
	util.TimeSpec
}

type GetPresencesByClassIDResult struct {
	Presences        []entity.Presence
	PaginationResult util.PaginationResult
}

type AddPresenceSpec struct {
	ClassID               entity.ClassID
	TeacherID             entity.TeacherID
	StudentID             entity.StudentID
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
}
