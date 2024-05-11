package teaching

import (
	"context"
	"sort"
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
	BalanceTopUp      int32      `json:"balanceTopUp"`
	PenaltyFeeValue   int32      `json:"penaltyFeeValue"`
	CourseFeeValue    int32      `json:"courseFeeValue"`
	TransportFeeValue int32      `json:"transportFeeValue"`
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

	GetSLTsByClassID(ctx context.Context, classID entity.ClassID) ([]StudentIDToSLTs, error)
	GetAttendancesByClassID(ctx context.Context, spec GetAttendancesByClassIDSpec) (GetAttendancesByClassIDResult, error)
	// AddAttendance creates attendance(s) based on spec, duplicated for every students who enroll in the class.
	//
	// Enabling "autoCreateSLT" will automatically create StudentLearningToken (SLT) with negative quota when any of the class' students have no SLT (due to no payment yet).
	AddAttendance(ctx context.Context, spec AddAttendanceSpec, autoCreateSLT bool) ([]entity.AttendanceID, error)
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
	TeacherID          entity.TeacherID
	AttendanceTimeSpec util.TimeSpec
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

// ============================== HELPER STRUCTS & CONSTRUCTORS ==============================

func calculateGrossCourseAndTransportFeeValue(attendance entity.Attendance) (int32, int32) {
	grossCourseFeeValue := int32(float64(attendance.StudentLearningToken.CourseFeeValue) * attendance.UsedStudentTokenQuota / float64(Default_OneCourseCycle))
	grossTransportFeeValue := int32(float64(attendance.StudentLearningToken.TransportFeeValue) * attendance.UsedStudentTokenQuota / float64(Default_OneCourseCycle))
	return grossCourseFeeValue, grossTransportFeeValue
}

// teacherPaymentInvoiceItemRaw is an intermediate struct to facilitate the creation of TeacherPaymentInvoiceItem,
// which currently can come from 2 different but very similar structs: "Attendance" and "TeacherPayment".
type teacherPaymentInvoiceItemRaw struct {
	Attendance             entity.Attendance
	GrossCourseFeeValue    int32
	GrossTransportFeeValue int32

	// these below 4 fields may be zero value, if this struct is constructed from "Attendance"
	TeacherPaymentID      entity.TeacherPaymentID
	PaidCourseFeeValue    int32
	PaidTransportFeeValue int32
	AddedAt               time.Time
}

func CreateTeacherPaymentInvoiceItemsRawFromAttendances(attendances []entity.Attendance) []teacherPaymentInvoiceItemRaw {
	result := make([]teacherPaymentInvoiceItemRaw, 0, len(attendances))
	for _, attendance := range attendances {
		grossCourseFeeValue, grossTransportFeeValue := calculateGrossCourseAndTransportFeeValue(attendance)
		result = append(result, teacherPaymentInvoiceItemRaw{
			Attendance:             attendance,
			GrossCourseFeeValue:    grossCourseFeeValue,
			GrossTransportFeeValue: grossTransportFeeValue,

			TeacherPaymentID:      entity.TeacherPaymentID_None,
			PaidCourseFeeValue:    0,
			PaidTransportFeeValue: 0,
			AddedAt:               time.Time{},
		})
	}
	return result
}

func CreateTeacherPaymentInvoiceItemRawFromTeacherPayment(teacherPayments []entity.TeacherPayment) []teacherPaymentInvoiceItemRaw {
	result := make([]teacherPaymentInvoiceItemRaw, 0, len(teacherPayments))
	for _, teacherPayment := range teacherPayments {
		result = append(result, teacherPaymentInvoiceItemRaw{
			Attendance:             teacherPayment.Attendance,
			GrossCourseFeeValue:    teacherPayment.GrossCourseFeeValue,
			GrossTransportFeeValue: teacherPayment.GrossTransportFeeValue,

			TeacherPaymentID:      teacherPayment.TeacherPaymentID,
			PaidCourseFeeValue:    teacherPayment.PaidCourseFeeValue,
			PaidTransportFeeValue: teacherPayment.PaidTransportFeeValue,
			AddedAt:               teacherPayment.AddedAt,
		})
	}
	return result
}

func CreateTeacherPaymentInvoiceItems(tpiisRaw []teacherPaymentInvoiceItemRaw) []TeacherPaymentInvoiceItem {
	// group the Attendances by StudentLearningToken
	sltIdToAttendances := make(map[entity.StudentLearningTokenID][]tpii_AttendanceWithTeacherPayment)
	sltIdToSLT := make(map[entity.StudentLearningTokenID]entity.StudentLearningToken_Minimal)
	sltIdToStudent := make(map[entity.StudentLearningTokenID]entity.StudentInfo_Minimal)
	studentIdToStudent := make(map[entity.StudentID]entity.StudentInfo_Minimal)
	studentIdToClass := make(map[entity.StudentID]entity.ClassInfo_Minimal)
	for _, teacherPaymentInvoiceItemRaw := range tpiisRaw {
		attendance := teacherPaymentInvoiceItemRaw.Attendance

		var courseFeeSharingPercentage float64 = Default_CourseFeeSharingPercentage
		var transportFeeSharingPercentage float64 = Default_TransportFeeSharingPercentage
		if teacherPaymentInvoiceItemRaw.PaidCourseFeeValue > 0 && teacherPaymentInvoiceItemRaw.GrossCourseFeeValue > 0 {
			courseFeeSharingPercentage = float64(teacherPaymentInvoiceItemRaw.PaidCourseFeeValue) / float64(teacherPaymentInvoiceItemRaw.GrossCourseFeeValue)
		}
		if teacherPaymentInvoiceItemRaw.PaidTransportFeeValue > 0 && teacherPaymentInvoiceItemRaw.GrossTransportFeeValue > 0 {
			courseFeeSharingPercentage = float64(teacherPaymentInvoiceItemRaw.PaidTransportFeeValue) / float64(teacherPaymentInvoiceItemRaw.GrossTransportFeeValue)
		}

		tpiiAttendance := tpii_AttendanceWithTeacherPayment{
			AttendanceInfo_Minimal: entity.AttendanceInfo_Minimal{
				AttendanceID:          attendance.AttendanceID,
				TeacherInfo:           attendance.TeacherInfo,
				Date:                  attendance.Date,
				UsedStudentTokenQuota: attendance.UsedStudentTokenQuota,
				Duration:              attendance.Duration,
				Note:                  attendance.Note,
				IsPaid:                attendance.IsPaid,
			},
			GrossCourseFeeValue:           teacherPaymentInvoiceItemRaw.GrossCourseFeeValue,
			GrossTransportFeeValue:        teacherPaymentInvoiceItemRaw.GrossTransportFeeValue,
			CourseFeeSharingPercentage:    courseFeeSharingPercentage,
			TransportFeeSharingPercentage: transportFeeSharingPercentage,

			TeacherPaymentID:      teacherPaymentInvoiceItemRaw.TeacherPaymentID,
			PaidCourseFeeValue:    teacherPaymentInvoiceItemRaw.PaidCourseFeeValue,
			PaidTransportFeeValue: teacherPaymentInvoiceItemRaw.PaidTransportFeeValue,
			AddedAt:               teacherPaymentInvoiceItemRaw.AddedAt,
		}

		sltId := attendance.StudentLearningToken.StudentLearningTokenID
		prevValues, ok := sltIdToAttendances[sltId]
		if ok {
			sltIdToAttendances[sltId] = append(prevValues, tpiiAttendance)
		} else {
			sltIdToAttendances[sltId] = []tpii_AttendanceWithTeacherPayment{tpiiAttendance}
		}
		sltIdToSLT[sltId] = attendance.StudentLearningToken
		sltIdToStudent[sltId] = attendance.StudentInfo
		studentIdToStudent[attendance.StudentInfo.StudentID] = attendance.StudentInfo
		studentIdToClass[attendance.StudentInfo.StudentID] = attendance.ClassInfo
	}

	// then group the StudentLearningTokens by Student
	studentIdToSLTs := make(map[entity.StudentID][]tpii_StudentLearningToken)
	for sltId, attendances := range sltIdToAttendances {
		studentLearningToken := sltIdToSLT[sltId]
		tpiiSLT := tpii_StudentLearningToken{
			StudentLearningToken_Minimal: studentLearningToken,
			Attendances:                  attendances,
		}

		student := sltIdToStudent[sltId]
		prevValues, ok := studentIdToSLTs[student.StudentID]
		if ok {
			studentIdToSLTs[student.StudentID] = append(prevValues, tpiiSLT)
		} else {
			studentIdToSLTs[student.StudentID] = []tpii_StudentLearningToken{tpiiSLT}
		}
	}

	// then group the Students by Class, with SLTs per Student are sorted by ID (the SLT ID is automatically generated, so sort by ID == sort by date)
	classIdToStudents := make(map[entity.ClassID][]tpii_Student)
	classIdToClass := make(map[entity.ClassID]entity.ClassInfo_Minimal)
	for studentId, slts := range studentIdToSLTs {
		sort.SliceStable(slts, func(i, j int) bool {
			return slts[i].StudentLearningTokenID < slts[j].StudentLearningTokenID
		})

		student := studentIdToStudent[studentId]
		class := studentIdToClass[studentId]
		tpiiStudent := tpii_Student{
			StudentInfo_Minimal:   student,
			StudentLearningTokens: slts,
		}

		prevValues, ok := classIdToStudents[class.ClassID]
		if ok {
			classIdToStudents[class.ClassID] = append(prevValues, tpiiStudent)
		} else {
			classIdToStudents[class.ClassID] = []tpii_Student{tpiiStudent}
		}
		classIdToClass[class.ClassID] = class
	}

	// construct the TeacherPaymentInvoiceItem, with Students per Class are sorted by Student name
	teacherPaymentInvoiceItems := make([]TeacherPaymentInvoiceItem, 0, 1)
	for classId, students := range classIdToStudents {
		sort.SliceStable(students, func(i, j int) bool {
			return students[i].StudentInfo_Minimal.String() < students[j].StudentInfo_Minimal.String()
		})

		class := classIdToClass[classId]
		teacherPaymentInvoiceItems = append(teacherPaymentInvoiceItems, TeacherPaymentInvoiceItem{
			ClassInfo_Minimal: class,
			Students:          students,
		})
	}
	// finally, sort the TeacherPaymentInvoiceItem by Class name
	sort.SliceStable(teacherPaymentInvoiceItems, func(i, j int) bool {
		return teacherPaymentInvoiceItems[i].ClassInfo_Minimal.String() < teacherPaymentInvoiceItems[j].ClassInfo_Minimal.String()
	})

	return teacherPaymentInvoiceItems
}
