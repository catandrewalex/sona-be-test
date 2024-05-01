package teaching

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
)

func NewUnpaidTeachersFromGetUnpaidTeachersRow(teacherRows []mysql.GetUnpaidTeachersRow) []teaching.UnpaidTeacher {
	teachers := make([]teaching.UnpaidTeacher, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		teachers = append(teachers, teaching.UnpaidTeacher{
			TeacherInfo_Minimal: entity.TeacherInfo_Minimal{
				TeacherID: entity.TeacherID(teacherRow.ID),
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   teacherRow.Username,
					UserDetail: identity.UnmarshalUserDetail(teacherRow.UserDetail, mainLog),
				},
			},
			TotalUnpaidAttendances: int32(teacherRow.TotalUnpaidAttendances.(float64)),
		})
	}

	return teachers
}
