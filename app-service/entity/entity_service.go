package entity

import (
	"context"
	"time"

	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
)

type Teacher struct {
	TeacherID TeacherID     `json:"teacherId"`
	User      identity.User `json:"user"`
}

// TeacherInfo_Minimal is a subset of struct Teacher that must have the same schema.
type TeacherInfo_Minimal struct {
	TeacherID        TeacherID                 `json:"teacherId"`
	UserInfo_Minimal identity.UserInfo_Minimal `json:"user"`
}

type Student struct {
	StudentID StudentID     `json:"studentId"`
	User      identity.User `json:"user"`
}

// StudentInfo_Minimal is a subset of struct Student that must have the same schema.
type StudentInfo_Minimal struct {
	StudentID        StudentID                 `json:"studentId"`
	UserInfo_Minimal identity.UserInfo_Minimal `json:"user"`
}

type Instrument struct {
	InstrumentID InstrumentID `json:"instrumentId"`
	Name         string       `json:"name"`
}

type Grade struct {
	GradeID GradeID `json:"gradeId"`
	Name    string  `json:"name"`
}

type Course struct {
	CourseID              CourseID   `json:"courseId"`
	Instrument            Instrument `json:"instrument"`
	Grade                 Grade      `json:"grade"`
	DefaultFee            int64      `json:"defaultFee"`
	DefaultDurationMinute int32      `json:"defaultDurationMinute"`
}

// CourseInfo_Minimal is a subset of struct Course that must have the same schema.
type CourseInfo_Minimal struct {
	CourseID   CourseID   `json:"courseId"`
	Instrument Instrument `json:"instrument"`
	Grade      Grade      `json:"grade"`
}

type Class struct {
	ClassID              ClassID               `json:"classId"`
	TeacherInfo_Minimal  *TeacherInfo_Minimal  `json:"teacher,omitempty"` // class without teacher is a valid class
	StudentInfos_Minimal []StudentInfo_Minimal `json:"students"`
	Course               Course                `json:"course"`
	TransportFee         int64                 `json:"transportFee"`
	IsDeactivated        bool                  `json:"isDeactivated"`
}

type ClassInfo_Minimal struct {
	ClassID       ClassID            `json:"classId"`
	CourseInfo    CourseInfo_Minimal `json:"course"`
	TransportFee  int64              `json:"transportFee"`
	IsDeactivated bool               `json:"isDeactivated"`
}

type StudentEnrollment struct {
	StudentEnrollmentID StudentEnrollmentID `json:"studentEnrollmentID"`
	StudentInfo         StudentInfo_Minimal `json:"student"`
	ClassInfo           ClassInfo_Minimal   `json:"class"`
}

type TeacherSpecialFee struct {
	TeacherSpecialFeeID TeacherSpecialFeeID `json:"teacherSpecialFeeId"`
	TeacherInfo         TeacherInfo_Minimal `json:"teacher"`
	CourseInfo          CourseInfo_Minimal  `json:"course"`
	Fee                 int64               `json:"fee"`
}

type EnrollmentPayment struct {
	EnrollmentPaymentID   EnrollmentPaymentID `json:"enrollmentPaymentId"`
	StudentEnrollmentInfo StudentEnrollment   `json:"studentEnrollment"`
	PaymentDate           time.Time           `json:"paymentDate"`
	BalanceTopUp          int32               `json:"balanceTopUp"`
	Value                 int32               `json:"value"`
	ValuePenalty          int32               `json:"valuePenalty"`
}

type StudentLearningToken struct {
	StudentLearningTokenID StudentLearningTokenID `json:"studentLearningTokenID"`
	Quota                  int32                  `json:"quota"`
	QuotaBonus             int32                  `json:"quotaBonus"`
	CourseFeeValue         int32                  `json:"courseFeeValue"`
	TransportFeeValue      int32                  `json:"transportFeeValue"`
	LastUpdatedAt          time.Time              `json:"lastUpdatedAt"`
	StudentEnrollmentInfo  StudentEnrollment      `json:"studentEnrollment"`
}

type TeacherID int64
type StudentID int64
type InstrumentID int64
type GradeID int64
type CourseID int64
type ClassID int64
type StudentEnrollmentID int64

type TeacherSpecialFeeID int64
type EnrollmentPaymentID int64
type StudentLearningTokenID int64
type PresenceID int64

const TeacherID_None TeacherID = iota
const StudentID_None StudentID = iota
const InstrumentID_None InstrumentID = iota
const GradeID_None GradeID = iota
const CourseID_None CourseID = iota
const ClassID_None ClassID = iota
const StudentEnrollmentID_None StudentEnrollmentID = iota

const TeacherSpecialFeeID_None TeacherSpecialFeeID = iota
const EnrollmentPaymentID_None EnrollmentPaymentID = iota
const StudentLearningTokenID_None StudentLearningTokenID = iota
const PresenceID_None PresenceID = iota

