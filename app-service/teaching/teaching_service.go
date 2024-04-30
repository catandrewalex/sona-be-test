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

type TeacherSalaryInvoiceItem struct {
	entity.ClassInfo_Minimal
	Students []Attendance_Student `json:"students"`
}

type Attendance_Student struct {
	entity.StudentInfo_Minimal
	StudentLearningTokens []Attendance_SLT `json:"studentLearningTokens"`
}

type Attendance_SLT struct {
	entity.StudentLearningToken_Minimal
	Attendances []AttendanceForInvoiceItem `json:"attendances"`
}

type AttendanceForInvoiceItem struct {
	entity.AttendanceInfo_Minimal
	// These 4 below fields are displayed in FE to simplify the calculation of PaidCourseFeeValue & PaidTransportFeeValue
	GrossCourseFeeValue           int32   `json:"grossCourseFeeValue"`
	GrossTransportFeeValue        int32   `json:"grossTransportFeeValue"`
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

	GetAttendancesByClassID(ctx context.Context, spec GetAttendancesByClassIDSpec) (GetAttendancesByClassIDResult, error)
	// AddAttendance creates attendance(s) based on spec, duplicated for every students who enroll in the class.
	//
	// Enabling "autoCreateSLT" will automatically create StudentLearningToken (SLT) with negative quota when any of the class' students have no SLT (due to no payment yet).
	AddAttendance(ctx context.Context, spec AddAttendanceSpec, autoCreateSLT bool) ([]entity.AttendanceID, error)
	EditAttendance(ctx context.Context, spec EditAttendanceSpec) ([]entity.AttendanceID, error)
	RemoveAttendance(ctx context.Context, attendanceID entity.AttendanceID) ([]entity.AttendanceID, error)

	// GetTeacherSalaryInvoiceItems returns list of Attendance, sort ascendingly by date, grouped by StudentLearningToken, then by Student, and finally by Class.
	//
	// The result will be used for SubmitTeacherSalaries spec.
	GetTeacherSalaryInvoiceItems(ctx context.Context, spec GetTeacherSalaryInvoiceItemsSpec) ([]TeacherSalaryInvoiceItem, error)
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

type GetAttendancesByClassIDSpec struct {
	ClassID   entity.ClassID
	StudentID entity.StudentID
	util.PaginationSpec
	util.TimeSpec
}

type GetAttendancesByClassIDResult struct {
	Attendances      []entity.Attendance
	PaginationResult util.PaginationResult
}

type AddAttendanceSpec struct {
	ClassID               entity.ClassID
	TeacherID             entity.TeacherID
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
}

type EditAttendanceSpec struct {
	AttendanceID          entity.AttendanceID
	TeacherID             entity.TeacherID
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
}

func (s EditAttendanceSpec) GetInt64ID() int64 {
	return int64(s.AttendanceID)
}

type GetTeacherSalaryInvoiceItemsSpec struct {
	TeacherID entity.TeacherID
	util.TimeSpec
}

type SubmitTeacherSalariesSpec struct {
	AttendanceID          entity.AttendanceID
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
