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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, userID := range userIDs {
			teacherID, err := s.mySQLQueries.InsertTeacher(ctx, int64(userID))
			if err != nil {
				return fmt.Errorf("qtx.InsertTeacher(): %w", err)
			}
			teacherIDs = append(teacherIDs, entity.TeacherID(teacherID))
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teacherIDs, nil
}

func (s entityServiceImpl) InsertTeachersWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]entity.TeacherID, error) {
	teacherIDs := make([]entity.TeacherID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		userIDs, err := s.identityService.InsertUsers(newCtx, specs)
		if err != nil {
			return fmt.Errorf("identityService.InsertUsers(): %w", err)
		}

		for _, userID := range userIDs {
			teacherID, err := qtx.InsertTeacher(newCtx, int64(userID))
			if err != nil {
				return fmt.Errorf("qtx.InsertTeacher(): %w", err)
			}
			teacherIDs = append(teacherIDs, entity.TeacherID(teacherID))
		}

		return nil
	})
	if err != nil {
		return []entity.TeacherID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		userIDs, err := s.identityService.InsertUsers(newCtx, specs)
		if err != nil {
			return fmt.Errorf("identityService.InsertUsers(): %w", err)
		}

		for _, userID := range userIDs {
			studentID, err := qtx.InsertStudent(newCtx, int64(userID))
			if err != nil {
				return fmt.Errorf("qtx.InsertStudent(): %w", err)
			}
			studentIDs = append(studentIDs, entity.StudentID(studentID))
		}

		return nil
	})
	if err != nil {
		return []entity.StudentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			instrumentID, err := qtx.InsertInstrument(newCtx, spec.Name)
			if err != nil {

				return fmt.Errorf("qtx.InsertInstrument(): %w", err)
			}
			instrumentIDs = append(instrumentIDs, entity.InstrumentID(instrumentID))
		}
		return nil
	})
	if err != nil {
		return []entity.InstrumentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return instrumentIDs, nil
}

func (s entityServiceImpl) UpdateInstruments(ctx context.Context, specs []entity.UpdateInstrumentSpec) ([]entity.InstrumentID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountInstrumentsByIds)
	if errV != nil {
		return []entity.InstrumentID{}, errV
	}

	instrumentIDs := make([]entity.InstrumentID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateInstrument(newCtx, mysql.UpdateInstrumentParams{
				Name: spec.Name,
				ID:   int64(spec.InstrumentID),
			})
			if err != nil {

				return fmt.Errorf("qtx.UpdateInstrument(): %w", err)
			}
			instrumentIDs = append(instrumentIDs, spec.InstrumentID)
		}
		return nil
	})
	if err != nil {
		return []entity.InstrumentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			gradeID, err := qtx.InsertGrade(newCtx, spec.Name)
			if err != nil {

				return fmt.Errorf("qtx.InsertGrade(): %w", err)
			}
			gradeIDs = append(gradeIDs, entity.GradeID(gradeID))
		}
		return nil
	})
	if err != nil {
		return []entity.GradeID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return gradeIDs, nil
}

func (s entityServiceImpl) UpdateGrades(ctx context.Context, specs []entity.UpdateGradeSpec) ([]entity.GradeID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountGradesByIds)
	if errV != nil {
		return []entity.GradeID{}, errV
	}

	gradeIDs := make([]entity.GradeID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateGrade(newCtx, mysql.UpdateGradeParams{
				Name: spec.Name,
				ID:   int64(spec.GradeID),
			})
			if err != nil {

				return fmt.Errorf("qtx.UpdateGrade(): %w", err)
			}
			gradeIDs = append(gradeIDs, spec.GradeID)
		}
		return nil
	})
	if err != nil {
		return []entity.GradeID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			courseID, err := qtx.InsertCourse(newCtx, mysql.InsertCourseParams{
				DefaultFee:            spec.DefaultFee,
				DefaultDurationMinute: spec.DefaultDurationMinute,
				InstrumentID:          int64(spec.InstrumentID),
				GradeID:               int64(spec.GradeID),
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertCourse(): %w", err)
			}
			courseIDs = append(courseIDs, entity.CourseID(courseID))
		}
		return nil
	})
	if err != nil {
		return []entity.CourseID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return courseIDs, nil
}

