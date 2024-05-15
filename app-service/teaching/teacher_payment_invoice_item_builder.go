package teaching

import (
	"sonamusica-backend/app-service/entity"
	"sort"
	"time"
)

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

func (t teacherPaymentInvoiceItemRaw) toTPIIAttendanceWithTeacherPayment() tpii_AttendanceWithTeacherPayment {
	attendance := t.Attendance

	var courseFeeSharingPercentage float64 = Default_CourseFeeSharingPercentage
	var transportFeeSharingPercentage float64 = Default_TransportFeeSharingPercentage
	if t.PaidCourseFeeValue > 0 && t.GrossCourseFeeValue > 0 {
		courseFeeSharingPercentage = float64(t.PaidCourseFeeValue) / float64(t.GrossCourseFeeValue)
	}
	if t.PaidTransportFeeValue > 0 && t.GrossTransportFeeValue > 0 {
		courseFeeSharingPercentage = float64(t.PaidTransportFeeValue) / float64(t.GrossTransportFeeValue)
	}

	return tpii_AttendanceWithTeacherPayment{
		AttendanceInfo_Minimal: entity.AttendanceInfo_Minimal{
			AttendanceID:          attendance.AttendanceID,
			TeacherInfo:           attendance.TeacherInfo,
			Date:                  attendance.Date,
			UsedStudentTokenQuota: attendance.UsedStudentTokenQuota,
			Duration:              attendance.Duration,
			Note:                  attendance.Note,
			IsPaid:                attendance.IsPaid,
		},
		GrossCourseFeeValue:           t.GrossCourseFeeValue,
		GrossTransportFeeValue:        t.GrossTransportFeeValue,
		CourseFeeSharingPercentage:    courseFeeSharingPercentage,
		TransportFeeSharingPercentage: transportFeeSharingPercentage,

		TeacherPaymentID:      t.TeacherPaymentID,
		PaidCourseFeeValue:    t.PaidCourseFeeValue,
		PaidTransportFeeValue: t.PaidTransportFeeValue,
		AddedAt:               t.AddedAt,
	}
}

// teacherPaymentInvoiceItemBuilder is a constructor to simplify "TeacherPaymentInvoiceItem" creation by isolating the complex struct-reshaping logics.
type teacherPaymentInvoiceItemBuilder struct {
	rawItems []teacherPaymentInvoiceItemRaw

	// these fields are populated and used by Build(), based on "rawItems"
	attendanceIdToTPIIRaw map[entity.AttendanceID]*teacherPaymentInvoiceItemRaw
	classIdToClass        map[entity.ClassID]entity.ClassInfo_Minimal
	studentIdToStudent    map[entity.StudentID]entity.StudentInfo_Minimal
	sltIdToSLT            map[entity.StudentLearningTokenID]entity.StudentLearningToken_Minimal
}

func NewTeacherPaymentInvoiceItemBuilder() *teacherPaymentInvoiceItemBuilder {
	return &teacherPaymentInvoiceItemBuilder{
		rawItems: make([]teacherPaymentInvoiceItemRaw, 0),

		attendanceIdToTPIIRaw: make(map[entity.AttendanceID]*teacherPaymentInvoiceItemRaw, 0),
		classIdToClass:        make(map[entity.ClassID]entity.ClassInfo_Minimal, 0),
		studentIdToStudent:    make(map[entity.StudentID]entity.StudentInfo_Minimal, 0),
		sltIdToSLT:            make(map[entity.StudentLearningTokenID]entity.StudentLearningToken_Minimal, 0),
	}
}

func (b *teacherPaymentInvoiceItemBuilder) AddAttendances(attendances []entity.Attendance) {
	for _, attendance := range attendances {
		grossCourseFeeValue, grossTransportFeeValue := calculateGrossCourseAndTransportFeeValue(attendance)
		b.rawItems = append(b.rawItems, teacherPaymentInvoiceItemRaw{
			Attendance:             attendance,
			GrossCourseFeeValue:    grossCourseFeeValue,
			GrossTransportFeeValue: grossTransportFeeValue,

			TeacherPaymentID:      entity.TeacherPaymentID_None,
			PaidCourseFeeValue:    0,
			PaidTransportFeeValue: 0,
			AddedAt:               time.Time{},
		})
	}
}

func calculateGrossCourseAndTransportFeeValue(attendance entity.Attendance) (int32, int32) {
	grossCourseFeeValue := int32(float64(attendance.StudentLearningToken.CourseFeeValue) * attendance.UsedStudentTokenQuota / float64(Default_OneCourseCycle))
	grossTransportFeeValue := int32(float64(attendance.StudentLearningToken.TransportFeeValue) * attendance.UsedStudentTokenQuota / float64(Default_OneCourseCycle))
	return grossCourseFeeValue, grossTransportFeeValue
}

func (b *teacherPaymentInvoiceItemBuilder) AddTeacherPayments(teacherPayments []entity.TeacherPayment) {
	for _, teacherPayment := range teacherPayments {
		b.rawItems = append(b.rawItems, teacherPaymentInvoiceItemRaw{
			Attendance:             teacherPayment.Attendance,
			GrossCourseFeeValue:    teacherPayment.GrossCourseFeeValue,
			GrossTransportFeeValue: teacherPayment.GrossTransportFeeValue,

			TeacherPaymentID:      teacherPayment.TeacherPaymentID,
			PaidCourseFeeValue:    teacherPayment.PaidCourseFeeValue,
			PaidTransportFeeValue: teacherPayment.PaidTransportFeeValue,
			AddedAt:               teacherPayment.AddedAt,
		})
	}
}

