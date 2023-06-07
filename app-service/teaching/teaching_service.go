package teaching

import (
	"context"

	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
)

type Teacher struct {
	TeacherID TeacherID     `json:"teacherId"`
	User      identity.User `json:"user"`
}

type Student struct {
	StudentID StudentID     `json:"studentId"`
	User      identity.User `json:"user"`
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

type Class struct {
	ClassID            ClassID                       `json:"classId"`
	TeacherInfo        *Class_TeacherInfo            `json:"teacherInfo,omitempty"` // class without teacher is a valid class
	StudentEnrollments []Class_StudentEnrollmentInfo `json:"studentEnrollments"`
	Course             Course                        `json:"course"`
	TransportFee       int64                         `json:"transportFee"`
	IsDeactivated      bool                          `json:"isDeactivated"`
}

type Class_TeacherInfo struct {
	TeacherID  TeacherID           `json:"teacherId"`
	Username   string              `json:"username"`
	UserDetail identity.UserDetail `json:"userDetail"`
}

type Class_StudentEnrollmentInfo struct {
	StudentEnrollmentID StudentEnrollmentID    `json:"studentEnrollmentId"`
	StudentInfo         Enrollment_StudentInfo `json:"studentInfo"`
}

type StudentEnrollment struct {
	StudentEnrollmentID StudentEnrollmentID    `json:"studentEnrollmentId"`
	StudentInfo         Enrollment_StudentInfo `json:"studentInfo"`
	ClassID             ClassID                `json:"classId"`
}

type Enrollment_StudentInfo struct {
	StudentID  StudentID           `json:"studentId"`
	Username   string              `json:"username"`
	UserDetail identity.UserDetail `json:"userDetail"`
}

type TeacherID int64
type StudentID int64
type InstrumentID int64
type GradeID int64
type CourseID int64
type ClassID int64
type StudentEnrollmentID int64
type StudentLearningTokenID int64
type TeacherSpecialFeeID int64
type PresenceID int64

const (
	TeacherID_None TeacherID = iota
)
const (
	StudentID_None StudentID = iota
)
const (
	InstrumentID_None InstrumentID = iota
)
const (
	GradeID_None GradeID = iota
)
const (
	CourseID_None CourseID = iota
)
const (
	ClassID_None ClassID = iota
)
const (
	StudentEnrollmentID_None StudentEnrollmentID = iota
)
const (
	StudentLearningTokenID_None StudentLearningTokenID = iota
)
const (
	TeacherSpecialFeeID_None TeacherSpecialFeeID = iota
)
const (
	PresenceID_None PresenceID = iota
)

type TeachingService interface {
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

// ============================== COURSE ==============================

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
	ClassID              ClassID
	TeacherID            TeacherID
	AddedStudentIDs      []StudentID
	DeletedEnrollmentIDs []StudentEnrollmentID
	TransportFee         int64
	IsDeactivated        bool
}
