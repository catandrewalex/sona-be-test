package impl

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/app-service/util"
	"sonamusica-backend/config"
	"sonamusica-backend/errs"
	"sonamusica-backend/logging"
	"sonamusica-backend/network"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("TeachingService", logging.GetLevel(configObject.LogLevel))
)

type teachingServiceImpl struct {
	mySQLQueries *relational_db.MySQLQueries

	identityService identity.IdentityService
}

var _ teaching.TeachingService = (*teachingServiceImpl)(nil)

func NewTeachingServiceImpl(mySQLQueries *relational_db.MySQLQueries, identityService identity.IdentityService) *teachingServiceImpl {
	return &teachingServiceImpl{
		mySQLQueries:    mySQLQueries,
		identityService: identityService,
	}
}

func (s teachingServiceImpl) GetTeachers(ctx context.Context, pagination util.PaginationSpec) (teaching.GetTeachersResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	teacherRows, err := s.mySQLQueries.GetTeachers(ctx, mysql.GetTeachersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return teaching.GetTeachersResult{}, fmt.Errorf("mySQLQueries.GetTeachers(): %w", err)
	}

	teachers := make([]teaching.Teacher, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		var userDetail identity.UserDetail
		err = json.Unmarshal(teacherRow.UserDetail, &userDetail)
		if err != nil {
			return teaching.GetTeachersResult{}, fmt.Errorf("json.Unmarshal(): %w", err)
		}

		teachers = append(teachers, teaching.Teacher{
			TeacherID: teaching.TeacherID(teacherRow.ID),
			User: identity.User{
				ID:            identity.UserID(teacherRow.UserID),
				Username:      teacherRow.Username,
				Email:         teacherRow.Email,
				UserDetail:    userDetail,
				PrivilegeType: identity.UserPrivilegeType(teacherRow.PrivilegeType),
				IsDeactivated: util.Int32ToBool(teacherRow.IsDeactivated),
				CreatedAt:     teacherRow.CreatedAt.Time,
			},
		})
	}

	totalResults, err := s.mySQLQueries.CountTeachers(ctx)

	return teaching.GetTeachersResult{
		Teachers:         teachers,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetTeacherById(ctx context.Context, id teaching.TeacherID) (teaching.Teacher, error) {
	teacher, err := s.mySQLQueries.GetTeacherById(ctx, int64(id))
	if err != nil {
		return teaching.Teacher{}, fmt.Errorf("mySQLQueries.GetTeacherById(): %w", err)
	}

	var userDetail identity.UserDetail
	err = json.Unmarshal(teacher.UserDetail, &userDetail)
	if err != nil {
		return teaching.Teacher{}, fmt.Errorf("json.Unmarshal(): %w", err)
	}

	return teaching.Teacher{
		TeacherID: teaching.TeacherID(teacher.ID),
		User: identity.User{
			ID:            identity.UserID(teacher.UserID),
			Username:      teacher.Username,
			Email:         teacher.Email,
			UserDetail:    userDetail,
			PrivilegeType: identity.UserPrivilegeType(teacher.PrivilegeType),
			IsDeactivated: util.Int32ToBool(teacher.IsDeactivated),
			CreatedAt:     teacher.CreatedAt.Time,
		},
	}, nil
}

func (s teachingServiceImpl) GetTeachersByIds(ctx context.Context, ids []teaching.TeacherID) ([]teaching.Teacher, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	teacherRows, err := s.mySQLQueries.GetTeachersByIds(ctx, idsInt)
	if err != nil {
		return []teaching.Teacher{}, fmt.Errorf("mySQLQueries.GetTeachersByIds(): %w", err)
	}

	teachers := make([]teaching.Teacher, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		var userDetail identity.UserDetail
		err = json.Unmarshal(teacherRow.UserDetail, &userDetail)
		if err != nil {
			return []teaching.Teacher{}, fmt.Errorf("json.Unmarshal(): %w", err)
		}

		teachers = append(teachers, teaching.Teacher{
			TeacherID: teaching.TeacherID(teacherRow.ID),
			User: identity.User{
				ID:            identity.UserID(teacherRow.UserID),
				Username:      teacherRow.Username,
				Email:         teacherRow.Email,
				UserDetail:    userDetail,
				PrivilegeType: identity.UserPrivilegeType(teacherRow.PrivilegeType),
				IsDeactivated: util.Int32ToBool(teacherRow.IsDeactivated),
				CreatedAt:     teacherRow.CreatedAt.Time,
			},
		})
	}

	return teachers, nil
}

func (s teachingServiceImpl) InsertTeachers(ctx context.Context, userIDs []identity.UserID) ([]teaching.TeacherID, error) {
	teacherIDs := make([]teaching.TeacherID, 0, len(userIDs))

	for _, userID := range userIDs {
		teacherID, err := s.mySQLQueries.InsertTeacher(ctx, int64(userID))
		if err != nil {
			return []teaching.TeacherID{}, fmt.Errorf("qtx.InsertTeacher(): %w", err)
		}
		teacherIDs = append(teacherIDs, teaching.TeacherID(teacherID))
	}

	return teacherIDs, nil
}

func (s teachingServiceImpl) InsertTeachersWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]teaching.TeacherID, error) {
	teacherIDs := make([]teaching.TeacherID, 0, len(specs))

	// TODO: move all mySQLQueries.* (Begin, Commit, etc.) to a new accessor service in lower level
	tx, err := s.mySQLQueries.DB.BeginTx(ctx, nil)
	if err != nil {
		return []teaching.TeacherID{}, fmt.Errorf("mySQLQueries.DB.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTx(tx)

	ctxWithSQLTx := network.NewContextWithSQLTx(ctx, tx)
	userIDs, err := s.identityService.InsertUsers(ctxWithSQLTx, specs)
	if err != nil {
		return []teaching.TeacherID{}, fmt.Errorf("identityService.InsertUsers(): %w", err)
	}

	for _, userID := range userIDs {
		teacherID, err := qtx.InsertTeacher(ctx, int64(userID))
		wrappedErr := errs.WrapMySQLError(err)
		if wrappedErr != nil {
			return []teaching.TeacherID{}, fmt.Errorf("qtx.InsertTeacher(): %w", wrappedErr)
		}
		teacherIDs = append(teacherIDs, teaching.TeacherID(teacherID))
	}

	err = tx.Commit()
	if err != nil {
		return []teaching.TeacherID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return teacherIDs, nil
}

func (s teachingServiceImpl) DeleteTeachers(ctx context.Context, ids []teaching.TeacherID) error {
	idsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt64 = append(idsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteTeachersByIds(ctx, idsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteTeacherByIds(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) GetStudents(ctx context.Context, pagination util.PaginationSpec) (teaching.GetStudentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	studentRows, err := s.mySQLQueries.GetStudents(ctx, mysql.GetStudentsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return teaching.GetStudentsResult{}, fmt.Errorf("mySQLQueries.GetStudents(): %w", err)
	}

	students := make([]teaching.Student, 0, len(studentRows))
	for _, studentRow := range studentRows {
		var userDetail identity.UserDetail
		err = json.Unmarshal(studentRow.UserDetail, &userDetail)
		if err != nil {
			return teaching.GetStudentsResult{}, fmt.Errorf("json.Unmarshal(): %w", err)
		}

		students = append(students, teaching.Student{
			StudentID: teaching.StudentID(studentRow.ID),
			User: identity.User{
				ID:            identity.UserID(studentRow.UserID),
				Username:      studentRow.Username,
				Email:         studentRow.Email,
				UserDetail:    userDetail,
				PrivilegeType: identity.UserPrivilegeType(studentRow.PrivilegeType),
				IsDeactivated: util.Int32ToBool(studentRow.IsDeactivated),
				CreatedAt:     studentRow.CreatedAt.Time,
			},
		})
	}

	totalResults, err := s.mySQLQueries.CountStudents(ctx)
	if err != nil {
		return teaching.GetStudentsResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return teaching.GetStudentsResult{
		Students:         students,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetStudentById(ctx context.Context, id teaching.StudentID) (teaching.Student, error) {
	student, err := s.mySQLQueries.GetStudentById(ctx, int64(id))
	if err != nil {
		return teaching.Student{}, fmt.Errorf("mySQLQueries.GetStudentById(): %w", err)
	}

	var userDetail identity.UserDetail
	err = json.Unmarshal(student.UserDetail, &userDetail)
	if err != nil {
		return teaching.Student{}, fmt.Errorf("json.Unmarshal(): %w", err)
	}

	return teaching.Student{
		StudentID: teaching.StudentID(student.ID),
		User: identity.User{
			ID:            identity.UserID(student.UserID),
			Username:      student.Username,
			Email:         student.Email,
			UserDetail:    userDetail,
			PrivilegeType: identity.UserPrivilegeType(student.PrivilegeType),
			IsDeactivated: util.Int32ToBool(student.IsDeactivated),
			CreatedAt:     student.CreatedAt.Time,
		},
	}, nil
}

func (s teachingServiceImpl) GetStudentsByIds(ctx context.Context, ids []teaching.StudentID) ([]teaching.Student, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	studentRows, err := s.mySQLQueries.GetStudentsByIds(ctx, idsInt)
	if err != nil {
		return []teaching.Student{}, fmt.Errorf("mySQLQueries.GetStudentsByIds(): %w", err)
	}

	students := make([]teaching.Student, 0, len(studentRows))
	for _, studentRow := range studentRows {
		var userDetail identity.UserDetail
		err = json.Unmarshal(studentRow.UserDetail, &userDetail)
		if err != nil {
			return []teaching.Student{}, fmt.Errorf("json.Unmarshal(): %w", err)
		}

		students = append(students, teaching.Student{
			StudentID: teaching.StudentID(studentRow.ID),
			User: identity.User{
				ID:            identity.UserID(studentRow.UserID),
				Username:      studentRow.Username,
				Email:         studentRow.Email,
				UserDetail:    userDetail,
				PrivilegeType: identity.UserPrivilegeType(studentRow.PrivilegeType),
				IsDeactivated: util.Int32ToBool(studentRow.IsDeactivated),
				CreatedAt:     studentRow.CreatedAt.Time,
			},
		})
	}

	return students, nil
}

func (s teachingServiceImpl) InsertStudents(ctx context.Context, userIDs []identity.UserID) ([]teaching.StudentID, error) {
	studentIDs := make([]teaching.StudentID, 0, len(userIDs))

	for _, userID := range userIDs {
		studentID, err := s.mySQLQueries.InsertStudent(ctx, int64(userID))
		if err != nil {
			return []teaching.StudentID{}, fmt.Errorf("qtx.InsertStudent(): %w", err)
		}
		studentIDs = append(studentIDs, teaching.StudentID(studentID))
	}

	return studentIDs, nil
}

func (s teachingServiceImpl) InsertStudentsWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]teaching.StudentID, error) {
	studentIDs := make([]teaching.StudentID, 0, len(specs))

	// TODO: move all mySQLQueries.* (Begin, Commit, etc.) to a new accessor service in lower level
	tx, err := s.mySQLQueries.DB.BeginTx(ctx, nil)
	if err != nil {
		return []teaching.StudentID{}, fmt.Errorf("mySQLQueries.DB.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTx(tx)

	ctxWithSQLTx := network.NewContextWithSQLTx(ctx, tx)
	userIDs, err := s.identityService.InsertUsers(ctxWithSQLTx, specs)
	if err != nil {
		return []teaching.StudentID{}, fmt.Errorf("identityService.InsertUsers(): %w", err)
	}

	for _, userID := range userIDs {
		studentID, err := qtx.InsertStudent(ctx, int64(userID))
		wrappedErr := errs.WrapMySQLError(err)
		if wrappedErr != nil {
			return []teaching.StudentID{}, fmt.Errorf("qtx.InsertStudent(): %w", wrappedErr)
		}
		studentIDs = append(studentIDs, teaching.StudentID(studentID))
	}

	err = tx.Commit()
	if err != nil {
		return []teaching.StudentID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return studentIDs, nil
}

func (s teachingServiceImpl) DeleteStudents(ctx context.Context, ids []teaching.StudentID) error {
	idsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt64 = append(idsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteStudentsByIds(ctx, idsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteStudentByIds(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) GetInstruments(ctx context.Context, pagination util.PaginationSpec) (teaching.GetInstrumentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	instrumentRows, err := s.mySQLQueries.GetInstruments(ctx, mysql.GetInstrumentsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return teaching.GetInstrumentsResult{}, fmt.Errorf("mySQLQueries.GetInstruments(): %w", err)
	}

	instruments := make([]teaching.Instrument, 0, len(instrumentRows))
	for _, instrumentRow := range instrumentRows {
		instruments = append(instruments, teaching.Instrument{
			ID:   teaching.InstrumentID(instrumentRow.ID),
			Name: instrumentRow.Name,
		})
	}

	totalResults, err := s.mySQLQueries.CountInstruments(ctx)

	return teaching.GetInstrumentsResult{
		Instruments:      instruments,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetInstrumentById(ctx context.Context, id teaching.InstrumentID) (teaching.Instrument, error) {
	instrument, err := s.mySQLQueries.GetInstrumentById(ctx, int64(id))
	if err != nil {
		return teaching.Instrument{}, fmt.Errorf("mySQLQueries.GetInstrumentById(): %w", err)
	}

	return teaching.Instrument{
		ID:   teaching.InstrumentID(instrument.ID),
		Name: instrument.Name,
	}, nil
}

func (s teachingServiceImpl) GetInstrumentsByIds(ctx context.Context, ids []teaching.InstrumentID) ([]teaching.Instrument, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	instrumentRows, err := s.mySQLQueries.GetInstrumentsByIds(ctx, idsInt)
	if err != nil {
		return []teaching.Instrument{}, fmt.Errorf("mySQLQueries.GetInstrumentsByIds(): %w", err)
	}

	instruments := make([]teaching.Instrument, 0, len(instrumentRows))
	for _, instrumentRow := range instrumentRows {
		instruments = append(instruments, teaching.Instrument{
			ID:   teaching.InstrumentID(instrumentRow.ID),
			Name: instrumentRow.Name,
		})
	}

	return instruments, nil
}

func (s teachingServiceImpl) InsertInstruments(ctx context.Context, specs []teaching.InsertInstrumentSpec) ([]teaching.InstrumentID, error) {
	instrumentIDs := make([]teaching.InstrumentID, 0, len(specs))

	for _, spec := range specs {
		instrumentID, err := s.mySQLQueries.InsertInstrument(ctx, spec.Name)
		if err != nil {
			return []teaching.InstrumentID{}, fmt.Errorf("qtx.InsertInstrument(): %w", err)
		}
		instrumentIDs = append(instrumentIDs, teaching.InstrumentID(instrumentID))
	}

	return instrumentIDs, nil
}

func (s teachingServiceImpl) UpdateInstruments(ctx context.Context, specs []teaching.UpdateInstrumentSpec) ([]teaching.InstrumentID, error) {
	instrumentIDs := make([]teaching.InstrumentID, 0, len(specs))

	for _, spec := range specs {
		err := s.mySQLQueries.UpdateInstrument(ctx, mysql.UpdateInstrumentParams{
			Name: spec.Name,
			ID:   int64(spec.ID),
		})
		if err != nil {
			return []teaching.InstrumentID{}, fmt.Errorf("qtx.UpdateInstrument(): %w", err)
		}
		instrumentIDs = append(instrumentIDs, spec.ID)
	}

	return instrumentIDs, nil
}

func (s teachingServiceImpl) DeleteInstruments(ctx context.Context, ids []teaching.InstrumentID) error {
	idsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt64 = append(idsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteInstrumentsByIds(ctx, idsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteInstrumentsByIds(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) GetGrades(ctx context.Context, pagination util.PaginationSpec) (teaching.GetGradesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	gradeRows, err := s.mySQLQueries.GetGrades(ctx, mysql.GetGradesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return teaching.GetGradesResult{}, fmt.Errorf("mySQLQueries.GetGrades(): %w", err)
	}

	grades := make([]teaching.Grade, 0, len(gradeRows))
	for _, gradeRow := range gradeRows {
		grades = append(grades, teaching.Grade{
			ID:   teaching.GradeID(gradeRow.ID),
			Name: gradeRow.Name,
		})
	}

	totalResults, err := s.mySQLQueries.CountGrades(ctx)

	return teaching.GetGradesResult{
		Grades:           grades,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetGradeById(ctx context.Context, id teaching.GradeID) (teaching.Grade, error) {
	grade, err := s.mySQLQueries.GetGradeById(ctx, int64(id))
	if err != nil {
		return teaching.Grade{}, fmt.Errorf("mySQLQueries.GetGradeById(): %w", err)
	}

	return teaching.Grade{
		ID:   teaching.GradeID(grade.ID),
		Name: grade.Name,
	}, nil
}

func (s teachingServiceImpl) GetGradesByIds(ctx context.Context, ids []teaching.GradeID) ([]teaching.Grade, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	gradeRows, err := s.mySQLQueries.GetGradesByIds(ctx, idsInt)
	if err != nil {
		return []teaching.Grade{}, fmt.Errorf("mySQLQueries.GetGradesByIds(): %w", err)
	}

	grades := make([]teaching.Grade, 0, len(gradeRows))
	for _, gradeRow := range gradeRows {
		grades = append(grades, teaching.Grade{
			ID:   teaching.GradeID(gradeRow.ID),
			Name: gradeRow.Name,
		})
	}

	return grades, nil
}

func (s teachingServiceImpl) InsertGrades(ctx context.Context, specs []teaching.InsertGradeSpec) ([]teaching.GradeID, error) {
	gradeIDs := make([]teaching.GradeID, 0, len(specs))

	for _, spec := range specs {
		gradeID, err := s.mySQLQueries.InsertGrade(ctx, spec.Name)
		if err != nil {
			return []teaching.GradeID{}, fmt.Errorf("qtx.InsertGrade(): %w", err)
		}
		gradeIDs = append(gradeIDs, teaching.GradeID(gradeID))
	}

	return gradeIDs, nil
}

func (s teachingServiceImpl) UpdateGrades(ctx context.Context, specs []teaching.UpdateGradeSpec) ([]teaching.GradeID, error) {
	gradeIDs := make([]teaching.GradeID, 0, len(specs))

	for _, spec := range specs {
		err := s.mySQLQueries.UpdateGrade(ctx, mysql.UpdateGradeParams{
			Name: spec.Name,
			ID:   int64(spec.ID),
		})
		if err != nil {
			return []teaching.GradeID{}, fmt.Errorf("qtx.UpdateGrade(): %w", err)
		}
		gradeIDs = append(gradeIDs, spec.ID)
	}

	return gradeIDs, nil
}

func (s teachingServiceImpl) DeleteGrades(ctx context.Context, ids []teaching.GradeID) error {
	idsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt64 = append(idsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteGradesByIds(ctx, idsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteGradeByIds(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) GetCourses(ctx context.Context, pagination util.PaginationSpec) (teaching.GetCoursesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	courseRows, err := s.mySQLQueries.GetCourses(ctx, mysql.GetCoursesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return teaching.GetCoursesResult{}, fmt.Errorf("mySQLQueries.GetCourses(): %w", err)
	}

	courses := make([]teaching.Course, 0, len(courseRows))
	for _, courseRow := range courseRows {
		courses = append(courses, teaching.Course{
			ID:                    teaching.CourseID(courseRow.CourseID),
			CompleteName:          courseRow.CourseName,
			DefaultFee:            courseRow.DefaultFee,
			DefaultDurationMinute: courseRow.DefaultDurationMinute,
		})
	}

	totalResults, err := s.mySQLQueries.CountCourses(ctx)

	return teaching.GetCoursesResult{
		Courses:          courses,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetCourseById(ctx context.Context, id teaching.CourseID) (teaching.Course, error) {
	course, err := s.mySQLQueries.GetCourseById(ctx, int64(id))
	if err != nil {
		return teaching.Course{}, fmt.Errorf("mySQLQueries.GetCourseById(): %w", err)
	}

	return teaching.Course{
		ID:                    teaching.CourseID(course.CourseID),
		CompleteName:          course.CourseName,
		DefaultFee:            course.DefaultFee,
		DefaultDurationMinute: course.DefaultDurationMinute,
	}, nil
}

func (s teachingServiceImpl) GetCoursesByIds(ctx context.Context, ids []teaching.CourseID) ([]teaching.Course, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	courseRows, err := s.mySQLQueries.GetCoursesByIds(ctx, idsInt)
	if err != nil {
		return []teaching.Course{}, fmt.Errorf("mySQLQueries.GetCoursesByIds(): %w", err)
	}

	courses := make([]teaching.Course, 0, len(courseRows))
	for _, courseRow := range courseRows {
		courses = append(courses, teaching.Course{
			ID:                    teaching.CourseID(courseRow.CourseID),
			CompleteName:          courseRow.CourseName,
			DefaultFee:            courseRow.DefaultFee,
			DefaultDurationMinute: courseRow.DefaultDurationMinute,
		})
	}

	return courses, nil
}

func (s teachingServiceImpl) InsertCourses(ctx context.Context, specs []teaching.InsertCourseSpec) ([]teaching.CourseID, error) {
	courseIDs := make([]teaching.CourseID, 0, len(specs))

	for _, spec := range specs {
		courseID, err := s.mySQLQueries.InsertCourse(ctx, mysql.InsertCourseParams{
			DefaultFee:            spec.DefaultFee,
			DefaultDurationMinute: spec.DefaultDurationMinute,
			InstrumentID:          int64(spec.InstrumentID),
			GradeID:               int64(spec.GradeID),
		})
		if err != nil {
			return []teaching.CourseID{}, fmt.Errorf("qtx.InsertCourse(): %w", err)
		}
		courseIDs = append(courseIDs, teaching.CourseID(courseID))
	}

	return courseIDs, nil
}

func (s teachingServiceImpl) UpdateCourses(ctx context.Context, specs []teaching.UpdateCourseSpec) ([]teaching.CourseID, error) {
	courseIDs := make([]teaching.CourseID, 0, len(specs))

	for _, spec := range specs {
		err := s.mySQLQueries.UpdateCourseInfo(ctx, mysql.UpdateCourseInfoParams{
			DefaultFee:            spec.DefaultFee,
			DefaultDurationMinute: spec.DefaultDurationMinute,
			ID:                    int64(spec.ID),
		})
		if err != nil {
			return []teaching.CourseID{}, fmt.Errorf("qtx.UpdateCourse(): %w", err)
		}
		courseIDs = append(courseIDs, spec.ID)
	}

	return courseIDs, nil
}

func (s teachingServiceImpl) DeleteCourses(ctx context.Context, ids []teaching.CourseID) error {
	idsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt64 = append(idsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteCoursesByIds(ctx, idsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteCourseByIds(): %w", err)
	}

	return nil
}

func (s teachingServiceImpl) GetClasses(ctx context.Context, pagination util.PaginationSpec) (teaching.GetClassesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	classRows, err := s.mySQLQueries.GetClasses(ctx, mysql.GetClassesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return teaching.GetClassesResult{}, fmt.Errorf("mySQLQueries.GetClasses(): %w", err)
	}

	classes := NewClassesFromGetClassesRow(classRows)
	totalResults, err := s.mySQLQueries.CountClasses(ctx)

	return teaching.GetClassesResult{
		Classes:          classes,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetClassById(ctx context.Context, id teaching.ClassID) (teaching.Class, error) {
	classRows, err := s.mySQLQueries.GetClassById(ctx, int64(id))
	if err != nil {
		return teaching.Class{}, fmt.Errorf("mySQLQueries.GetClassById(): %w", err)
	}

	if len(classRows) == 0 {
		return teaching.Class{}, sql.ErrNoRows
	}

	classRowsConverted := make([]mysql.GetClassesRow, 0, len(classRows))
	for _, row := range classRows {
		classRowsConverted = append(classRowsConverted, row.ToGetClassesRow())
	}

	class := NewClassesFromGetClassesRow(classRowsConverted)[0]

	return class, nil
}

func (s teachingServiceImpl) GetClassesByIds(ctx context.Context, ids []teaching.ClassID) ([]teaching.Class, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	classRows, err := s.mySQLQueries.GetClassesByIds(ctx, idsInt)
	if err != nil {
		return []teaching.Class{}, fmt.Errorf("mySQLQueries.GetClassesByIds(): %w", err)
	}

	classRowsConverted := make([]mysql.GetClassesRow, 0, len(classRows))
	for _, row := range classRows {
		classRowsConverted = append(classRowsConverted, row.ToGetClassesRow())
	}

	classes := NewClassesFromGetClassesRow(classRowsConverted)

	return classes, nil
}

func (s teachingServiceImpl) InsertClasses(ctx context.Context, specs []teaching.InsertClassSpec) ([]teaching.ClassID, error) {
	classIDs := make([]teaching.ClassID, 0, len(specs))

	tx, err := s.mySQLQueries.DB.BeginTx(ctx, nil)
	if err != nil {
		return []teaching.ClassID{}, fmt.Errorf("mySQLQueries.DB.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTx(tx)

	for _, spec := range specs {
		classID, err := qtx.InsertClass(ctx, mysql.InsertClassParams{
			TransportFee: spec.TransportFee,
			TeacherID:    sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != teaching.TeacherID_None},
			CourseID:     int64(spec.CourseID),
		})
		if err != nil {
			wrappedErr := errs.WrapMySQLError(err)
			return []teaching.ClassID{}, fmt.Errorf("qtx.InsertClass(): %w", wrappedErr)
		}
		classIDs = append(classIDs, teaching.ClassID(classID))

		for _, studentId := range spec.StudentIDs {
			err := qtx.InsertStudentEnrollment(ctx, mysql.InsertStudentEnrollmentParams{
				StudentID: int64(studentId),
				ClassID:   classID,
			})
			if err != nil {
				wrappedErr := errs.WrapMySQLError(err)
				return []teaching.ClassID{}, fmt.Errorf("qtx.InsertStudentEnrollment(): %w", wrappedErr)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return []teaching.ClassID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return classIDs, nil
}

func (s teachingServiceImpl) UpdateClasses(ctx context.Context, specs []teaching.UpdateClassSpec) ([]teaching.ClassID, error) {
	classIDs := make([]teaching.ClassID, 0, len(specs))

	tx, err := s.mySQLQueries.DB.BeginTx(ctx, nil)
	if err != nil {
		return []teaching.ClassID{}, fmt.Errorf("mySQLQueries.DB.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTx(tx)

	for _, spec := range specs {
		// Updated class
		err := s.mySQLQueries.UpdateClass(ctx, mysql.UpdateClassParams{
			TransportFee:  spec.TransportFee,
			TeacherID:     sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != teaching.TeacherID_None},
			IsDeactivated: util.BoolToInt32(spec.IsDeactivated),
			ID:            int64(spec.ID),
		})
		if err != nil {
			return []teaching.ClassID{}, fmt.Errorf("qtx.UpdateClass(): %w", err)
		}
		classIDs = append(classIDs, spec.ID)

		// Added students
		for _, studentId := range spec.AddedStudentIDs {
			err = qtx.InsertStudentEnrollment(ctx, mysql.InsertStudentEnrollmentParams{
				StudentID: int64(studentId),
				ClassID:   int64(spec.ID),
			})
			if err != nil {
				wrappedErr := errs.WrapMySQLError(err)
				return []teaching.ClassID{}, fmt.Errorf("qtx.InsertStudentEnrollment(): %w", wrappedErr)
			}
		}

		// Removed enrollments
		idsInt := make([]int64, 0, len(spec.DeletedEnrollmentIDs))
		for _, deletedEnrollmentID := range spec.DeletedEnrollmentIDs {
			idsInt = append(idsInt, int64(deletedEnrollmentID))
		}
		err = qtx.DeleteStudentEnrollmentsByIds(ctx, idsInt)
		if err != nil {
			wrappedErr := errs.WrapMySQLError(err)
			return []teaching.ClassID{}, fmt.Errorf("qtx.DeleteStudentEnrollmentsByIds(): %w", wrappedErr)
		}
	}

	err = tx.Commit()
	if err != nil {
		return []teaching.ClassID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return classIDs, nil
}

func (s teachingServiceImpl) DeleteClasses(ctx context.Context, ids []teaching.ClassID) error {
	idsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt64 = append(idsInt64, int64(id))
	}

	tx, err := s.mySQLQueries.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DB.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTx(tx)

	err = qtx.DeleteStudentEnrollmentByClassIds(ctx, idsInt64)
	if err != nil {
		wrappedErr := errs.WrapMySQLError(err)
		return fmt.Errorf("qtx.DeleteStudentEnrollmentByClassIds(): %w", wrappedErr)
	}

	err = qtx.DeleteClassesByIds(ctx, idsInt64)
	if err != nil {
		wrappedErr := errs.WrapMySQLError(err)
		return fmt.Errorf("qtx.DeleteClassesByIds(): %w", wrappedErr)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx.Commit(): %w", err)
	}

	return nil
}
