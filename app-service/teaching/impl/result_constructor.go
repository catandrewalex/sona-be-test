package impl

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/app-service/util"
)

func NewClassesFromGetClassesRow(classRows []mysql.GetClassesRow) []teaching.Class {
	classes := make([]teaching.Class, 0, len(classRows))

	prevClassId := teaching.ClassID_None
	for _, classRow := range classRows {
		if classRow.ClassID != int64(prevClassId) {
			var teacherInfo *teaching.Class_TeacherInfo
			teacherId := teaching.TeacherID(classRow.TeacherID.Int64)
			if classRow.TeacherID.Valid && teacherId != teaching.TeacherID_None {
				teacherInfo = &teaching.Class_TeacherInfo{
					TeacherID:  teacherId,
					Username:   classRow.TeacherUsername.String,
					UserDetail: identity.UnmarshalUserDetail(classRow.TeacherDetail, mainLog),
				}
			}

			studentEnrollments := make([]teaching.Class_StudentEnrollmentInfo, 0)
			studentId := teaching.StudentID(classRow.StudentID.Int64)
			if classRow.StudentID.Valid && studentId != teaching.StudentID_None {
				studentEnrollments = append(studentEnrollments, teaching.Class_StudentEnrollmentInfo{
					ID: teaching.StudentEnrollmentID(classRow.EnrollmentID.Int64),
					StudentInfo: teaching.Enrollment_StudentInfo{
						StudentID:  studentId,
						Username:   classRow.StudentUsername.String,
						UserDetail: identity.UnmarshalUserDetail(classRow.StudentDetail, mainLog),
					},
				})
			}

			course := teaching.Course{
				ID:                    teaching.CourseID(classRow.CourseID),
				CompleteName:          classRow.CourseName,
				DefaultFee:            classRow.DefaultFee,
				DefaultDurationMinute: classRow.DefaultDurationMinute,
			}

			classes = append(classes, teaching.Class{
				ID:                 teaching.ClassID(classRow.ClassID),
				TeacherInfo:        teacherInfo,
				StudentEnrollments: studentEnrollments,
				Course:             course,
				TransportFee:       classRow.TransportFee,
				IsDeactivated:      util.Int32ToBool(classRow.IsDeactivated),
			})
		} else {
			// Populate students
			studentId := teaching.StudentID(classRow.StudentID.Int64)
			if classRow.StudentID.Valid && studentId != teaching.StudentID_None {
				prevIdx := len(classes) - 1
				classes[prevIdx].StudentEnrollments = append(classes[prevIdx].StudentEnrollments, teaching.Class_StudentEnrollmentInfo{
					ID: teaching.StudentEnrollmentID(classRow.EnrollmentID.Int64),
					StudentInfo: teaching.Enrollment_StudentInfo{
						StudentID:  studentId,
						Username:   classRow.StudentUsername.String,
						UserDetail: identity.UnmarshalUserDetail(classRow.StudentDetail, mainLog),
					},
				})
			}
		}
	}

	return classes
}
