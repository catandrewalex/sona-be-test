package impl

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
)

func NewTeachersFromGetTeachersRow(teacherRows []mysql.GetTeachersRow) []entity.Teacher {
	teachers := make([]entity.Teacher, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		teachers = append(teachers, entity.Teacher{
			TeacherID: entity.TeacherID(teacherRow.ID),
			User: identity.User{
				UserID:        identity.UserID(teacherRow.UserID),
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

func NewStudentsFromGetStudentsRow(studentRows []mysql.GetStudentsRow) []entity.Student {
	students := make([]entity.Student, 0, len(studentRows))
	for _, studentRow := range studentRows {
		students = append(students, entity.Student{
			StudentID: entity.StudentID(studentRow.ID),
			User: identity.User{
				UserID:        identity.UserID(studentRow.UserID),
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

func NewInstrumentsFromMySQLInstruments(instrumentRows []mysql.Instrument) []entity.Instrument {
	instruments := make([]entity.Instrument, 0, len(instrumentRows))
	for _, instrumentRow := range instrumentRows {
		instruments = append(instruments, entity.Instrument{
			InstrumentID: entity.InstrumentID(instrumentRow.ID),
			Name:         instrumentRow.Name,
		})
	}

	return instruments
}

func NewGradesFromMySQLGrades(gradeRows []mysql.Grade) []entity.Grade {
	grades := make([]entity.Grade, 0, len(gradeRows))
	for _, gradeRow := range gradeRows {
		grades = append(grades, entity.Grade{
			GradeID: entity.GradeID(gradeRow.ID),
			Name:    gradeRow.Name,
		})
	}

	return grades
}

func NewCoursesFromGetCoursesRow(courseRows []mysql.GetCoursesRow) []entity.Course {
	courses := make([]entity.Course, 0, len(courseRows))
	for _, courseRow := range courseRows {
		courses = append(courses, entity.Course{
			CourseID:              entity.CourseID(courseRow.CourseID),
			Instrument:            NewInstrumentsFromMySQLInstruments([]mysql.Instrument{courseRow.Instrument})[0],
			Grade:                 NewGradesFromMySQLGrades([]mysql.Grade{courseRow.Grade})[0],
			DefaultFee:            courseRow.DefaultFee,
			DefaultDurationMinute: courseRow.DefaultDurationMinute,
		})
	}

	return courses
}

func NewClassesFromGetClassesRow(classRows []mysql.GetClassesRow) []entity.Class {
	classes := make([]entity.Class, 0, len(classRows))

	prevClassId := entity.ClassID_None
	for _, classRow := range classRows {
		classId := entity.ClassID(classRow.ClassID)
		if classId != prevClassId {
			var teacherInfo *entity.TeacherInfo_Minimal
			teacherId := entity.TeacherID(classRow.TeacherID.Int64)
			if classRow.TeacherID.Valid && teacherId != entity.TeacherID_None {
				teacherInfo = &entity.TeacherInfo_Minimal{
					TeacherID: teacherId,
					UserInfo_Minimal: identity.UserInfo_Minimal{
						Username:   classRow.TeacherUsername.String,
						UserDetail: identity.UnmarshalUserDetail(classRow.TeacherDetail, mainLog),
					},
				}
			}

			studentInfos := make([]entity.StudentInfo_Minimal, 0)
			studentId := entity.StudentID(classRow.StudentID.Int64)
			if classRow.StudentID.Valid && studentId != entity.StudentID_None {
				studentInfos = append(studentInfos, entity.StudentInfo_Minimal{
					StudentID: studentId,
					UserInfo_Minimal: identity.UserInfo_Minimal{
						Username:   classRow.StudentUsername.String,
						UserDetail: identity.UnmarshalUserDetail(classRow.StudentDetail, mainLog),
					},
				})
			}

			course := entity.Course{
				CourseID:              entity.CourseID(classRow.CourseID),
				Instrument:            NewInstrumentsFromMySQLInstruments([]mysql.Instrument{classRow.Instrument})[0],
				Grade:                 NewGradesFromMySQLGrades([]mysql.Grade{classRow.Grade})[0],
				DefaultFee:            classRow.DefaultFee,
				DefaultDurationMinute: classRow.DefaultDurationMinute,
			}

			classes = append(classes, entity.Class{
				ClassID:              classId,
				TeacherInfo_Minimal:  teacherInfo,
				StudentInfos_Minimal: studentInfos,
				Course:               course,
				TransportFee:         classRow.TransportFee,
				IsDeactivated:        util.Int32ToBool(classRow.IsDeactivated),
			})
		} else {
			// Populate students
			studentId := entity.StudentID(classRow.StudentID.Int64)
			if classRow.StudentID.Valid && studentId != entity.StudentID_None {
				prevIdx := len(classes) - 1
				classes[prevIdx].StudentInfos_Minimal = append(classes[prevIdx].StudentInfos_Minimal, entity.StudentInfo_Minimal{
					StudentID: studentId,
					UserInfo_Minimal: identity.UserInfo_Minimal{
						Username:   classRow.StudentUsername.String,
						UserDetail: identity.UnmarshalUserDetail(classRow.StudentDetail, mainLog),
					},
				})
			}
		}
	}

	return classes
}

func NewStudentEnrollmentsFromGetStudentEnrollmentsRow(studentEnrollmentRows []mysql.GetStudentEnrollmentsRow) []entity.StudentEnrollment {
	studentEnrollments := make([]entity.StudentEnrollment, 0, len(studentEnrollmentRows))
	for _, studentEnrollmentRow := range studentEnrollmentRows {
		studentEnrollments = append(studentEnrollments, entity.StudentEnrollment{
			StudentEnrollmentID: entity.StudentEnrollmentID(studentEnrollmentRow.StudentEnrollmentID),
			StudentInfo: entity.StudentInfo_Minimal{
				StudentID: entity.StudentID(studentEnrollmentRow.StudentID),
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   studentEnrollmentRow.StudentUsername,
					UserDetail: identity.UnmarshalUserDetail(studentEnrollmentRow.StudentDetail, mainLog),
				},
			},
			ClassInfo: entity.ClassInfo_Minimal{
				ClassID: entity.ClassID(studentEnrollmentRow.Class.ID),
				CourseInfo: entity.CourseInfo_Minimal{
					CourseID:   entity.CourseID(studentEnrollmentRow.Course.ID),
					Instrument: NewInstrumentsFromMySQLInstruments([]mysql.Instrument{studentEnrollmentRow.Instrument})[0],
					Grade:      NewGradesFromMySQLGrades([]mysql.Grade{studentEnrollmentRow.Grade})[0],
				},
				TransportFee:  studentEnrollmentRow.Class.TransportFee,
				IsDeactivated: util.Int32ToBool(studentEnrollmentRow.Class.IsDeactivated),
			},
		})
	}

	return studentEnrollments
}

func NewTeacherSpecialFeesFromGetTeacherSpecialFeesRow(teacherSpecialFeeRows []mysql.GetTeacherSpecialFeesRow) []entity.TeacherSpecialFee {
	teacherSpecialFees := make([]entity.TeacherSpecialFee, 0, len(teacherSpecialFeeRows))
	for _, teacherSpecialFeeRow := range teacherSpecialFeeRows {
		teacherSpecialFees = append(teacherSpecialFees, entity.TeacherSpecialFee{
			TeacherSpecialFeeID: entity.TeacherSpecialFeeID(teacherSpecialFeeRow.TeacherSpecialFeeID),
			TeacherInfo: entity.TeacherInfo_Minimal{
				TeacherID: entity.TeacherID(teacherSpecialFeeRow.TeacherID),
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   teacherSpecialFeeRow.TeacherUsername,
					UserDetail: identity.UnmarshalUserDetail(teacherSpecialFeeRow.TeacherDetail, mainLog),
				},
			},
			CourseInfo: entity.CourseInfo_Minimal{
				CourseID:   entity.CourseID(teacherSpecialFeeRow.CourseID),
				Instrument: NewInstrumentsFromMySQLInstruments([]mysql.Instrument{teacherSpecialFeeRow.Instrument})[0],
				Grade:      NewGradesFromMySQLGrades([]mysql.Grade{teacherSpecialFeeRow.Grade})[0],
			},
			Fee: teacherSpecialFeeRow.Fee,
		})
	}

	return teacherSpecialFees
}

func NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow(enrollmentPaymentRows []mysql.GetEnrollmentPaymentsRow) []entity.EnrollmentPayment {
	enrollmentPayments := make([]entity.EnrollmentPayment, 0, len(enrollmentPaymentRows))
	for _, enrollmentPaymentRow := range enrollmentPaymentRows {
		enrollmentPayments = append(enrollmentPayments, entity.EnrollmentPayment{
			EnrollmentPaymentID: entity.EnrollmentPaymentID(enrollmentPaymentRow.EnrollmentPaymentID),
			StudentEnrollmentInfo: entity.StudentEnrollment{
				StudentEnrollmentID: entity.StudentEnrollmentID(enrollmentPaymentRow.StudentEnrollmentID),
				StudentInfo: entity.StudentInfo_Minimal{
					StudentID: entity.StudentID(enrollmentPaymentRow.StudentID),
					UserInfo_Minimal: identity.UserInfo_Minimal{
						Username:   enrollmentPaymentRow.StudentUsername,
						UserDetail: identity.UnmarshalUserDetail(enrollmentPaymentRow.StudentDetail, mainLog),
					},
				},
				ClassInfo: entity.ClassInfo_Minimal{
					ClassID: entity.ClassID(enrollmentPaymentRow.Class.ID),
					CourseInfo: entity.CourseInfo_Minimal{
						CourseID:   entity.CourseID(enrollmentPaymentRow.Course.ID),
						Instrument: NewInstrumentsFromMySQLInstruments([]mysql.Instrument{enrollmentPaymentRow.Instrument})[0],
						Grade:      NewGradesFromMySQLGrades([]mysql.Grade{enrollmentPaymentRow.Grade})[0],
					},
					TransportFee:  enrollmentPaymentRow.Class.TransportFee,
					IsDeactivated: util.Int32ToBool(enrollmentPaymentRow.Class.IsDeactivated),
				},
			},
			PaymentDate:  enrollmentPaymentRow.PaymentDate,
			BalanceTopUp: enrollmentPaymentRow.BalanceTopUp,
			Value:        enrollmentPaymentRow.Value,
			ValuePenalty: enrollmentPaymentRow.ValuePenalty,
		})
	}

	return enrollmentPayments
}

func NewStudentLearningTokensFromGetStudentLearningTokensRow(studentLearningTokenRows []mysql.GetStudentLearningTokensRow) []entity.StudentLearningToken {
	studentLearningTokens := make([]entity.StudentLearningToken, 0, len(studentLearningTokenRows))
	for _, sltRow := range studentLearningTokenRows {
		studentLearningTokens = append(studentLearningTokens, entity.StudentLearningToken{
			StudentLearningTokenID: entity.StudentLearningTokenID(sltRow.StudentLearningTokenID),
			StudentEnrollmentInfo: entity.StudentEnrollment{
				StudentEnrollmentID: entity.StudentEnrollmentID(sltRow.StudentEnrollmentID),
				StudentInfo: entity.StudentInfo_Minimal{
					StudentID: entity.StudentID(sltRow.StudentID),
					UserInfo_Minimal: identity.UserInfo_Minimal{
						Username:   sltRow.StudentUsername,
						UserDetail: identity.UnmarshalUserDetail(sltRow.StudentDetail, mainLog),
					},
				},
				ClassInfo: entity.ClassInfo_Minimal{
					ClassID: entity.ClassID(sltRow.Class.ID),
					CourseInfo: entity.CourseInfo_Minimal{
						CourseID:   entity.CourseID(sltRow.Course.ID),
						Instrument: NewInstrumentsFromMySQLInstruments([]mysql.Instrument{sltRow.Instrument})[0],
						Grade:      NewGradesFromMySQLGrades([]mysql.Grade{sltRow.Grade})[0],
					},
					TransportFee:  sltRow.Class.TransportFee,
					IsDeactivated: util.Int32ToBool(sltRow.Class.IsDeactivated),
				},
			},
			Quota:             sltRow.Quota,
			QuotaBonus:        sltRow.QuotaBonus,
			CourseFeeValue:    sltRow.CourseFeeValue,
			TransportFeeValue: sltRow.TransportFeeValue,
			LastUpdatedAt:     sltRow.LastUpdatedAt,
		})
	}

	return studentLearningTokens
}

func NewPresencesFromGetPresencesRow(presenceRows []mysql.GetPresencesRow) []entity.Presence {
	presences := make([]entity.Presence, 0, len(presenceRows))
	for _, presenceRow := range presenceRows {
		var classInfo *entity.ClassInfo_Minimal
		classId := entity.ClassID(presenceRow.Class.ID)
		if classId != entity.ClassID(entity.ClassID_None) {
			classInfo = &entity.ClassInfo_Minimal{
				ClassID: entity.ClassID(presenceRow.Class.ID),
				CourseInfo: entity.CourseInfo_Minimal{
					CourseID:   entity.CourseID(presenceRow.Course.ID),
					Instrument: NewInstrumentsFromMySQLInstruments([]mysql.Instrument{presenceRow.Instrument})[0],
					Grade:      NewGradesFromMySQLGrades([]mysql.Grade{presenceRow.Grade})[0],
				},
				TransportFee:  presenceRow.Class.TransportFee,
				IsDeactivated: util.Int32ToBool(presenceRow.Class.IsDeactivated),
			}
		}

		var teacherInfo *entity.TeacherInfo_Minimal
		teacherId := entity.TeacherID(presenceRow.TeacherID.Int64)
		if presenceRow.TeacherID.Valid && teacherId != entity.TeacherID_None {
			teacherInfo = &entity.TeacherInfo_Minimal{
				TeacherID: teacherId,
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   presenceRow.TeacherUsername.String,
					UserDetail: identity.UnmarshalUserDetail(presenceRow.TeacherDetail, mainLog),
				},
			}
		}

		var studentInfo *entity.StudentInfo_Minimal
		studentId := entity.StudentID(presenceRow.StudentID.Int64)
		if presenceRow.StudentID.Valid && studentId != entity.StudentID_None {
			studentInfo = &entity.StudentInfo_Minimal{
				StudentID: studentId,
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   presenceRow.StudentUsername.String,
					UserDetail: identity.UnmarshalUserDetail(presenceRow.StudentDetail, mainLog),
				},
			}
		}

		presences = append(presences, entity.Presence{
			PresenceID:  entity.PresenceID(presenceRow.PresenceID),
			ClassInfo:   classInfo,
			TeacherInfo: teacherInfo,
			StudentInfo: studentInfo,
			StudentLearningToken: entity.StudentLearningToken_Minimal{
				StudentLearningTokenID: entity.StudentLearningTokenID(presenceRow.StudentLearningToken.ID),
				Quota:                  presenceRow.StudentLearningToken.Quota,
				QuotaBonus:             presenceRow.StudentLearningToken.QuotaBonus,
				CourseFeeValue:         presenceRow.StudentLearningToken.CourseFeeValue,
				TransportFeeValue:      presenceRow.StudentLearningToken.TransportFeeValue,
				LastUpdatedAt:          presenceRow.StudentLearningToken.LastUpdatedAt,
			},
			Date:                  presenceRow.Date,
			UsedStudentTokenQuota: presenceRow.UsedStudentTokenQuota,
			Duration:              presenceRow.Duration,
		})
	}

	return presences
}