func (s entityServiceImpl) UpdateCourses(ctx context.Context, specs []entity.UpdateCourseSpec) ([]entity.CourseID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountCoursesByIds)
	if errV != nil {
		return []entity.CourseID{}, errV
	}

	courseIDs := make([]entity.CourseID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateCourseInfo(newCtx, mysql.UpdateCourseInfoParams{
				DefaultFee:            spec.DefaultFee,
				DefaultDurationMinute: spec.DefaultDurationMinute,
				ID:                    int64(spec.CourseID),
			})
			if err != nil {

				return fmt.Errorf("qtx.UpdateCourseInfo(): %w", err)
			}
			courseIDs = append(courseIDs, spec.CourseID)
		}
		return nil
	})
	if err != nil {
		return []entity.CourseID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

func (s entityServiceImpl) GetClasses(ctx context.Context, pagination util.PaginationSpec, spec entity.GetClassesSpec) (entity.GetClassesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	isDeactivatedFilters := []int32{0}
	if spec.IncludeDeactivated {
		isDeactivatedFilters = append(isDeactivatedFilters, 1)
	}

	teacherID := sql.NullInt64{Int64: int64(spec.TeacherID), Valid: true}
	studentID := sql.NullInt64{Int64: int64(spec.StudentID), Valid: true}
	courseID := sql.NullInt64{Int64: int64(spec.CourseID), Valid: true}

	useTeacherFilter := spec.TeacherID != entity.TeacherID_None
	useStudentFilter := spec.StudentID != entity.StudentID_None
	useCourseFilter := spec.CourseID != entity.CourseID_None

	classRows, err := s.mySQLQueries.GetClasses(ctx, mysql.GetClassesParams{
		IsDeactivateds:   isDeactivatedFilters,
		TeacherID:        teacherID,
		UseTeacherFilter: useTeacherFilter,
		StudentID:        studentID,
		UseStudentFilter: useStudentFilter,
		CourseID:         courseID,
		UseCourseFilter:  useCourseFilter,
		Limit:            int32(limit),
		Offset:           int32(offset),
	})
	if err != nil {
		return entity.GetClassesResult{}, fmt.Errorf("mySQLQueries.GetClasses(): %w", err)
	}

	classes := NewClassesFromGetClassesRow(classRows)

	totalResults, err := s.mySQLQueries.CountClasses(ctx, mysql.CountClassesParams{
		IsDeactivateds:   isDeactivatedFilters,
		TeacherID:        teacherID,
		UseTeacherFilter: useTeacherFilter,
		StudentID:        studentID,
		UseStudentFilter: useStudentFilter,
		CourseID:         courseID,
		UseCourseFilter:  useCourseFilter,
	})
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			classID, err := qtx.InsertClass(newCtx, mysql.InsertClassParams{
				TransportFee: spec.TransportFee,
				TeacherID:    sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != entity.TeacherID_None},
				CourseID:     int64(spec.CourseID),
			})
			if err != nil {

				return fmt.Errorf("qtx.InsertClass(): %w", err)
			}
			classIDs = append(classIDs, entity.ClassID(classID))

			for _, studentId := range spec.StudentIDs {
				err := qtx.InsertStudentEnrollment(newCtx, mysql.InsertStudentEnrollmentParams{
					StudentID: int64(studentId),
					ClassID:   classID,
				})
				if err != nil {

					return fmt.Errorf("qtx.InsertStudentEnrollment(): %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return []entity.ClassID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return classIDs, nil
}

func (s entityServiceImpl) UpdateClasses(ctx context.Context, specs []entity.UpdateClassSpec) ([]entity.ClassID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountClassesByIds)
	if errV != nil {
		return []entity.ClassID{}, errV
	}

	classIDs := make([]entity.ClassID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			classId := int64(spec.ClassID)
			// Updated class
			err := s.mySQLQueries.UpdateClass(newCtx, mysql.UpdateClassParams{
				TransportFee:  spec.TransportFee,
				TeacherID:     sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != entity.TeacherID_None},
				CourseID:      int64(spec.CourseID),
				IsDeactivated: util.BoolToInt32(spec.IsDeactivated),
				ID:            classId,
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateClass(): %w", err)
			}
			classIDs = append(classIDs, spec.ClassID)

			// we only know the initial & final states of the class' students.
			// so, we need to calculate the difference manually to know which DB action to be executed (insert, delete, or update [enable/disable]).
			studentDifference, err := calculateClassStudentsDifference(newCtx, qtx, spec.ClassID, spec.StudentIDs)
			if err != nil {
				return fmt.Errorf("calculateClassStudentsDifference(): %w", err)
			}

			// Added students
			for _, studentId := range studentDifference.addedStudentIDs {
				err = qtx.InsertStudentEnrollment(newCtx, mysql.InsertStudentEnrollmentParams{
					StudentID: int64(studentId),
					ClassID:   classId,
				})
				if err != nil {

					return fmt.Errorf("qtx.InsertStudentEnrollment(): %w", err)
				}
			}

			// Updated (re-enabled) enrollments
			for _, updatedEnrollmentID := range studentDifference.enabledStudentEnrollmentIDs {
				err = qtx.EnableStudentEnrollment(newCtx, int64(updatedEnrollmentID))
				if err != nil {

					return fmt.Errorf("qtx.EnableStudentEnrollment(): %w", err)
				}
			}

			// Delete or disable enrollments
			for _, disabledEnrollmentID := range studentDifference.disabledStudentEnrollmentIDs {
				err = qtx.DeleteStudentEnrollmentById(newCtx, int64(disabledEnrollmentID))
				if err == nil {
					// the enrollment is still deletable (not referenced by any other entity), then we straightforwardly delete it
					continue
				}

				err = qtx.DisableStudentEnrollment(newCtx, int64(disabledEnrollmentID))
				if err != nil {
					return fmt.Errorf("qtx.DisableStudentEnrollment(): %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return []entity.ClassID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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
		studentIDToEnrollmentIDMap[entity.StudentID(enrollment.StudentID)] = entity.StudentEnrollmentID(enrollment.StudentEnrollmentID)
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
			disabledStudentEnrollmentIDs = append(disabledStudentEnrollmentIDs, entity.StudentEnrollmentID(enrollment.StudentEnrollmentID))
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := qtx.DeleteStudentEnrollmentByClassIds(newCtx, classIdsInt64)
		if err != nil {
			return fmt.Errorf("qtx.DeleteStudentEnrollmentByClassIds(): %w", err)
		}

		err = qtx.DeleteClassesByIds(newCtx, classIdsInt64)
		if err != nil {
			return fmt.Errorf("qtx.DeleteClassesByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetStudentEnrollments(ctx context.Context, pagination util.PaginationSpec) (entity.GetStudentEnrollmentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	studentEnrollmentRows, err := s.mySQLQueries.GetStudentEnrollments(ctx, mysql.GetStudentEnrollmentsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetStudentEnrollmentsResult{}, fmt.Errorf("mySQLQueries.GetStudentEnrollments(): %w", err)
	}

	studentEnrollments := NewStudentEnrollmentsFromGetStudentEnrollmentsRow(studentEnrollmentRows)

	totalResults, err := s.mySQLQueries.CountStudentEnrollments(ctx)
	if err != nil {
		return entity.GetStudentEnrollmentsResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetStudentEnrollmentsResult{
		StudentEnrollments: studentEnrollments,
		PaginationResult:   *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetStudentEnrollmentById(ctx context.Context, id entity.StudentEnrollmentID) (entity.StudentEnrollment, error) {
	studentEnrollmentRow, err := s.mySQLQueries.GetStudentEnrollmentById(ctx, int64(id))
	if err != nil {
		return entity.StudentEnrollment{}, fmt.Errorf("mySQLQueries.GetStudentEnrollmentById(): %w", err)
	}

	teacherSpecialFee := NewStudentEnrollmentsFromGetStudentEnrollmentsRow([]mysql.GetStudentEnrollmentsRow{studentEnrollmentRow.ToGetStudentEnrollmentsRow()})[0]

	return teacherSpecialFee, nil
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			teacherSpecialFeeID, err := qtx.InsertTeacherSpecialFee(newCtx, mysql.InsertTeacherSpecialFeeParams{
				Fee:       spec.Fee,
				TeacherID: int64(spec.TeacherID),
				CourseID:  int64(spec.CourseID),
			})
			if err != nil {

				return fmt.Errorf("qtx.InsertTeacherSpecialFee(): %w", err)
			}
			teacherSpecialFeeIDs = append(teacherSpecialFeeIDs, entity.TeacherSpecialFeeID(teacherSpecialFeeID))
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherSpecialFeeID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teacherSpecialFeeIDs, nil
}

func (s entityServiceImpl) UpdateTeacherSpecialFees(ctx context.Context, specs []entity.UpdateTeacherSpecialFeeSpec) ([]entity.TeacherSpecialFeeID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountTeacherSpecialFeesByIds)
	if errV != nil {
		return []entity.TeacherSpecialFeeID{}, errV
	}

	teacherSpecialFeeIDs := make([]entity.TeacherSpecialFeeID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateTeacherSpecialFee(newCtx, mysql.UpdateTeacherSpecialFeeParams{
				Fee: spec.Fee,
				ID:  int64(spec.TeacherSpecialFeeID),
			})
			if err != nil {

				return fmt.Errorf("qtx.UpdateTeacherSpecialFee(): %w", err)
			}
			teacherSpecialFeeIDs = append(teacherSpecialFeeIDs, spec.TeacherSpecialFeeID)
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherSpecialFeeID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

func (s entityServiceImpl) GetEnrollmentPayments(ctx context.Context, pagination util.PaginationSpec, timeFilter util.TimeSpec, sortRecent bool) (entity.GetEnrollmentPaymentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	timeFilter.SetDefaultForZeroValues()
	enrollmentPayments := make([]entity.EnrollmentPayment, 0)

	if !sortRecent { // sqlc dynamic query is so bad that we need to do this :(
		enrollmentPaymentRows, err := s.mySQLQueries.GetEnrollmentPayments(ctx, mysql.GetEnrollmentPaymentsParams{
			StartDate: timeFilter.StartDatetime,
			EndDate:   timeFilter.EndDatetime,
			Limit:     int32(limit),
			Offset:    int32(offset),
		})
		if err != nil {
			return entity.GetEnrollmentPaymentsResult{}, fmt.Errorf("mySQLQueries.GetEnrollmentPayments(): %w", err)
		}
		enrollmentPayments = NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow(enrollmentPaymentRows)

	} else {
		enrollmentPaymentRows, err := s.mySQLQueries.GetEnrollmentPaymentsDescendingDate(ctx, mysql.GetEnrollmentPaymentsDescendingDateParams{
			StartDate: timeFilter.StartDatetime,
			EndDate:   timeFilter.EndDatetime,
			Limit:     int32(limit),
			Offset:    int32(offset),
		})
		if err != nil {
			return entity.GetEnrollmentPaymentsResult{}, fmt.Errorf("mySQLQueries.GetEnrollmentPaymentsDescendingDate(): %w", err)
		}

		enrollmentPaymentRowsConverted := make([]mysql.GetEnrollmentPaymentsRow, 0, len(enrollmentPaymentRows))
		for _, row := range enrollmentPaymentRows {
			enrollmentPaymentRowsConverted = append(enrollmentPaymentRowsConverted, row.ToGetEnrollmentPaymentsRow())
		}
		enrollmentPayments = NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow(enrollmentPaymentRowsConverted)
	}

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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			enrollmentPaymentID, err := qtx.InsertEnrollmentPayment(newCtx, mysql.InsertEnrollmentPaymentParams{
				PaymentDate:       spec.PaymentDate,
				BalanceTopUp:      spec.BalanceTopUp,
				CourseFeeValue:    spec.CourseFeeValue,
				TransportFeeValue: spec.TransportFeeValue,
				PenaltyFeeValue:   spec.PenaltyFeeValue,
				EnrollmentID:      sql.NullInt64{Int64: int64(spec.StudentEnrollmentID), Valid: true},
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertEnrollmentPayment(): %w", err)
			}
			enrollmentPaymentIDs = append(enrollmentPaymentIDs, entity.EnrollmentPaymentID(enrollmentPaymentID))
		}
		return nil
	})
	if err != nil {
		return []entity.EnrollmentPaymentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}
	return enrollmentPaymentIDs, nil
}

func (s entityServiceImpl) UpdateEnrollmentPayments(ctx context.Context, specs []entity.UpdateEnrollmentPaymentSpec) ([]entity.EnrollmentPaymentID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountEnrollmentPaymentsByIds)
	if errV != nil {
		return []entity.EnrollmentPaymentID{}, errV
	}

	enrollmentPaymentIDs := make([]entity.EnrollmentPaymentID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateEnrollmentPayment(newCtx, mysql.UpdateEnrollmentPaymentParams{
				PaymentDate:       spec.PaymentDate,
				BalanceTopUp:      spec.BalanceTopUp,
				CourseFeeValue:    spec.CourseFeeValue,
				TransportFeeValue: spec.TransportFeeValue,
				PenaltyFeeValue:   spec.PenaltyFeeValue,
				ID:                int64(spec.EnrollmentPaymentID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateEnrollmentPayment(): %w", err)
			}
			enrollmentPaymentIDs = append(enrollmentPaymentIDs, spec.EnrollmentPaymentID)
		}
		return nil
	})
	if err != nil {
		return []entity.EnrollmentPaymentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

func (s entityServiceImpl) GetStudentLearningTokens(ctx context.Context, pagination util.PaginationSpec) (entity.GetStudentLearningTokensResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	studentLearningTokenRows, err := s.mySQLQueries.GetStudentLearningTokens(ctx, mysql.GetStudentLearningTokensParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return entity.GetStudentLearningTokensResult{}, fmt.Errorf("mySQLQueries.GetStudentLearningTokens(): %w", err)
	}

	studentLearningTokens := NewStudentLearningTokensFromGetStudentLearningTokensRow(studentLearningTokenRows)

	totalResults, err := s.mySQLQueries.CountStudentLearningTokens(ctx)
	if err != nil {
		return entity.GetStudentLearningTokensResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetStudentLearningTokensResult{
		StudentLearningTokens: studentLearningTokens,
		PaginationResult:      *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetStudentLearningTokenById(ctx context.Context, id entity.StudentLearningTokenID) (entity.StudentLearningToken, error) {
	studentLearningTokenRow, err := s.mySQLQueries.GetStudentLearningTokenById(ctx, int64(id))
	if err != nil {
		return entity.StudentLearningToken{}, fmt.Errorf("mySQLQueries.GetStudentLearningTokenById(): %w", err)
	}

	studentLearningToken := NewStudentLearningTokensFromGetStudentLearningTokensRow([]mysql.GetStudentLearningTokensRow{studentLearningTokenRow.ToGetStudentLearningTokensRow()})[0]

	return studentLearningToken, nil
}

func (s entityServiceImpl) GetStudentLearningTokensByIds(ctx context.Context, ids []entity.StudentLearningTokenID) ([]entity.StudentLearningToken, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	studentLearningTokenRows, err := s.mySQLQueries.GetStudentLearningTokensByIds(ctx, idsInt)
	if err != nil {
		return []entity.StudentLearningToken{}, fmt.Errorf("mySQLQueries.GetStudentLearningTokensByIds(): %w", err)
	}

	studentLearningTokenRowsConverted := make([]mysql.GetStudentLearningTokensRow, 0, len(studentLearningTokenRows))
	for _, row := range studentLearningTokenRows {
		studentLearningTokenRowsConverted = append(studentLearningTokenRowsConverted, row.ToGetStudentLearningTokensRow())
	}

	studentLearningTokens := NewStudentLearningTokensFromGetStudentLearningTokensRow(studentLearningTokenRowsConverted)

	return studentLearningTokens, nil
}

func (s entityServiceImpl) InsertStudentLearningTokens(ctx context.Context, specs []entity.InsertStudentLearningTokenSpec) ([]entity.StudentLearningTokenID, error) {
	studentLearningTokenIDs := make([]entity.StudentLearningTokenID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			studentLearningTokenID, err := qtx.InsertStudentLearningToken(newCtx, mysql.InsertStudentLearningTokenParams{
				Quota:             spec.Quota,
				CourseFeeValue:    spec.CourseFeeValue,
				TransportFeeValue: spec.TransportFeeValue,
				EnrollmentID:      int64(spec.StudentEnrollmentID),
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertStudentLearningToken(): %w", err)
			}
			studentLearningTokenIDs = append(studentLearningTokenIDs, entity.StudentLearningTokenID(studentLearningTokenID))
		}
		return nil
	})
	if err != nil {
		return []entity.StudentLearningTokenID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return studentLearningTokenIDs, nil
}

func (s entityServiceImpl) UpdateStudentLearningTokens(ctx context.Context, specs []entity.UpdateStudentLearningTokenSpec) ([]entity.StudentLearningTokenID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountStudentLearningTokensByIds)
	if errV != nil {
		return []entity.StudentLearningTokenID{}, errV
	}

	studentLearningTokenIDs := make([]entity.StudentLearningTokenID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateStudentLearningToken(newCtx, mysql.UpdateStudentLearningTokenParams{
				Quota:             spec.Quota,
				CourseFeeValue:    spec.CourseFeeValue,
				TransportFeeValue: spec.TransportFeeValue,
				ID:                int64(spec.StudentLearningTokenID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateStudentLearningToken(): %w", err)
			}
			studentLearningTokenIDs = append(studentLearningTokenIDs, spec.StudentLearningTokenID)
		}
		return nil
	})
	if err != nil {
		return []entity.StudentLearningTokenID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return studentLearningTokenIDs, nil
}

func (s entityServiceImpl) DeleteStudentLearningTokens(ctx context.Context, ids []entity.StudentLearningTokenID) error {
	studentLearningTokenIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		studentLearningTokenIdsInt64 = append(studentLearningTokenIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteStudentLearningTokensByIds(ctx, studentLearningTokenIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteStudentLearningTokenByIds(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetPresences(ctx context.Context, pagination util.PaginationSpec, timeFilter util.TimeSpec) (entity.GetPresencesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	timeFilter.SetDefaultForZeroValues()
	presenceRows, err := s.mySQLQueries.GetPresences(ctx, mysql.GetPresencesParams{
		StartDate: timeFilter.StartDatetime,
		EndDate:   timeFilter.EndDatetime,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return entity.GetPresencesResult{}, fmt.Errorf("mySQLQueries.GetPresences(): %w", err)
	}

	presences := NewPresencesFromGetPresencesRow(presenceRows)

	totalResults, err := s.mySQLQueries.CountPresences(ctx, mysql.CountPresencesParams{
		Date:   timeFilter.StartDatetime,
		Date_2: timeFilter.EndDatetime,
	})
	if err != nil {
		return entity.GetPresencesResult{}, fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
	}

	return entity.GetPresencesResult{
		Presences:        presences,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetPresenceById(ctx context.Context, id entity.PresenceID) (entity.Presence, error) {
	presenceRow, err := s.mySQLQueries.GetPresenceById(ctx, int64(id))
	if err != nil {
		return entity.Presence{}, fmt.Errorf("mySQLQueries.GetPresenceById(): %w", err)
	}

	presence := NewPresencesFromGetPresencesRow([]mysql.GetPresencesRow{presenceRow.ToGetPresencesRow()})[0]

	return presence, nil
}

func (s entityServiceImpl) GetPresencesByIds(ctx context.Context, ids []entity.PresenceID) ([]entity.Presence, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	presenceRows, err := s.mySQLQueries.GetPresencesByIds(ctx, idsInt)
	if err != nil {
		return []entity.Presence{}, fmt.Errorf("mySQLQueries.GetPresencesByIds(): %w", err)
	}

	presenceRowsConverted := make([]mysql.GetPresencesRow, 0, len(presenceRows))
	for _, row := range presenceRows {
		presenceRowsConverted = append(presenceRowsConverted, row.ToGetPresencesRow())
	}

	presences := NewPresencesFromGetPresencesRow(presenceRowsConverted)

	return presences, nil
}

func (s entityServiceImpl) InsertPresences(ctx context.Context, specs []entity.InsertPresenceSpec) ([]entity.PresenceID, error) {
	presenceIDs := make([]entity.PresenceID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			presenceID, err := qtx.InsertPresence(newCtx, mysql.InsertPresenceParams{
				Date:                  spec.Date,
				UsedStudentTokenQuota: spec.UsedStudentTokenQuota,
				Duration:              spec.Duration,
				Note:                  spec.Note,
				ClassID:               sql.NullInt64{Int64: int64(spec.ClassID), Valid: true},
				TeacherID:             sql.NullInt64{Int64: int64(spec.TeacherID), Valid: true},
				StudentID:             sql.NullInt64{Int64: int64(spec.StudentID), Valid: true},
				TokenID:               int64(spec.StudentLearningTokenID),
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertPresence(): %w", err)
			}
			presenceIDs = append(presenceIDs, entity.PresenceID(presenceID))
		}
		return nil
	})
	if err != nil {
		return []entity.PresenceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}
	return presenceIDs, nil
}

func (s entityServiceImpl) UpdatePresences(ctx context.Context, specs []entity.UpdatePresenceSpec) ([]entity.PresenceID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountPresencesByIds)
	if errV != nil {
		return []entity.PresenceID{}, errV
	}

	presenceIDs := make([]entity.PresenceID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdatePresence(ctx, mysql.UpdatePresenceParams{
				Date:                  spec.Date,
				UsedStudentTokenQuota: spec.UsedStudentTokenQuota,
				Duration:              spec.Duration,
				Note:                  spec.Note,
				ClassID:               sql.NullInt64{Int64: int64(spec.ClassID), Valid: true},
				TeacherID:             sql.NullInt64{Int64: int64(spec.TeacherID), Valid: true},
				StudentID:             sql.NullInt64{Int64: int64(spec.StudentID), Valid: true},
				TokenID:               int64(spec.StudentLearningTokenID),
				ID:                    int64(spec.PresenceID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdatePresence(): %w", err)
			}
			presenceIDs = append(presenceIDs, spec.PresenceID)
		}
		return nil
	})
	if err != nil {
		return []entity.PresenceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return presenceIDs, nil
}

func (s entityServiceImpl) DeletePresences(ctx context.Context, ids []entity.PresenceID) error {
	presenceIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		presenceIdsInt64 = append(presenceIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeletePresencesByIds(ctx, presenceIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeletePresenceByIds(): %w", err)
	}

	return nil
}
