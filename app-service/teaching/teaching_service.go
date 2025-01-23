package teaching

import (
	"context"
	"time"

	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
)

const (
	Default_OneCourseCycle                = 4
	Default_BalanceTopUp                  = Default_OneCourseCycle
	Default_PenaltyFeeValue               = 10000
	Default_CourseFeeSharingPercentage    = 0.5
	Default_TransportFeeSharingPercentage = 1.0
)

type UserTeachingInfo struct {
	TeacherID entity.TeacherID `json:"teacherId"`
	StudentID entity.StudentID `json:"studentId"`
	IsTeacher bool             `json:"isTeacher"`
	IsStudent bool             `json:"isStudent"`
}

type StudentEnrollmentInvoice struct {
	BalanceTopUp      int32      `json:"balanceTopUp"`
	BalanceBonus      int32      `json:"balanceBonus"`
	CourseFeeValue    int32      `json:"courseFeeValue"`
	TransportFeeValue int32      `json:"transportFeeValue"`
	PenaltyFeeValue   int32      `json:"penaltyFeeValue"`
	DiscountFeeValue  int32      `json:"discountFeeValue"`
	LastPaymentDate   *time.Time `json:"lastPaymentDate,omitempty"`
	DaysLate          int32      `json:"daysLate"`
}

type StudentIDToSLTs struct {
	StudentID             entity.StudentID                      `json:"studentId"`
	StudentLearningTokens []entity.StudentLearningToken_Minimal `json:"studentLearningTokens"`
}

type TeacherForPayment struct {
	entity.TeacherInfo_Minimal
	TotalAttendances int32 `json:"totalAttendances"`
}

// TeacherPaymentInvoiceItem is an "Attendance"/"TeacherPayment" which is reshaped (grouped-by in multi-level) for a rendering requirement in FE's TeacherPayment page.
//
// All information inside this nested struct is literally extracted from "Attendance"/"TeacherPayment".
type TeacherPaymentInvoiceItem struct {
	entity.ClassInfo_Minimal
	Students []tpii_Student `json:"students"`
}

type tpii_Student struct {
	entity.StudentInfo_Minimal
	StudentLearningTokens []tpii_StudentLearningToken `json:"studentLearningTokens"`
}

type tpii_StudentLearningToken struct {
	entity.StudentLearningToken_Minimal
	Attendances []tpii_AttendanceWithTeacherPayment `json:"attendances"`
}

type tpii_AttendanceWithTeacherPayment struct {
	entity.AttendanceInfo_Minimal
	// These 4 below fields are displayed in FE to simplify the calculation of PaidCourseFeeValue & PaidTransportFeeValue
	GrossCourseFeeValue           int32   `json:"grossCourseFeeValue"`
	GrossTransportFeeValue        int32   `json:"grossTransportFeeValue"`
	CourseFeeSharingPercentage    float64 `json:"courseFeeSharingPercentage"`
	TransportFeeSharingPercentage float64 `json:"transportFeeSharingPercentage"`

	// we allow ",omitempty", as these fields are only populated when this struct is created from "TeacherPayment" instead of "Attendance"
	// please refer to struct "teacherPaymentInvoiceItemRaw" for more information.
	TeacherPaymentID      entity.TeacherPaymentID `json:"teacherPaymentId,omitempty"`
	PaidCourseFeeValue    int32                   `json:"paidCourseFeeValue,omitempty"`
	PaidTransportFeeValue int32                   `json:"paidTransportFeeValue,omitempty"`
	AddedAt               time.Time               `json:"addedAt,omitempty"`
}