func (b *teacherPaymentInvoiceItemBuilder) Build() []TeacherPaymentInvoiceItem {
	teacherPaymentInvoiceItems := make([]TeacherPaymentInvoiceItem, 0, 1)

	classIdToAttendanceIds := make(map[entity.ClassID][]entity.AttendanceID, 1)
	for _, teacherPaymentInvoiceItemRaw := range b.rawItems {
		attendanceId := teacherPaymentInvoiceItemRaw.Attendance.AttendanceID
		// we need this to prevent Go from reusing address of teacherPaymentInvoiceItemraw
		//   -> resulting in all elements of b.attendanceIdToTPIIRaw to contain same elements
		tpiiRaw := teacherPaymentInvoiceItemRaw
		b.attendanceIdToTPIIRaw[attendanceId] = &tpiiRaw

		class := teacherPaymentInvoiceItemRaw.Attendance.ClassInfo
		student := teacherPaymentInvoiceItemRaw.Attendance.StudentInfo
		slt := teacherPaymentInvoiceItemRaw.Attendance.StudentLearningToken

		b.classIdToClass[class.ClassID] = class
		b.sltIdToSLT[slt.StudentLearningTokenID] = slt
		b.studentIdToStudent[student.StudentID] = student

		if prevValues, ok := classIdToAttendanceIds[class.ClassID]; ok {
			classIdToAttendanceIds[class.ClassID] = append(prevValues, attendanceId)
		} else {
			classIdToAttendanceIds[class.ClassID] = []entity.AttendanceID{attendanceId}
		}
	}

	for classId, _attendanceIds := range classIdToAttendanceIds {
		students := b.createTPIIStudents(_attendanceIds)
		teacherPaymentInvoiceItems = append(teacherPaymentInvoiceItems, TeacherPaymentInvoiceItem{
			ClassInfo_Minimal: b.classIdToClass[classId],
			Students:          students,
		})
	}
	sort.SliceStable(teacherPaymentInvoiceItems, func(i, j int) bool {
		return teacherPaymentInvoiceItems[i].ClassInfo_Minimal.String() < teacherPaymentInvoiceItems[j].ClassInfo_Minimal.String()
	})

	return teacherPaymentInvoiceItems
}

func (b *teacherPaymentInvoiceItemBuilder) createTPIIStudents(attendanceIds []entity.AttendanceID) []tpii_Student {
	tpiiStudents := make([]tpii_Student, 0, 1)

	studentIdToAttendanceIds := make(map[entity.StudentID][]entity.AttendanceID, 2)
	for _, attendanceId := range attendanceIds {
		tpiiRaw := *b.attendanceIdToTPIIRaw[attendanceId]
		studentId := tpiiRaw.Attendance.StudentInfo.StudentID

		if prevValues, ok := studentIdToAttendanceIds[studentId]; ok {
			studentIdToAttendanceIds[studentId] = append(prevValues, attendanceId)
		} else {
			studentIdToAttendanceIds[studentId] = []entity.AttendanceID{attendanceId}
		}
	}

	for studentId, _attendanceIds := range studentIdToAttendanceIds {
		slts := b.createTPIIStudentLearningTokens(_attendanceIds)
		tpiiStudents = append(tpiiStudents, tpii_Student{
			StudentInfo_Minimal:   b.studentIdToStudent[studentId],
			StudentLearningTokens: slts,
		})
	}
	sort.SliceStable(tpiiStudents, func(i, j int) bool {
		return tpiiStudents[i].StudentInfo_Minimal.String() < tpiiStudents[j].StudentInfo_Minimal.String()
	})

	return tpiiStudents
}

func (b *teacherPaymentInvoiceItemBuilder) createTPIIStudentLearningTokens(attendanceIds []entity.AttendanceID) []tpii_StudentLearningToken {
	tpii_SLTS := make([]tpii_StudentLearningToken, 0, 1)

	sltIdToAttendanceIds := make(map[entity.StudentLearningTokenID][]entity.AttendanceID, 2)
	for _, attendanceId := range attendanceIds {
		tpiiRaw := *b.attendanceIdToTPIIRaw[attendanceId]
		sltId := tpiiRaw.Attendance.StudentLearningToken.StudentLearningTokenID

		if prevValues, ok := sltIdToAttendanceIds[sltId]; ok {
			sltIdToAttendanceIds[sltId] = append(prevValues, attendanceId)
		} else {
			sltIdToAttendanceIds[sltId] = []entity.AttendanceID{attendanceId}
		}
	}

	for sltId, _attendanceIds := range sltIdToAttendanceIds {
		attendances := make([]tpii_AttendanceWithTeacherPayment, 0, len(_attendanceIds))
		for _, attendanceId := range _attendanceIds {
			attendances = append(attendances, b.attendanceIdToTPIIRaw[attendanceId].toTPIIAttendanceWithTeacherPayment())
		}

		tpii_SLTS = append(tpii_SLTS, tpii_StudentLearningToken{
			StudentLearningToken_Minimal: b.sltIdToSLT[sltId],
			Attendances:                  attendances,
		})
	}
	sort.SliceStable(tpii_SLTS, func(i, j int) bool {
		return tpii_SLTS[i].StudentLearningTokenID < tpii_SLTS[j].StudentLearningTokenID
	})

	return tpii_SLTS
}
