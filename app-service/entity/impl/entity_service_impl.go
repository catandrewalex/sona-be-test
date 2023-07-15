package impl

import (
	"context"
	"database/sql"
	"fmt"

	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/entity"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
	"sonamusica-backend/config"
	"sonamusica-backend/logging"
	"sonamusica-backend/network"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("EntityService", logging.GetLevel(configObject.LogLevel))
)

type entityServiceImpl struct {
	mySQLQueries *relational_db.MySQLQueries

	identityService identity.IdentityService
}

var _ entity.EntityService = (*entityServiceImpl)(nil)

func NewEntityServiceImpl(mySQLQueries *relational_db.MySQLQueries, identityService identity.IdentityService) *entityServiceImpl {
	return &entityServiceImpl{
		mySQLQueries:    mySQLQueries,
		identityService: identityService,
	}
}

func (s entityServiceImpl) GetTeachers(ctx context.Context, pagination util.PaginationSpec) (entity.GetTeachersResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	teacherRows, err := s.mySQLQueries.GetTeachers(ctx, mysql.GetTeachersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetTeachersResult{}, fmt.Errorf("mySQLQueries.GetTeachers(): %w", err)
	}

	teachers := NewTeachersFromGetTeachersRow(teacherRows)

	totalResults, err := s.mySQLQueries.CountTeachers(ctx)
	if err != nil {
		return entity.GetTeachersResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetTeachersResult{
		Teachers:         teachers,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetTeacherById(ctx context.Context, id entity.TeacherID) (entity.Teacher, error) {
	teacherRow, err := s.mySQLQueries.GetTeacherById(ctx, int64(id))
	if err != nil {
		return entity.Teacher{}, fmt.Errorf("mySQLQueries.GetTeacherById(): %w", err)
	}

	teacher := NewTeachersFromGetTeachersRow([]mysql.GetTeachersRow{teacherRow.ToGetTeachersRow()})[0]

	return teacher, nil
}

func (s entityServiceImpl) GetTeachersByIds(ctx context.Context, ids []entity.TeacherID) ([]entity.Teacher, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	teacherRows, err := s.mySQLQueries.GetTeachersByIds(ctx, idsInt)
	if err != nil {
		return []entity.Teacher{}, fmt.Errorf("mySQLQueries.GetTeachersByIds(): %w", err)
	}

	teacherRowsConverted := make([]mysql.GetTeachersRow, 0, len(teacherRows))
	for _, teacherRow := range teacherRows {
		teacherRowsConverted = append(teacherRowsConverted, teacherRow.ToGetTeachersRow())
	}

	teachers := NewTeachersFromGetTeachersRow(teacherRowsConverted)

	return teachers, nil
}

func (s entityServiceImpl) InsertTeachers(ctx context.Context, userIDs []identity.UserID) ([]entity.TeacherID, error) {
	teacherIDs := make([]entity.TeacherID, 0, len(userIDs))

	for _, userID := range userIDs {
		teacherID, err := s.mySQLQueries.InsertTeacher(ctx, int64(userID))
		if err != nil {
			return []entity.TeacherID{}, fmt.Errorf("qtx.InsertTeacher(): %w", err)
		}
		teacherIDs = append(teacherIDs, entity.TeacherID(teacherID))
	}

	return teacherIDs, nil
}

func (s entityServiceImpl) InsertTeachersWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]entity.TeacherID, error) {
	teacherIDs := make([]entity.TeacherID, 0, len(specs))

	// TODO: move all mySQLQueries.* (Begin, Commit, etc.) to a new accessor service in lower level
	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.TeacherID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	ctxWithSQLTx := network.NewContextWithSQLTx(ctx, tx)
	userIDs, err := s.identityService.InsertUsers(ctxWithSQLTx, specs)
	if err != nil {
		return []entity.TeacherID{}, fmt.Errorf("identityService.InsertUsers(): %w", err)
	}

	for _, userID := range userIDs {
		teacherID, err := qtx.InsertTeacher(ctx, int64(userID))
		if err != nil {
			return []entity.TeacherID{}, fmt.Errorf("qtx.InsertTeacher(): %w", err)
		}
		teacherIDs = append(teacherIDs, entity.TeacherID(teacherID))
	}

	err = tx.Commit()
	if err != nil {
		return []entity.TeacherID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return teacherIDs, nil
}

func (s entityServiceImpl) DeleteTeachers(ctx context.Context, ids []entity.TeacherID) error {
	teacherIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		teacherIdsInt64 = append(teacherIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteTeachersByIds(ctx, teacherIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteTeacherByIds(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetStudents(ctx context.Context, pagination util.PaginationSpec) (entity.GetStudentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	studentRows, err := s.mySQLQueries.GetStudents(ctx, mysql.GetStudentsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetStudentsResult{}, fmt.Errorf("mySQLQueries.GetStudents(): %w", err)
	}

	students := NewStudentsFromGetStudentsRow(studentRows)

	totalResults, err := s.mySQLQueries.CountStudents(ctx)
	if err != nil {
		return entity.GetStudentsResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetStudentsResult{
		Students:         students,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetStudentById(ctx context.Context, id entity.StudentID) (entity.Student, error) {
	studentRow, err := s.mySQLQueries.GetStudentById(ctx, int64(id))
	if err != nil {
		return entity.Student{}, fmt.Errorf("mySQLQueries.GetStudentById(): %w", err)
	}

	student := NewStudentsFromGetStudentsRow([]mysql.GetStudentsRow{studentRow.ToGetStudentsRow()})[0]

	return student, nil
}

func (s entityServiceImpl) GetStudentsByIds(ctx context.Context, ids []entity.StudentID) ([]entity.Student, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	studentRows, err := s.mySQLQueries.GetStudentsByIds(ctx, idsInt)
	if err != nil {
		return []entity.Student{}, fmt.Errorf("mySQLQueries.GetStudentsByIds(): %w", err)
	}

	studentRowsConverted := make([]mysql.GetStudentsRow, 0, len(studentRows))
	for _, studentRow := range studentRows {
		studentRowsConverted = append(studentRowsConverted, studentRow.ToGetStudentsRow())
	}

	students := NewStudentsFromGetStudentsRow(studentRowsConverted)

	return students, nil
}

func (s entityServiceImpl) InsertStudents(ctx context.Context, userIDs []identity.UserID) ([]entity.StudentID, error) {
	studentIDs := make([]entity.StudentID, 0, len(userIDs))

	for _, userID := range userIDs {
		studentID, err := s.mySQLQueries.InsertStudent(ctx, int64(userID))
		if err != nil {
			return []entity.StudentID{}, fmt.Errorf("qtx.InsertStudent(): %w", err)
		}
		studentIDs = append(studentIDs, entity.StudentID(studentID))
	}

	return studentIDs, nil
}

func (s entityServiceImpl) InsertStudentsWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]entity.StudentID, error) {
	studentIDs := make([]entity.StudentID, 0, len(specs))

	// TODO: move all mySQLQueries.* (Begin, Commit, etc.) to a new accessor service in lower level
	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.StudentID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	ctxWithSQLTx := network.NewContextWithSQLTx(ctx, tx)
	userIDs, err := s.identityService.InsertUsers(ctxWithSQLTx, specs)
	if err != nil {
		return []entity.StudentID{}, fmt.Errorf("identityService.InsertUsers(): %w", err)
	}

	for _, userID := range userIDs {
		studentID, err := qtx.InsertStudent(ctx, int64(userID))
		if err != nil {
			return []entity.StudentID{}, fmt.Errorf("qtx.InsertStudent(): %w", err)
		}
		studentIDs = append(studentIDs, entity.StudentID(studentID))
	}

	err = tx.Commit()
	if err != nil {
		return []entity.StudentID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return studentIDs, nil
}

func (s entityServiceImpl) DeleteStudents(ctx context.Context, ids []entity.StudentID) error {
	studentIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		studentIdsInt64 = append(studentIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteStudentsByIds(ctx, studentIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteStudentByIds(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetInstruments(ctx context.Context, pagination util.PaginationSpec) (entity.GetInstrumentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	instrumentRows, err := s.mySQLQueries.GetInstruments(ctx, mysql.GetInstrumentsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetInstrumentsResult{}, fmt.Errorf("mySQLQueries.GetInstruments(): %w", err)
	}

	instruments := NewInstrumentsFromMySQLInstruments(instrumentRows)

	totalResults, err := s.mySQLQueries.CountInstruments(ctx)
	if err != nil {
		return entity.GetInstrumentsResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetInstrumentsResult{
		Instruments:      instruments,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetInstrumentById(ctx context.Context, id entity.InstrumentID) (entity.Instrument, error) {
	instrumentRow, err := s.mySQLQueries.GetInstrumentById(ctx, int64(id))
	if err != nil {
		return entity.Instrument{}, fmt.Errorf("mySQLQueries.GetInstrumentById(): %w", err)
	}

	instrument := NewInstrumentsFromMySQLInstruments([]mysql.Instrument{instrumentRow})[0]

	return instrument, nil
}

func (s entityServiceImpl) GetInstrumentsByIds(ctx context.Context, ids []entity.InstrumentID) ([]entity.Instrument, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	instrumentRows, err := s.mySQLQueries.GetInstrumentsByIds(ctx, idsInt)
	if err != nil {
		return []entity.Instrument{}, fmt.Errorf("mySQLQueries.GetInstrumentsByIds(): %w", err)
	}

	instruments := NewInstrumentsFromMySQLInstruments(instrumentRows)

	return instruments, nil
}

func (s entityServiceImpl) InsertInstruments(ctx context.Context, specs []entity.InsertInstrumentSpec) ([]entity.InstrumentID, error) {
	instrumentIDs := make([]entity.InstrumentID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.InstrumentID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		instrumentID, err := qtx.InsertInstrument(ctx, spec.Name)
		if err != nil {

			return []entity.InstrumentID{}, fmt.Errorf("qtx.InsertInstrument(): %w", err)
		}
		instrumentIDs = append(instrumentIDs, entity.InstrumentID(instrumentID))
	}

	err = tx.Commit()
	if err != nil {
		return []entity.InstrumentID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return instrumentIDs, nil
}

func (s entityServiceImpl) UpdateInstruments(ctx context.Context, specs []entity.UpdateInstrumentSpec) ([]entity.InstrumentID, error) {
	instrumentIDs := make([]entity.InstrumentID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.InstrumentID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		err := qtx.UpdateInstrument(ctx, mysql.UpdateInstrumentParams{
			Name: spec.Name,
			ID:   int64(spec.InstrumentID),
		})
		if err != nil {

			return []entity.InstrumentID{}, fmt.Errorf("qtx.UpdateInstrument(): %w", err)
		}
		instrumentIDs = append(instrumentIDs, spec.InstrumentID)
	}

	err = tx.Commit()
	if err != nil {
		return []entity.InstrumentID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return instrumentIDs, nil
}

func (s entityServiceImpl) DeleteInstruments(ctx context.Context, ids []entity.InstrumentID) error {
	instrumentIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		instrumentIdsInt64 = append(instrumentIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteInstrumentsByIds(ctx, instrumentIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteInstrumentsByIds(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetGrades(ctx context.Context, pagination util.PaginationSpec) (entity.GetGradesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	gradeRows, err := s.mySQLQueries.GetGrades(ctx, mysql.GetGradesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetGradesResult{}, fmt.Errorf("mySQLQueries.GetGrades(): %w", err)
	}

	grades := NewGradesFromMySQLGrades(gradeRows)

	totalResults, err := s.mySQLQueries.CountGrades(ctx)
	if err != nil {
		return entity.GetGradesResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetGradesResult{
		Grades:           grades,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetGradeById(ctx context.Context, id entity.GradeID) (entity.Grade, error) {
	gradeRow, err := s.mySQLQueries.GetGradeById(ctx, int64(id))
	if err != nil {
		return entity.Grade{}, fmt.Errorf("mySQLQueries.GetGradeById(): %w", err)
	}

	grade := NewGradesFromMySQLGrades([]mysql.Grade{gradeRow})[0]

	return grade, nil
}

func (s entityServiceImpl) GetGradesByIds(ctx context.Context, ids []entity.GradeID) ([]entity.Grade, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	gradeRows, err := s.mySQLQueries.GetGradesByIds(ctx, idsInt)
	if err != nil {
		return []entity.Grade{}, fmt.Errorf("mySQLQueries.GetGradesByIds(): %w", err)
	}

	grades := NewGradesFromMySQLGrades(gradeRows)

	return grades, nil
}

func (s entityServiceImpl) InsertGrades(ctx context.Context, specs []entity.InsertGradeSpec) ([]entity.GradeID, error) {
	gradeIDs := make([]entity.GradeID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.GradeID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		gradeID, err := qtx.InsertGrade(ctx, spec.Name)
		if err != nil {

			return []entity.GradeID{}, fmt.Errorf("qtx.InsertGrade(): %w", err)
		}
		gradeIDs = append(gradeIDs, entity.GradeID(gradeID))
	}

	err = tx.Commit()
	if err != nil {
		return []entity.GradeID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return gradeIDs, nil
}

func (s entityServiceImpl) UpdateGrades(ctx context.Context, specs []entity.UpdateGradeSpec) ([]entity.GradeID, error) {
	gradeIDs := make([]entity.GradeID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.GradeID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		err := qtx.UpdateGrade(ctx, mysql.UpdateGradeParams{
			Name: spec.Name,
			ID:   int64(spec.GradeID),
		})
		if err != nil {

			return []entity.GradeID{}, fmt.Errorf("qtx.UpdateGrade(): %w", err)
		}
		gradeIDs = append(gradeIDs, spec.GradeID)
	}

	err = tx.Commit()
	if err != nil {
		return []entity.GradeID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return gradeIDs, nil
}

func (s entityServiceImpl) DeleteGrades(ctx context.Context, ids []entity.GradeID) error {
	gradeIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		gradeIdsInt64 = append(gradeIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteGradesByIds(ctx, gradeIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteGradeByIds(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetCourses(ctx context.Context, pagination util.PaginationSpec) (entity.GetCoursesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	courseRows, err := s.mySQLQueries.GetCourses(ctx, mysql.GetCoursesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetCoursesResult{}, fmt.Errorf("mySQLQueries.GetCourses(): %w", err)
	}

	courses := NewCoursesFromGetCoursesRow(courseRows)

	totalResults, err := s.mySQLQueries.CountCourses(ctx)
	if err != nil {
		return entity.GetCoursesResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetCoursesResult{
		Courses:          courses,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetCourseById(ctx context.Context, id entity.CourseID) (entity.Course, error) {
	courseRow, err := s.mySQLQueries.GetCourseById(ctx, int64(id))
	if err != nil {
		return entity.Course{}, fmt.Errorf("mySQLQueries.GetCourseById(): %w", err)
	}

	course := NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{courseRow.ToGetCoursesRow()})[0]

	return course, nil
}

func (s entityServiceImpl) GetCoursesByIds(ctx context.Context, ids []entity.CourseID) ([]entity.Course, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	courseRows, err := s.mySQLQueries.GetCoursesByIds(ctx, idsInt)
	if err != nil {
		return []entity.Course{}, fmt.Errorf("mySQLQueries.GetCoursesByIds(): %w", err)
	}

	courseRowsConverted := make([]mysql.GetCoursesRow, 0, len(courseRows))
	for _, row := range courseRows {
		courseRowsConverted = append(courseRowsConverted, row.ToGetCoursesRow())
	}

	courses := NewCoursesFromGetCoursesRow(courseRowsConverted)

	return courses, nil
}

func (s entityServiceImpl) InsertCourses(ctx context.Context, specs []entity.InsertCourseSpec) ([]entity.CourseID, error) {
	courseIDs := make([]entity.CourseID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.CourseID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		courseID, err := qtx.InsertCourse(ctx, mysql.InsertCourseParams{
			DefaultFee:            spec.DefaultFee,
			DefaultDurationMinute: spec.DefaultDurationMinute,
			InstrumentID:          int64(spec.InstrumentID),
			GradeID:               int64(spec.GradeID),
		})
		if err != nil {

			return []entity.CourseID{}, fmt.Errorf("qtx.InsertCourse(): %w", err)
		}
		courseIDs = append(courseIDs, entity.CourseID(courseID))
	}

	err = tx.Commit()
	if err != nil {
		return []entity.CourseID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return courseIDs, nil
}

func (s entityServiceImpl) UpdateCourses(ctx context.Context, specs []entity.UpdateCourseSpec) ([]entity.CourseID, error) {
	courseIDs := make([]entity.CourseID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.CourseID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		err := qtx.UpdateCourseInfo(ctx, mysql.UpdateCourseInfoParams{
			DefaultFee:            spec.DefaultFee,
			DefaultDurationMinute: spec.DefaultDurationMinute,
			ID:                    int64(spec.CourseID),
		})
		if err != nil {

			return []entity.CourseID{}, fmt.Errorf("qtx.UpdateCourseInfo(): %w", err)
		}
		courseIDs = append(courseIDs, spec.CourseID)
	}

	err = tx.Commit()
	if err != nil {
		return []entity.CourseID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return courseIDs, nil
}

func (s entityServiceImpl) DeleteCourses(ctx context.Context, ids []entity.CourseID) error {
	courseIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		courseIdsInt64 = append(courseIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteCoursesByIds(ctx, courseIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteCourseByIds(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetClasses(ctx context.Context, pagination util.PaginationSpec, includeDeactivated bool) (entity.GetClassesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	isDeactivatedFilters := []int32{0}
	if includeDeactivated {
		isDeactivatedFilters = append(isDeactivatedFilters, 1)
	}

	classRows, err := s.mySQLQueries.GetClasses(ctx, mysql.GetClassesParams{
		IsDeactivateds: isDeactivatedFilters,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		return entity.GetClassesResult{}, fmt.Errorf("mySQLQueries.GetClasses(): %w", err)
	}

	classes := NewClassesFromGetClassesRow(classRows)

	totalResults, err := s.mySQLQueries.CountClasses(ctx, isDeactivatedFilters)
	if err != nil {
		return entity.GetClassesResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetClassesResult{
		Classes:          classes,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetClassById(ctx context.Context, id entity.ClassID) (entity.Class, error) {
	classRows, err := s.mySQLQueries.GetClassById(ctx, int64(id))
	if err != nil {
		return entity.Class{}, fmt.Errorf("mySQLQueries.GetClassById(): %w", err)
	}

	if len(classRows) == 0 {
		return entity.Class{}, sql.ErrNoRows
	}

	classRowsConverted := make([]mysql.GetClassesRow, 0, len(classRows))
	for _, row := range classRows {
		classRowsConverted = append(classRowsConverted, row.ToGetClassesRow())
	}

	class := NewClassesFromGetClassesRow(classRowsConverted)[0]

	return class, nil
}

func (s entityServiceImpl) GetClassesByIds(ctx context.Context, ids []entity.ClassID) ([]entity.Class, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	classRows, err := s.mySQLQueries.GetClassesByIds(ctx, idsInt)
	if err != nil {
		return []entity.Class{}, fmt.Errorf("mySQLQueries.GetClassesByIds(): %w", err)
	}

	classRowsConverted := make([]mysql.GetClassesRow, 0, len(classRows))
	for _, row := range classRows {
		classRowsConverted = append(classRowsConverted, row.ToGetClassesRow())
	}

	classes := NewClassesFromGetClassesRow(classRowsConverted)

	return classes, nil
}

func (s entityServiceImpl) InsertClasses(ctx context.Context, specs []entity.InsertClassSpec) ([]entity.ClassID, error) {
	classIDs := make([]entity.ClassID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.ClassID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		classID, err := qtx.InsertClass(ctx, mysql.InsertClassParams{
			TransportFee: spec.TransportFee,
			TeacherID:    sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != entity.TeacherID_None},
			CourseID:     int64(spec.CourseID),
		})
		if err != nil {

			return []entity.ClassID{}, fmt.Errorf("qtx.InsertClass(): %w", err)
		}
		classIDs = append(classIDs, entity.ClassID(classID))

		for _, studentId := range spec.StudentIDs {
			err := qtx.InsertStudentEnrollment(ctx, mysql.InsertStudentEnrollmentParams{
				StudentID: int64(studentId),
				ClassID:   classID,
			})
			if err != nil {

				return []entity.ClassID{}, fmt.Errorf("qtx.InsertStudentEnrollment(): %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return []entity.ClassID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return classIDs, nil
}

func (s entityServiceImpl) UpdateClasses(ctx context.Context, specs []entity.UpdateClassSpec) ([]entity.ClassID, error) {
	classIDs := make([]entity.ClassID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.ClassID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		classId := int64(spec.ClassID)
		// Updated class
		err := s.mySQLQueries.UpdateClass(ctx, mysql.UpdateClassParams{
			TransportFee:  spec.TransportFee,
			TeacherID:     sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != entity.TeacherID_None},
			IsDeactivated: util.BoolToInt32(spec.IsDeactivated),
			ID:            classId,
		})
		if err != nil {
			return []entity.ClassID{}, fmt.Errorf("qtx.UpdateClass(): %w", err)
		}
		classIDs = append(classIDs, spec.ClassID)

		studentDifference, err := calculateClassStudentsDifference(ctx, qtx, spec.ClassID, spec.StudentIDs)
		if err != nil {
			return []entity.ClassID{}, fmt.Errorf("calculateClassStudentsDifference(): %w", err)
		}

		// Added students
		for _, studentId := range studentDifference.addedStudentIDs {
			err = qtx.InsertStudentEnrollment(ctx, mysql.InsertStudentEnrollmentParams{
				StudentID: int64(studentId),
				ClassID:   classId,
			})
			if err != nil {

				return []entity.ClassID{}, fmt.Errorf("qtx.InsertStudentEnrollment(): %w", err)
			}
		}

		// Updated (re-enabled) enrollments
		for _, updatedEnrollmentID := range studentDifference.enabledStudentEnrollmentIDs {
			err = qtx.EnableStudentEnrollment(ctx, int64(updatedEnrollmentID))
			if err != nil {

				return []entity.ClassID{}, fmt.Errorf("qtx.EnableStudentEnrollment(): %w", err)
			}
		}

		// Delete or disable enrollments
		for _, disabledEnrollmentID := range studentDifference.disabledStudentEnrollmentIDs {
			err = qtx.DeleteStudentEnrollmentById(ctx, int64(disabledEnrollmentID))
			if err == nil {
				// the enrollment is still deletable (not referenced by any other entity), then we straightforwardly delete it
				continue
			}

			err = qtx.DisableStudentEnrollment(ctx, int64(disabledEnrollmentID))
			if err != nil {
				return []entity.ClassID{}, fmt.Errorf("qtx.DisableStudentEnrollment(): %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return []entity.ClassID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return classIDs, nil
}

type classStudentsDifference struct {
	addedStudentIDs              []entity.StudentID
	enabledStudentEnrollmentIDs  []entity.StudentEnrollmentID
	disabledStudentEnrollmentIDs []entity.StudentEnrollmentID
}

func calculateClassStudentsDifference(ctx context.Context, qtx *mysql.Queries, classId entity.ClassID, finalStudentIDs []entity.StudentID) (classStudentsDifference, error) {
	addedStudentIDs := make([]entity.StudentID, 0)
	enabledStudentEnrollmentIDs := make([]entity.StudentEnrollmentID, 0)
	disabledStudentEnrollmentIDs := make([]entity.StudentEnrollmentID, 0)

	enrollments, err := qtx.GetStudentEnrollmentsByClassId(ctx, int64(classId))
	if err != nil {
		return classStudentsDifference{}, fmt.Errorf("qtx.GetStudentEnrollmentsByClassId(): %w", err)
	}

	studentIDToEnrollmentIDMap := make(map[entity.StudentID]entity.StudentEnrollmentID, len(enrollments))
	for _, enrollment := range enrollments {
		studentIDToEnrollmentIDMap[entity.StudentID(enrollment.StudentID)] = entity.StudentEnrollmentID(enrollment.ID)
	}

	finalStudentIDsMap := make(map[entity.StudentID]bool, 0)
	for _, studentID := range finalStudentIDs {
		if enrollmentID, ok := studentIDToEnrollmentIDMap[studentID]; ok {
			enabledStudentEnrollmentIDs = append(enabledStudentEnrollmentIDs, enrollmentID)
		} else {
			addedStudentIDs = append(addedStudentIDs, studentID)
		}
		finalStudentIDsMap[studentID] = true
	}
	for _, enrollment := range enrollments {
		if _, ok := finalStudentIDsMap[entity.StudentID(enrollment.StudentID)]; !ok {
			disabledStudentEnrollmentIDs = append(disabledStudentEnrollmentIDs, entity.StudentEnrollmentID(enrollment.ID))
		}
	}

	mainLog.Debug("calculateClassStudentsDifference():")
	mainLog.Debug("finalStudentIDs: %v", finalStudentIDs)
	mainLog.Debug("enrollments: %+v", enrollments)
	mainLog.Debug("classStudentsDifference: \n\taddedStudentIDs: %v\n\tenabledStudentEnrollmentIDs: %v\n\tdisabledStudentEnrollmentIDs: %v", addedStudentIDs, enabledStudentEnrollmentIDs, disabledStudentEnrollmentIDs)

	return classStudentsDifference{
		addedStudentIDs:              addedStudentIDs,
		enabledStudentEnrollmentIDs:  enabledStudentEnrollmentIDs,
		disabledStudentEnrollmentIDs: disabledStudentEnrollmentIDs,
	}, nil
}

func (s entityServiceImpl) DeleteClasses(ctx context.Context, ids []entity.ClassID) error {
	classIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		classIdsInt64 = append(classIdsInt64, int64(id))
	}

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	err = qtx.DeleteStudentEnrollmentByClassIds(ctx, classIdsInt64)
	if err != nil {
		return fmt.Errorf("qtx.DeleteStudentEnrollmentByClassIds(): %w", err)
	}

	err = qtx.DeleteClassesByIds(ctx, classIdsInt64)
	if err != nil {
		return fmt.Errorf("qtx.DeleteClassesByIds(): %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx.Commit(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetTeacherSpecialFees(ctx context.Context, pagination util.PaginationSpec) (entity.GetTeacherSpecialFeesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	teacherSpecialFeeRows, err := s.mySQLQueries.GetTeacherSpecialFees(ctx, mysql.GetTeacherSpecialFeesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetTeacherSpecialFeesResult{}, fmt.Errorf("mySQLQueries.GetTeacherSpecialFees(): %w", err)
	}

	teacherSpecialFees := NewTeacherSpecialFeesFromGetTeacherSpecialFeesRow(teacherSpecialFeeRows)

	totalResults, err := s.mySQLQueries.CountTeacherSpecialFees(ctx)
	if err != nil {
		return entity.GetTeacherSpecialFeesResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetTeacherSpecialFeesResult{
		TeacherSpecialFees: teacherSpecialFees,
		PaginationResult:   *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetTeacherSpecialFeeById(ctx context.Context, id entity.TeacherSpecialFeeID) (entity.TeacherSpecialFee, error) {
	teacherSpecialFeeRow, err := s.mySQLQueries.GetTeacherSpecialFeeById(ctx, int64(id))
	if err != nil {
		return entity.TeacherSpecialFee{}, fmt.Errorf("mySQLQueries.GetTeacherSpecialFeeById(): %w", err)
	}

	teacherSpecialFee := NewTeacherSpecialFeesFromGetTeacherSpecialFeesRow([]mysql.GetTeacherSpecialFeesRow{teacherSpecialFeeRow.ToGetTeacherSpecialFeesRow()})[0]

	return teacherSpecialFee, nil
}

func (s entityServiceImpl) GetTeacherSpecialFeesByIds(ctx context.Context, ids []entity.TeacherSpecialFeeID) ([]entity.TeacherSpecialFee, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	teacherSpecialFeeRows, err := s.mySQLQueries.GetTeacherSpecialFeesByIds(ctx, idsInt)
	if err != nil {
		return []entity.TeacherSpecialFee{}, fmt.Errorf("mySQLQueries.GetTeacherSpecialFeesByIds(): %w", err)
	}

	teacherSpecialFeeRowsConverted := make([]mysql.GetTeacherSpecialFeesRow, 0, len(teacherSpecialFeeRows))
	for _, row := range teacherSpecialFeeRows {
		teacherSpecialFeeRowsConverted = append(teacherSpecialFeeRowsConverted, row.ToGetTeacherSpecialFeesRow())
	}

	teacherSpecialFees := NewTeacherSpecialFeesFromGetTeacherSpecialFeesRow(teacherSpecialFeeRowsConverted)

	return teacherSpecialFees, nil
}

func (s entityServiceImpl) InsertTeacherSpecialFees(ctx context.Context, specs []entity.InsertTeacherSpecialFeeSpec) ([]entity.TeacherSpecialFeeID, error) {
	teacherSpecialFeeIDs := make([]entity.TeacherSpecialFeeID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.TeacherSpecialFeeID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		teacherSpecialFeeID, err := qtx.InsertTeacherSpecialFee(ctx, mysql.InsertTeacherSpecialFeeParams{
			Fee:       spec.Fee,
			TeacherID: int64(spec.TeacherID),
			CourseID:  int64(spec.CourseID),
		})
		if err != nil {

			return []entity.TeacherSpecialFeeID{}, fmt.Errorf("qtx.InsertTeacherSpecialFee(): %w", err)
		}
		teacherSpecialFeeIDs = append(teacherSpecialFeeIDs, entity.TeacherSpecialFeeID(teacherSpecialFeeID))
	}

	err = tx.Commit()
	if err != nil {
		return []entity.TeacherSpecialFeeID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return teacherSpecialFeeIDs, nil
}

func (s entityServiceImpl) UpdateTeacherSpecialFees(ctx context.Context, specs []entity.UpdateTeacherSpecialFeeSpec) ([]entity.TeacherSpecialFeeID, error) {
	teacherSpecialFeeIDs := make([]entity.TeacherSpecialFeeID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.TeacherSpecialFeeID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		err := qtx.UpdateTeacherSpecialFee(ctx, mysql.UpdateTeacherSpecialFeeParams{
			Fee: spec.Fee,
			ID:  int64(spec.TeacherSpecialFeeID),
		})
		if err != nil {

			return []entity.TeacherSpecialFeeID{}, fmt.Errorf("qtx.UpdateTeacherSpecialFee(): %w", err)
		}
		teacherSpecialFeeIDs = append(teacherSpecialFeeIDs, spec.TeacherSpecialFeeID)
	}

	err = tx.Commit()
	if err != nil {
		return []entity.TeacherSpecialFeeID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return teacherSpecialFeeIDs, nil
}

func (s entityServiceImpl) DeleteTeacherSpecialFees(ctx context.Context, ids []entity.TeacherSpecialFeeID) error {
	teacherSpecialFeeIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		teacherSpecialFeeIdsInt64 = append(teacherSpecialFeeIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteTeacherSpecialFeesByIds(ctx, teacherSpecialFeeIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteTeacherSpecialFeeByIds(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetEnrollmentPayments(ctx context.Context, pagination util.PaginationSpec) (entity.GetEnrollmentPaymentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	enrollmentPaymentRows, err := s.mySQLQueries.GetEnrollmentPayments(ctx, mysql.GetEnrollmentPaymentsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetEnrollmentPaymentsResult{}, fmt.Errorf("mySQLQueries.GetEnrollmentPayments(): %w", err)
	}

	enrollmentPayments := NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow(enrollmentPaymentRows)

	totalResults, err := s.mySQLQueries.CountEnrollmentPayments(ctx)
	if err != nil {
		return entity.GetEnrollmentPaymentsResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetEnrollmentPaymentsResult{
		EnrollmentPayments: enrollmentPayments,
		PaginationResult:   *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetEnrollmentPaymentById(ctx context.Context, id entity.EnrollmentPaymentID) (entity.EnrollmentPayment, error) {
	enrollmentPaymentRow, err := s.mySQLQueries.GetEnrollmentPaymentById(ctx, int64(id))
	if err != nil {
		return entity.EnrollmentPayment{}, fmt.Errorf("mySQLQueries.GetEnrollmentPaymentById(): %w", err)
	}

	enrollmentPayment := NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow([]mysql.GetEnrollmentPaymentsRow{enrollmentPaymentRow.ToGetEnrollmentPaymentsRow()})[0]

	return enrollmentPayment, nil
}

func (s entityServiceImpl) GetEnrollmentPaymentsByIds(ctx context.Context, ids []entity.EnrollmentPaymentID) ([]entity.EnrollmentPayment, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	enrollmentPaymentRows, err := s.mySQLQueries.GetEnrollmentPaymentsByIds(ctx, idsInt)
	if err != nil {
		return []entity.EnrollmentPayment{}, fmt.Errorf("mySQLQueries.GetEnrollmentPaymentsByIds(): %w", err)
	}

	enrollmentPaymentRowsConverted := make([]mysql.GetEnrollmentPaymentsRow, 0, len(enrollmentPaymentRows))
	for _, row := range enrollmentPaymentRows {
		enrollmentPaymentRowsConverted = append(enrollmentPaymentRowsConverted, row.ToGetEnrollmentPaymentsRow())
	}

	enrollmentPayments := NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow(enrollmentPaymentRowsConverted)

	return enrollmentPayments, nil
}

func (s entityServiceImpl) InsertEnrollmentPayments(ctx context.Context, specs []entity.InsertEnrollmentPaymentSpec) ([]entity.EnrollmentPaymentID, error) {
	enrollmentPaymentIDs := make([]entity.EnrollmentPaymentID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.EnrollmentPaymentID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		enrollmentPaymentID, err := qtx.InsertEnrollmentPayment(ctx, mysql.InsertEnrollmentPaymentParams{
			PaymentDate:  spec.PaymentDate,
			BalanceTopUp: spec.BalanceTopUp,
			Value:        spec.Value,
			ValuePenalty: spec.ValuePenalty,
			EnrollmentID: sql.NullInt64{Int64: int64(spec.StudentEnrollmentID), Valid: true},
		})
		if err != nil {

			return []entity.EnrollmentPaymentID{}, fmt.Errorf("qtx.InsertEnrollmentPayment(): %w", err)
		}
		enrollmentPaymentIDs = append(enrollmentPaymentIDs, entity.EnrollmentPaymentID(enrollmentPaymentID))
	}

	err = tx.Commit()
	if err != nil {
		return []entity.EnrollmentPaymentID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return enrollmentPaymentIDs, nil
}

func (s entityServiceImpl) UpdateEnrollmentPayments(ctx context.Context, specs []entity.UpdateEnrollmentPaymentSpec) ([]entity.EnrollmentPaymentID, error) {
	enrollmentPaymentIDs := make([]entity.EnrollmentPaymentID, 0, len(specs))

	tx, err := s.mySQLQueries.BeginTx(ctx, nil)
	if err != nil {
		return []entity.EnrollmentPaymentID{}, fmt.Errorf("mySQLQueries.BeginTx(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTxWrappedError(tx)

	for _, spec := range specs {
		err := qtx.UpdateEnrollmentPayment(ctx, mysql.UpdateEnrollmentPaymentParams{
			PaymentDate:  spec.PaymentDate,
			BalanceTopUp: spec.BalanceTopUp,
			Value:        spec.Value,
			ValuePenalty: spec.ValuePenalty,
			ID:           int64(spec.EnrollmentPaymentID),
		})
		if err != nil {

			return []entity.EnrollmentPaymentID{}, fmt.Errorf("qtx.UpdateEnrollmentPayment(): %w", err)
		}
		enrollmentPaymentIDs = append(enrollmentPaymentIDs, spec.EnrollmentPaymentID)
	}

	err = tx.Commit()
	if err != nil {
		return []entity.EnrollmentPaymentID{}, fmt.Errorf("tx.Commit(): %w", err)
	}

	return enrollmentPaymentIDs, nil
}

func (s entityServiceImpl) DeleteEnrollmentPayments(ctx context.Context, ids []entity.EnrollmentPaymentID) error {
	enrollmentPaymentIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		enrollmentPaymentIdsInt64 = append(enrollmentPaymentIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteEnrollmentPaymentsByIds(ctx, enrollmentPaymentIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteEnrollmentPaymentByIds(): %w", err)
	}

	return nil
}
