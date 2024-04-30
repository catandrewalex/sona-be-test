package impl

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
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
				Email:         teacherRow.Email.String,
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
				Email:         studentRow.Email.String,
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

			course := NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{
				{
					CourseID:              classRow.CourseID,
					Instrument:            classRow.Instrument,
					Grade:                 classRow.Grade,
					DefaultFee:            classRow.DefaultFee,
					DefaultDurationMinute: classRow.DefaultDurationMinute,
				},
			})[0]

			classes = append(classes, entity.Class{
				ClassID:              classId,
				TeacherInfo_Minimal:  teacherInfo,
				StudentInfos_Minimal: studentInfos,
				Course:               course,
				TransportFee:         classRow.TransportFee,
				IsDeactivated:        util.Int32ToBool(classRow.IsDeactivated),
			})
			prevClassId = classId
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
		var classTeacherInfo *entity.TeacherInfo_Minimal
		teacherId := entity.TeacherID(studentEnrollmentRow.ClassTeacherID.Int64)
		if studentEnrollmentRow.ClassTeacherID.Valid && teacherId != entity.TeacherID_None {
			classTeacherInfo = &entity.TeacherInfo_Minimal{
				TeacherID: teacherId,
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   studentEnrollmentRow.ClassTeacherUsername.String,
					UserDetail: identity.UnmarshalUserDetail(studentEnrollmentRow.ClassTeacherDetail, mainLog),
				},
			}
		}

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
				ClassID:             entity.ClassID(studentEnrollmentRow.Class.ID),
				TeacherInfo_Minimal: classTeacherInfo,
				Course: NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{
					{
						CourseID:              studentEnrollmentRow.Course.ID,
						Instrument:            studentEnrollmentRow.Instrument,
						Grade:                 studentEnrollmentRow.Grade,
						DefaultFee:            studentEnrollmentRow.Course.DefaultFee,
						DefaultDurationMinute: studentEnrollmentRow.Course.DefaultDurationMinute,
					},
				})[0],
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
			Course: NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{
				{
					CourseID:              teacherSpecialFeeRow.Course.ID,
					Instrument:            teacherSpecialFeeRow.Instrument,
					Grade:                 teacherSpecialFeeRow.Grade,
					DefaultFee:            teacherSpecialFeeRow.Course.DefaultFee,
					DefaultDurationMinute: teacherSpecialFeeRow.Course.DefaultDurationMinute,
				},
			})[0],
			Fee: teacherSpecialFeeRow.Fee,
		})
	}

	return teacherSpecialFees
}

func NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow(enrollmentPaymentRows []mysql.GetEnrollmentPaymentsRow) []entity.EnrollmentPayment {
	enrollmentPayments := make([]entity.EnrollmentPayment, 0, len(enrollmentPaymentRows))
	for _, enrollmentPaymentRow := range enrollmentPaymentRows {
		var classTeacherInfo *entity.TeacherInfo_Minimal
		teacherId := entity.TeacherID(enrollmentPaymentRow.ClassTeacherID.Int64)
		if enrollmentPaymentRow.ClassTeacherID.Valid && teacherId != entity.TeacherID_None {
			classTeacherInfo = &entity.TeacherInfo_Minimal{
				TeacherID: teacherId,
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   enrollmentPaymentRow.ClassTeacherUsername.String,
					UserDetail: identity.UnmarshalUserDetail(enrollmentPaymentRow.ClassTeacherDetail, mainLog),
				},
			}
		}

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
					ClassID:             entity.ClassID(enrollmentPaymentRow.Class.ID),
					TeacherInfo_Minimal: classTeacherInfo,
					Course: NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{
						{
							CourseID:              enrollmentPaymentRow.Course.ID,
							Instrument:            enrollmentPaymentRow.Instrument,
							Grade:                 enrollmentPaymentRow.Grade,
							DefaultFee:            enrollmentPaymentRow.Course.DefaultFee,
							DefaultDurationMinute: enrollmentPaymentRow.Course.DefaultDurationMinute,
						},
					})[0],
					TransportFee:  enrollmentPaymentRow.Class.TransportFee,
					IsDeactivated: util.Int32ToBool(enrollmentPaymentRow.Class.IsDeactivated),
				},
			},
			PaymentDate:       enrollmentPaymentRow.PaymentDate,
			BalanceTopUp:      enrollmentPaymentRow.BalanceTopUp,
			CourseFeeValue:    enrollmentPaymentRow.CourseFeeValue,
			TransportFeeValue: enrollmentPaymentRow.TransportFeeValue,
			PenaltyFeeValue:   enrollmentPaymentRow.PenaltyFeeValue,
		})
	}

	return enrollmentPayments
}

