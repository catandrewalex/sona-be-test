package teaching

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
)

func NewTeacherForPaymentsFromGetUnpaidTeachersRow(teacherRows []mysql.GetUnpaidTeachersRow) []teaching.TeacherForPayment {
	teachers := make([]teaching.TeacherForPayment, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		teachers = append(teachers, teaching.TeacherForPayment{
			TeacherInfo_Minimal: entity.TeacherInfo_Minimal{
				TeacherID: entity.TeacherID(teacherRow.ID),
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   teacherRow.Username,
					UserDetail: identity.UnmarshalUserDetail(teacherRow.UserDetail, mainLog),
				},
			},
			TotalAttendances:             teacherRow.TotalAttendances.(float64),
			TotalAttendancesWithoutToken: teacherRow.TotalAttendancesWithoutToken.(float64),
		})
	}

	return teachers
}

func NewTeacherForPaymentsFromGetPaidTeachersRow(teacherRows []mysql.GetPaidTeachersRow) []teaching.TeacherForPayment {
	// `GetPaidTeachersRow` shares the same struct as `GetUnpaidTeachersRow`.
	// So, we can use this workaround to avoid copy-pasting code.
	temp := make([]mysql.GetUnpaidTeachersRow, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		temp = append(temp, mysql.GetUnpaidTeachersRow(teacherRow))
	}
	teachers := NewTeacherForPaymentsFromGetUnpaidTeachersRow(temp)

	return teachers
}
