package teaching

import (
	"context"
	"time"

	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/util"
)

const (
	Default_OneCourseCycle                = 4
	Default_BalanceTopUp                  = Default_OneCourseCycle
	Default_PenaltyFeeValue               = 10000
	Default_CourseFeeSharingPercentage    = 0.5
	Default_TransportFeeSharingPercentage = 1.0
)

type StudentEnrollmentInvoice struct {
	BalanceTopUp      int32 `json:"balanceTopUp"`
	PenaltyFeeValue   int32 `json:"penaltyFeeValue"`
	CourseFeeValue    int32 `json:"courseFeeValue"`
	TransportFeeValue int32 `json:"transportFeeValue"`
}

type TeacherSalaryInvoice struct {
	Presence entity.Presence `json:"presence"`
	// These 4 below fields are displayed in FE to simplify the calculation of PaidCourseFeeValue & PaidTransportFeeValue
	CourseFeeFullValue            int32   `json:"courseFeeFullValue"`
	TransportFeeFullValue         int32   `json:"transportFeeFullValue"`
	CourseFeeSharingPercentage    float64 `json:"courseFeeSharingPercentage"`
	TransportFeeSharingPercentage float64 `json:"transportFeeSharingPercentage"`
}

type TeachingService interface {
	SearchEnrollmentPayment(ctx context.Context, timeFilter util.TimeSpec) ([]entity.EnrollmentPayment, error)
	// GetEnrollmentPaymentInvoice returns values for used by SubmitEnrollmentPayment.
	// This includes calculating teacherSpecialFee, and penaltyFee.
	GetEnrollmentPaymentInvoice(ctx context.Context, studentEnrollmentID entity.StudentEnrollmentID) (StudentEnrollmentInvoice, error)
	// SubmitEnrollmentPayment adds new enrollmentPayment, then upsert StudentLearningToken (insert new, or update quota).
	// The SLT update will sum up spec.BalanceTopUp with all negative quota, set them to 0, and set the summed quota for the earliest available SLT.
	SubmitEnrollmentPayment(ctx context.Context, spec SubmitStudentEnrollmentPaymentSpec) error
	EditEnrollmentPayment(ctx context.Context, spec EditStudentEnrollmentPaymentSpec) (entity.EnrollmentPaymentID, error)
	RemoveEnrollmentPayment(ctx context.Context, enrollmentPaymentID entity.EnrollmentPaymentID) error

	SearchClass(ctx context.Context, spec SearchClassSpec) ([]entity.Class, error)

	GetPresencesByClassID(ctx context.Context, spec GetPresencesByClassIDSpec) (GetPresencesByClassIDResult, error)
	// AddPresence creates presence(s) based on spec, duplicated for every students who enroll in the class.
	//
	// Enabling "autoCreateSLT" will automatically create StudentLearningToken (SLT) with negative quota when any of the class' students have no SLT (due to no payment yet).
	AddPresence(ctx context.Context, spec AddPresenceSpec, autoCreateSLT bool) ([]entity.PresenceID, error)
	EditPresence(ctx context.Context, spec EditPresenceSpec) ([]entity.PresenceID, error)
	RemovePresence(ctx context.Context, presenceID entity.PresenceID) ([]entity.PresenceID, error)

	GetTeacherSalaryInvoices(ctx context.Context, spec GetTeacherSalaryInvoicesSpec) ([]TeacherSalaryInvoice, error)
	SubmitTeacherSalaries(ctx context.Context, specs []SubmitTeacherSalariesSpec) error
	EditTeacherSalaries(ctx context.Context, specs []EditTeacherSalariesSpec) ([]entity.TeacherSalaryID, error)
	RemoveTeacherSalaries(ctx context.Context, teacherSalaryIDs []entity.TeacherSalaryID) error
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
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
}

type EditPresenceSpec struct {
	PresenceID            entity.PresenceID
	TeacherID             entity.TeacherID
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
}

func (s EditPresenceSpec) GetInt64ID() int64 {
	return int64(s.PresenceID)
}

type GetTeacherSalaryInvoicesSpec struct {
	TeacherID entity.TeacherID
	ClassID   entity.ClassID
	util.TimeSpec
}

type SubmitTeacherSalariesSpec struct {
	PresenceID            entity.PresenceID
	PaidCourseFeeValue    int32
	PaidTransportFeeValue int32
}

type EditTeacherSalariesSpec struct {
	TeacherSalaryID       entity.TeacherSalaryID
	PaidCourseFeeValue    int32
	PaidTransportFeeValue int32
}

func (s EditTeacherSalariesSpec) GetInt64ID() int64 {
	return int64(s.TeacherSalaryID)
}