func NewStudentLearningTokensFromGetStudentLearningTokensRow(studentLearningTokenRows []mysql.GetStudentLearningTokensRow) []entity.StudentLearningToken {
	studentLearningTokens := make([]entity.StudentLearningToken, 0, len(studentLearningTokenRows))
	for _, sltRow := range studentLearningTokenRows {
		var classTeacherInfo *entity.TeacherInfo_Minimal
		teacherId := entity.TeacherID(sltRow.ClassTeacherID.Int64)
		if sltRow.ClassTeacherID.Valid && teacherId != entity.TeacherID_None {
			classTeacherInfo = &entity.TeacherInfo_Minimal{
				TeacherID: teacherId,
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   sltRow.ClassTeacherUsername.String,
					UserDetail: identity.UnmarshalUserDetail(sltRow.ClassTeacherDetail, mainLog),
				},
			}
		}

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
					ClassID:             entity.ClassID(sltRow.Class.ID),
					TeacherInfo_Minimal: classTeacherInfo,
					Course: NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{
						{
							CourseID:              sltRow.Course.ID,
							Instrument:            sltRow.Instrument,
							Grade:                 sltRow.Grade,
							DefaultFee:            sltRow.Course.DefaultFee,
							DefaultDurationMinute: sltRow.Course.DefaultDurationMinute,
						},
					})[0],
					TransportFee:  sltRow.Class.TransportFee,
					IsDeactivated: util.Int32ToBool(sltRow.Class.IsDeactivated),
				},
			},
			Quota:             sltRow.Quota,
			CourseFeeValue:    sltRow.CourseFeeValue,
			TransportFeeValue: sltRow.TransportFeeValue,
			CreatedAt:         sltRow.CreatedAt,
			LastUpdatedAt:     sltRow.LastUpdatedAt,
		})
	}

	return studentLearningTokens
}

func NewAttendancesFromGetAttendancesRow(attendanceRows []mysql.GetAttendancesRow) []entity.Attendance {
	attendances := make([]entity.Attendance, 0, len(attendanceRows))
	for _, attendanceRow := range attendanceRows {
		var classInfo *entity.ClassInfo_Minimal
		classId := entity.ClassID(attendanceRow.Class.ID)
		if classId != entity.ClassID(entity.ClassID_None) {
			// attendance.teacher & attendance.class.teacher may differ, as the class-registered teacher may be absent, and is replaced by another teacher
			var classTeacherInfo *entity.TeacherInfo_Minimal
			teacherId := entity.TeacherID(attendanceRow.ClassTeacherID.Int64)
			if attendanceRow.ClassTeacherID.Valid && teacherId != entity.TeacherID_None {
				classTeacherInfo = &entity.TeacherInfo_Minimal{
					TeacherID: teacherId,
					UserInfo_Minimal: identity.UserInfo_Minimal{
						Username:   attendanceRow.ClassTeacherUsername.String,
						UserDetail: identity.UnmarshalUserDetail(attendanceRow.ClassTeacherDetail, mainLog),
					},
				}
			}

			classInfo = &entity.ClassInfo_Minimal{
				ClassID:             entity.ClassID(attendanceRow.Class.ID),
				TeacherInfo_Minimal: classTeacherInfo,
				Course: NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{
					{
						CourseID:              attendanceRow.Course.ID,
						Instrument:            attendanceRow.Instrument,
						Grade:                 attendanceRow.Grade,
						DefaultFee:            attendanceRow.Course.DefaultFee,
						DefaultDurationMinute: attendanceRow.Course.DefaultDurationMinute,
					},
				})[0],
				TransportFee:  attendanceRow.Class.TransportFee,
				IsDeactivated: util.Int32ToBool(attendanceRow.Class.IsDeactivated),
			}
		}

		// attendance.teacher & attendance.class.teacher may differ, as the class-registered teacher may be absent, and is replaced by another teacher
		var teacherInfo *entity.TeacherInfo_Minimal
		teacherId := entity.TeacherID(attendanceRow.TeacherID.Int64)
		if attendanceRow.TeacherID.Valid && teacherId != entity.TeacherID_None {
			teacherInfo = &entity.TeacherInfo_Minimal{
				TeacherID: teacherId,
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   attendanceRow.TeacherUsername.String,
					UserDetail: identity.UnmarshalUserDetail(attendanceRow.TeacherDetail, mainLog),
				},
			}
		}

		var studentInfo *entity.StudentInfo_Minimal
		studentId := entity.StudentID(attendanceRow.StudentID.Int64)
		if attendanceRow.StudentID.Valid && studentId != entity.StudentID_None {
			studentInfo = &entity.StudentInfo_Minimal{
				StudentID: studentId,
				UserInfo_Minimal: identity.UserInfo_Minimal{
					Username:   attendanceRow.StudentUsername.String,
					UserDetail: identity.UnmarshalUserDetail(attendanceRow.StudentDetail, mainLog),
				},
			}
		}

		attendances = append(attendances, entity.Attendance{
			AttendanceID: entity.AttendanceID(attendanceRow.AttendanceID),
			ClassInfo:    classInfo,
			TeacherInfo:  teacherInfo,
			StudentInfo:  studentInfo,
			StudentLearningToken: entity.StudentLearningToken_Minimal{
				StudentLearningTokenID: entity.StudentLearningTokenID(attendanceRow.StudentLearningToken.ID),
				CourseFeeValue:         attendanceRow.StudentLearningToken.CourseFeeValue,
				TransportFeeValue:      attendanceRow.StudentLearningToken.TransportFeeValue,
				LastUpdatedAt:          attendanceRow.StudentLearningToken.LastUpdatedAt,
			},
			Date:                  attendanceRow.Date,
			UsedStudentTokenQuota: attendanceRow.UsedStudentTokenQuota,
			Duration:              attendanceRow.Duration,
			Note:                  attendanceRow.Note,
			IsPaid:                util.Int32ToBool(attendanceRow.IsPaid),
		})
	}

	return attendances
}

