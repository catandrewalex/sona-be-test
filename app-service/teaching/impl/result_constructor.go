package teaching

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
)

func NewTeacherForPaymentsFromGetTeachersForPaymentsRow(teacherRows []mysql.GetTeachersForPaymentsRow) []teaching.TeacherForPayment {
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
			TotalAttendances: int32(teacherRow.TotalAttendances.(float64)),
		})
	}

	return teachers
}