type TeachingService interface {
	GetUserTeachingInfo(ctx context.Context, id identity.UserID) (UserTeachingInfo, error)
	IsUserInvolvedInClass(ctx context.Context, userId identity.UserID, classId entity.ClassID) (bool, error)
	IsUserInvolvedInAttendance(ctx context.Context, userId identity.UserID, attendanceId entity.AttendanceID) (bool, error)

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
	EditClassesConfigs(ctx context.Context, specs []EditClassConfigSpec) error
	EditClassesCourses(ctx context.Context, specs []EditClassCourseSpec) error

	GetSLTsByClassID(ctx context.Context, classID entity.ClassID) ([]StudentIDToSLTs, error)
	GetAttendancesByClassID(ctx context.Context, spec GetAttendancesByClassIDSpec) (GetAttendancesByClassIDResult, error)
	// AddAttendancesBatch is the batch version of AddAttendance().
	AddAttendancesBatch(ctx context.Context, specs []AddAttendanceSpec) ([]entity.AttendanceID, error)
	// AddAttendance creates attendance(s) based on spec, duplicated for every students who enroll in the class.
	//
	// Depend on the `Class` setting ("autoOweAttendanceToken"), by default this will automatically create StudentLearningToken (SLT) with negative quota when any of the class' students have no SLT (due to no payment yet).
	AddAttendance(ctx context.Context, spec AddAttendanceSpec) ([]entity.AttendanceID, error)
	AssignAttendanceToken(ctx context.Context, spec AssignAttendanceTokenSpec) error
	EditAttendance(ctx context.Context, spec EditAttendanceSpec) ([]entity.AttendanceID, error)
	RemoveAttendance(ctx context.Context, attendanceID entity.AttendanceID) ([]entity.AttendanceID, error)

	GetTeachersForPayment(ctx context.Context, spec GetTeachersForPaymentSpec) (GetTeachersForPaymentResult, error)
	// GetTeacherPaymentInvoiceItems returns list of Attendance, sort ascendingly by date, grouped by StudentLearningToken, then by Student, and finally by Class.
	//
	// The result will be used for SubmitTeacherPayments spec.
	GetTeacherPaymentInvoiceItems(ctx context.Context, spec GetTeacherPaymentInvoiceItemsSpec) ([]TeacherPaymentInvoiceItem, error)
	GetExistingTeacherPaymentInvoiceItems(ctx context.Context, spec GetExistingTeacherPaymentInvoiceItemsSpec) ([]TeacherPaymentInvoiceItem, error)
	SubmitTeacherPayments(ctx context.Context, specs []SubmitTeacherPaymentsSpec) error
	ModifyTeacherPayments(ctx context.Context, specs []ModifyTeacherPaymentsSpec) (ModifyTeacherPaymentsResult, error)
	EditTeacherPayments(ctx context.Context, specs []EditTeacherPaymentsSpec) ([]entity.TeacherPaymentID, error)
	RemoveTeacherPayments(ctx context.Context, teacherPaymentIDs []entity.TeacherPaymentID) error
}

type SubmitStudentEnrollmentPaymentSpec struct {
	StudentEnrollmentID entity.StudentEnrollmentID
	PaymentDate         time.Time

	BalanceTopUp      int32
	BalanceBonus      int32
	CourseFeeValue    int32
	TransportFeeValue int32
	PenaltyFeeValue   int32
	DiscountFeeValue  int32
}
type EditStudentEnrollmentPaymentSpec struct {
	EnrollmentPaymentID entity.EnrollmentPaymentID
	PaymentDate         time.Time
	BalanceBonus        int32
	DiscountFeeValue    int32
}

type SearchClassSpec struct {
	TeacherID entity.TeacherID
	StudentID entity.StudentID
	CourseID  entity.CourseID
}

type EditClassCourseSpec struct {
	ClassID  entity.ClassID
	CourseID entity.CourseID
}

type EditClassConfigSpec struct {
	ClassID                entity.ClassID
	IsDeactivated          *bool
	AutoOweAttendanceToken *bool
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

type AssignAttendanceTokenSpec struct {
	AttendanceID           entity.AttendanceID
	StudentLearningTokenID entity.StudentLearningTokenID
}

func (s AssignAttendanceTokenSpec) GetInt64ID() int64 {
	return int64(s.AttendanceID)
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

type GetTeachersForPaymentSpec struct {
	IsPaid     bool
	Pagination util.PaginationSpec
	util.TimeSpec
}

type GetTeachersForPaymentResult struct {
	TeachersForPayment []TeacherForPayment
	PaginationResult   util.PaginationResult
}

type GetTeacherPaymentInvoiceItemsSpec struct {
	TeacherID entity.TeacherID
	util.TimeSpec
}

type GetExistingTeacherPaymentInvoiceItemsSpec struct {
	TeacherID entity.TeacherID
	TimeSpec  util.TimeSpec
}

type SubmitTeacherPaymentsSpec struct {
	AttendanceID          entity.AttendanceID
	PaidCourseFeeValue    int32
	PaidTransportFeeValue int32
}

type ModifyTeacherPaymentsSpec struct {
	TeacherPaymentID      entity.TeacherPaymentID
	PaidCourseFeeValue    int32
	PaidTransportFeeValue int32
	IsDeleted             bool
}

func (s ModifyTeacherPaymentsSpec) GetInt64ID() int64 {
	return int64(s.TeacherPaymentID)
}

type ModifyTeacherPaymentsResult struct {
	EditedTeacherPaymentIDs  []entity.TeacherPaymentID
	DeletedTeacherPaymentIDs []entity.TeacherPaymentID
}

type EditTeacherPaymentsSpec struct {
	TeacherPaymentID      entity.TeacherPaymentID
	PaidCourseFeeValue    int32
	PaidTransportFeeValue int32
}

func (s EditTeacherPaymentsSpec) GetInt64ID() int64 {
	return int64(s.TeacherPaymentID)
}