func NewTeacherSalariesFromGetTeacherSalariesRow(teacherSalaryRows []mysql.GetTeacherSalariesRow) []entity.TeacherSalary {
	teacherSalaries := make([]entity.TeacherSalary, 0, len(teacherSalaryRows))
	for _, tsRow := range teacherSalaryRows {
		teacherSalaries = append(teacherSalaries, entity.TeacherSalary{
			TeacherSalaryID: entity.TeacherSalaryID(tsRow.TeacherSalaryID),
			Attendance: NewAttendancesFromGetAttendancesRow([]mysql.GetAttendancesRow{
				{
					AttendanceID:          tsRow.Attendance.ID,
					Date:                  tsRow.Attendance.Date,
					UsedStudentTokenQuota: tsRow.Attendance.UsedStudentTokenQuota,
					Duration:              tsRow.Attendance.Duration,
					Note:                  tsRow.Attendance.Note,
					IsPaid:                tsRow.Attendance.IsPaid,
					Class:                 tsRow.Class,
					Course:                tsRow.Course,
					Instrument:            tsRow.Instrument,
					Grade:                 tsRow.Grade,
					TeacherID:             tsRow.TeacherID,
					TeacherUsername:       tsRow.TeacherUsername,
					TeacherDetail:         tsRow.TeacherDetail,
					StudentID:             tsRow.StudentID,
					StudentUsername:       tsRow.StudentUsername,
					StudentDetail:         tsRow.StudentDetail,
					ClassTeacherID:        tsRow.ClassTeacherID,
					ClassTeacherUsername:  tsRow.ClassTeacherUsername,
					ClassTeacherDetail:    tsRow.ClassTeacherDetail,
					StudentLearningToken:  tsRow.StudentLearningToken,
				},
			})[0],
			PaidCourseFeeValue:     tsRow.PaidCourseFeeValue,
			PaidTransportFeeValue:  tsRow.PaidTransportFeeValue,
			AddedAt:                tsRow.AddedAt,
			GrossCourseFeeValue:    int32(float64(tsRow.StudentLearningToken.CourseFeeValue) * tsRow.Attendance.UsedStudentTokenQuota / float64(teaching.Default_OneCourseCycle)),
			GrossTransportFeeValue: int32(float64(tsRow.StudentLearningToken.TransportFeeValue) * tsRow.Attendance.UsedStudentTokenQuota / float64(teaching.Default_OneCourseCycle)),
		})
	}

	return teacherSalaries
}