type EntityService interface {
	GetTeachers(ctx context.Context, pagination util.PaginationSpec) (GetTeachersResult, error)
	GetTeacherById(ctx context.Context, id TeacherID) (Teacher, error)
	GetTeachersByIds(ctx context.Context, ids []TeacherID) ([]Teacher, error)
	InsertTeachers(ctx context.Context, userIDs []identity.UserID) ([]TeacherID, error)
	InsertTeachersWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]TeacherID, error)
	DeleteTeachers(ctx context.Context, ids []TeacherID) error

	GetStudents(ctx context.Context, pagination util.PaginationSpec) (GetStudentsResult, error)
	GetStudentById(ctx context.Context, id StudentID) (Student, error)
	GetStudentsByIds(ctx context.Context, ids []StudentID) ([]Student, error)
	InsertStudents(ctx context.Context, userIDs []identity.UserID) ([]StudentID, error)
	InsertStudentsWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]StudentID, error)
	DeleteStudents(ctx context.Context, ids []StudentID) error

	GetInstruments(ctx context.Context, pagination util.PaginationSpec) (GetInstrumentsResult, error)
	GetInstrumentById(ctx context.Context, id InstrumentID) (Instrument, error)
	GetInstrumentsByIds(ctx context.Context, ids []InstrumentID) ([]Instrument, error)
	InsertInstruments(ctx context.Context, specs []InsertInstrumentSpec) ([]InstrumentID, error)
	UpdateInstruments(ctx context.Context, specs []UpdateInstrumentSpec) ([]InstrumentID, error)
	DeleteInstruments(ctx context.Context, ids []InstrumentID) error

	GetGrades(ctx context.Context, pagination util.PaginationSpec) (GetGradesResult, error)
	GetGradeById(ctx context.Context, id GradeID) (Grade, error)
	GetGradesByIds(ctx context.Context, ids []GradeID) ([]Grade, error)
	InsertGrades(ctx context.Context, specs []InsertGradeSpec) ([]GradeID, error)
	UpdateGrades(ctx context.Context, specs []UpdateGradeSpec) ([]GradeID, error)
	DeleteGrades(ctx context.Context, ids []GradeID) error

	GetCourses(ctx context.Context, pagination util.PaginationSpec) (GetCoursesResult, error)
	GetCourseById(ctx context.Context, id CourseID) (Course, error)
	GetCoursesByIds(ctx context.Context, ids []CourseID) ([]Course, error)
	InsertCourses(ctx context.Context, specs []InsertCourseSpec) ([]CourseID, error)
	UpdateCourses(ctx context.Context, specs []UpdateCourseSpec) ([]CourseID, error)
	DeleteCourses(ctx context.Context, ids []CourseID) error

	GetClasses(ctx context.Context, pagination util.PaginationSpec, includeDeactivated bool) (GetClassesResult, error)
	GetClassById(ctx context.Context, id ClassID) (Class, error)
	GetClassesByIds(ctx context.Context, ids []ClassID) ([]Class, error)
	InsertClasses(ctx context.Context, specs []InsertClassSpec) ([]ClassID, error)
	UpdateClasses(ctx context.Context, specs []UpdateClassSpec) ([]ClassID, error)
	DeleteClasses(ctx context.Context, ids []ClassID) error

	GetStudentEnrollments(ctx context.Context, pagination util.PaginationSpec) (GetStudentEnrollmentsResult, error)

	GetTeacherSpecialFees(ctx context.Context, pagination util.PaginationSpec) (GetTeacherSpecialFeesResult, error)
	GetTeacherSpecialFeeById(ctx context.Context, id TeacherSpecialFeeID) (TeacherSpecialFee, error)
	GetTeacherSpecialFeesByIds(ctx context.Context, ids []TeacherSpecialFeeID) ([]TeacherSpecialFee, error)
	InsertTeacherSpecialFees(ctx context.Context, specs []InsertTeacherSpecialFeeSpec) ([]TeacherSpecialFeeID, error)
	UpdateTeacherSpecialFees(ctx context.Context, specs []UpdateTeacherSpecialFeeSpec) ([]TeacherSpecialFeeID, error)
	DeleteTeacherSpecialFees(ctx context.Context, ids []TeacherSpecialFeeID) error

	GetEnrollmentPayments(ctx context.Context, pagination util.PaginationSpec) (GetEnrollmentPaymentsResult, error)
	GetEnrollmentPaymentById(ctx context.Context, id EnrollmentPaymentID) (EnrollmentPayment, error)
	GetEnrollmentPaymentsByIds(ctx context.Context, ids []EnrollmentPaymentID) ([]EnrollmentPayment, error)
	InsertEnrollmentPayments(ctx context.Context, specs []InsertEnrollmentPaymentSpec) ([]EnrollmentPaymentID, error)
	UpdateEnrollmentPayments(ctx context.Context, specs []UpdateEnrollmentPaymentSpec) ([]EnrollmentPaymentID, error)
	DeleteEnrollmentPayments(ctx context.Context, ids []EnrollmentPaymentID) error

	GetStudentLearningTokens(ctx context.Context, pagination util.PaginationSpec) (GetStudentLearningTokensResult, error)
	GetStudentLearningTokenById(ctx context.Context, id StudentLearningTokenID) (StudentLearningToken, error)
	GetStudentLearningTokensByIds(ctx context.Context, ids []StudentLearningTokenID) ([]StudentLearningToken, error)
	InsertStudentLearningTokens(ctx context.Context, specs []InsertStudentLearningTokenSpec) ([]StudentLearningTokenID, error)
	UpdateStudentLearningTokens(ctx context.Context, specs []UpdateStudentLearningTokenSpec) ([]StudentLearningTokenID, error)
	DeleteStudentLearningTokens(ctx context.Context, ids []StudentLearningTokenID) error
}

