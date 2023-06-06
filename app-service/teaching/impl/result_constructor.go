package impl

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/app-service/util"
)

func NewTeachersFromGetTeachersRow(teacherRows []mysql.GetTeachersRow) []teaching.Teacher {
	teachers := make([]teaching.Teacher, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		teachers = append(teachers, teaching.Teacher{
			TeacherID: teaching.TeacherID(teacherRow.ID),
			User: identity.User{
				ID:            identity.UserID(teacherRow.UserID),
				Username:      teacherRow.Username,
				Email:         teacherRow.Email,
				UserDetail:    identity.UnmarshalUserDetail(teacherRow.UserDetail, mainLog),
				PrivilegeType: identity.UserPrivilegeType(teacherRow.PrivilegeType),
				IsDeactivated: util.Int32ToBool(teacherRow.IsDeactivated),
				CreatedAt:     teacherRow.CreatedAt.Time,
			},
		})
	}

	return teachers
}

func NewStudentsFromGetStudentsRow(studentRows []mysql.GetStudentsRow) []teaching.Student {
	students := make([]teaching.Student, 0, len(studentRows))
	for _, studentRow := range studentRows {
		students = append(students, teaching.Student{
			StudentID: teaching.StudentID(studentRow.ID),
			User: identity.User{
				ID:            identity.UserID(studentRow.UserID),
				Username:      studentRow.Username,
				Email:         studentRow.Email,
				UserDetail:    identity.UnmarshalUserDetail(studentRow.UserDetail, mainLog),
				PrivilegeType: identity.UserPrivilegeType(studentRow.PrivilegeType),
				IsDeactivated: util.Int32ToBool(studentRow.IsDeactivated),
				CreatedAt:     studentRow.CreatedAt.Time,
			},
		})
	}

	return students
}

func NewInstrumentsFromMySQLInstrument(instrumentRows []mysql.Instrument) []teaching.Instrument {
	instruments := make([]teaching.Instrument, 0, len(instrumentRows))
	for _, instrumentRow := range instrumentRows {
		instruments = append(instruments, teaching.Instrument{
			ID:   teaching.InstrumentID(instrumentRow.ID),
			Name: instrumentRow.Name,
		})
	}

	return instruments
}

func NewGradesFromMySQLGrade(gradeRows []mysql.Grade) []teaching.Grade {
	grades := make([]teaching.Grade, 0, len(gradeRows))
	for _, gradeRow := range gradeRows {
		grades = append(grades, teaching.Grade{
			ID:   teaching.GradeID(gradeRow.ID),
			Name: gradeRow.Name,
		})
	}

	return grades
}

func NewCoursesFromGetCoursesRow(courseRows []mysql.GetCoursesRow) []teaching.Course {
	courses := make([]teaching.Course, 0, len(courseRows))
	for _, courseRow := range courseRows {
		courses = append(courses, teaching.Course{
			ID:                    teaching.CourseID(courseRow.CourseID),
			CompleteName:          courseRow.CourseName,
			DefaultFee:            courseRow.DefaultFee,
			DefaultDurationMinute: courseRow.DefaultDurationMinute,
		})
	}

	return courses
}

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