// ============================== STUDENT & TEACHER ==============================

type GetTeachersResult struct {
	Teachers         []Teacher
	PaginationResult util.PaginationResult
}

type GetStudentsResult struct {
	Students         []Student
	PaginationResult util.PaginationResult
}

// ============================== INSTRUMENT ==============================

type GetInstrumentsResult struct {
	Instruments      []Instrument
	PaginationResult util.PaginationResult
}

type InsertInstrumentSpec struct {
	Name string
}

type UpdateInstrumentSpec struct {
	InstrumentID InstrumentID
	Name         string
}

func (s UpdateInstrumentSpec) GetInt64ID() int64 {
	return int64(s.InstrumentID)
}

// ============================== GRADE ==============================

type GetGradesResult struct {
	Grades           []Grade
	PaginationResult util.PaginationResult
}

type InsertGradeSpec struct {
	Name string
}

type UpdateGradeSpec struct {
	GradeID GradeID
	Name    string
}

func (s UpdateGradeSpec) GetInt64ID() int64 {
	return int64(s.GradeID)
}

// ============================== COURSE ==============================

type GetCoursesResult struct {
	Courses          []Course
	PaginationResult util.PaginationResult
}

type InsertCourseSpec struct {
	InstrumentID          InstrumentID
	GradeID               GradeID
	DefaultFee            int64
	DefaultDurationMinute int32
}

type UpdateCourseSpec struct {
	CourseID              CourseID
	DefaultFee            int64
	DefaultDurationMinute int32
}

func (s UpdateCourseSpec) GetInt64ID() int64 {
	return int64(s.CourseID)
}

// ============================== CLASS ==============================

type GetClassesResult struct {
	Classes          []Class
	PaginationResult util.PaginationResult
}

type InsertClassSpec struct {
	TeacherID    TeacherID
	StudentIDs   []StudentID
	CourseID     CourseID
	TransportFee int64
}

type UpdateClassSpec struct {
	ClassID       ClassID
	TeacherID     TeacherID
	StudentIDs    []StudentID
	TransportFee  int64
	IsDeactivated bool
}

func (s UpdateClassSpec) GetInt64ID() int64 {
	return int64(s.ClassID)
}

// ============================== STUDENT_ENROLLMENT ==============================

type GetStudentEnrollmentsResult struct {
	StudentEnrollments []StudentEnrollment
	PaginationResult   util.PaginationResult
}

// ============================== TEACHER_SPECIAL_FEE ==============================

type GetTeacherSpecialFeesResult struct {
	TeacherSpecialFees []TeacherSpecialFee
	PaginationResult   util.PaginationResult
}

type InsertTeacherSpecialFeeSpec struct {
	TeacherID TeacherID
	CourseID  CourseID
	Fee       int64
}

type UpdateTeacherSpecialFeeSpec struct {
	TeacherSpecialFeeID TeacherSpecialFeeID
	Fee                 int64
}

func (s UpdateTeacherSpecialFeeSpec) GetInt64ID() int64 {
	return int64(s.TeacherSpecialFeeID)
}

// ============================== ENROLLMENT_PAYMENT ==============================

type GetEnrollmentPaymentsResult struct {
	EnrollmentPayments []EnrollmentPayment
	PaginationResult   util.PaginationResult
}

type InsertEnrollmentPaymentSpec struct {
	StudentEnrollmentID StudentEnrollmentID
	PaymentDate         time.Time
	BalanceTopUp        int32
	Value               int32
	ValuePenalty        int32
}

type UpdateEnrollmentPaymentSpec struct {
	EnrollmentPaymentID EnrollmentPaymentID
	PaymentDate         time.Time
	BalanceTopUp        int32
	Value               int32
	ValuePenalty        int32
}

func (s UpdateEnrollmentPaymentSpec) GetInt64ID() int64 {
	return int64(s.EnrollmentPaymentID)
}

// ============================== ENROLLMENT_PAYMENT ==============================

type GetStudentLearningTokensResult struct {
	StudentLearningTokens []StudentLearningToken
	PaginationResult      util.PaginationResult
}

type InsertStudentLearningTokenSpec struct {
	StudentEnrollmentID StudentEnrollmentID
	Quota               int32
	QuotaBonus          int32
	CourseFeeValue      int32
	TransportFeeValue   int32
}

type UpdateStudentLearningTokenSpec struct {
	StudentLearningTokenID StudentLearningTokenID
	Quota                  int32
	QuotaBonus             int32
	CourseFeeValue         int32
	TransportFeeValue      int32
}

func (s UpdateStudentLearningTokenSpec) GetInt64ID() int64 {
	return int64(s.StudentLearningTokenID)
}
