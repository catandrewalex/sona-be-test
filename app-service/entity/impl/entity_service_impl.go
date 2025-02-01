package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

	var teacherRows = make([]mysql.GetTeachersRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherRows, err = s.mySQLQueries.GetTeachers(newCtx, mysql.GetTeachersParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeachers(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountTeachers(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountTeachers(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetTeachersResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teachers := NewTeachersFromGetTeachersRow(teacherRows)

	return entity.GetTeachersResult{
		Teachers:         teachers,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetTeacherById(ctx context.Context, id entity.TeacherID) (entity.Teacher, error) {
	var teacherRow mysql.GetTeacherByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherRow, err = s.mySQLQueries.GetTeacherById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.Teacher{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teacher := NewTeachersFromGetTeachersRow([]mysql.GetTeachersRow{teacherRow.ToGetTeachersRow()})[0]

	return teacher, nil
}

func (s entityServiceImpl) GetTeachersByIds(ctx context.Context, ids []entity.TeacherID) ([]entity.Teacher, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var teacherRows = make([]mysql.GetTeachersByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherRows, err = s.mySQLQueries.GetTeachersByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeachersByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.Teacher{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteTeachersByIds(newCtx, teacherIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteTeacherByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetStudents(ctx context.Context, pagination util.PaginationSpec) (entity.GetStudentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	var studentRows = make([]mysql.GetStudentsRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentRows, err = s.mySQLQueries.GetStudents(newCtx, mysql.GetStudentsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudents(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountStudents(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountStudents(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetStudentsResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	students := NewStudentsFromGetStudentsRow(studentRows)

	return entity.GetStudentsResult{
		Students:         students,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetStudentById(ctx context.Context, id entity.StudentID) (entity.Student, error) {
	var studentRow mysql.GetStudentByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentRow, err = s.mySQLQueries.GetStudentById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.Student{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	student := NewStudentsFromGetStudentsRow([]mysql.GetStudentsRow{studentRow.ToGetStudentsRow()})[0]

	return student, nil
}

func (s entityServiceImpl) GetStudentsByIds(ctx context.Context, ids []entity.StudentID) ([]entity.Student, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var studentRows = make([]mysql.GetStudentsByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentRows, err = s.mySQLQueries.GetStudentsByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentsByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.Student{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, userID := range userIDs {
			studentID, err := s.mySQLQueries.InsertStudent(newCtx, int64(userID))
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteStudentsByIds(newCtx, studentIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteStudentByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetInstruments(ctx context.Context, pagination util.PaginationSpec) (entity.GetInstrumentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	var instrumentRows = make([]mysql.Instrument, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		instrumentRows, err = s.mySQLQueries.GetInstruments(newCtx, mysql.GetInstrumentsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetInstruments(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountInstruments(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountInstruments(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetInstrumentsResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	instruments := NewInstrumentsFromMySQLInstruments(instrumentRows)

	return entity.GetInstrumentsResult{
		Instruments:      instruments,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetInstrumentById(ctx context.Context, id entity.InstrumentID) (entity.Instrument, error) {
	var instrumentRow mysql.Instrument
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		instrumentRow, err = s.mySQLQueries.GetInstrumentById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetInstrumentById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.Instrument{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	instrument := NewInstrumentsFromMySQLInstruments([]mysql.Instrument{instrumentRow})[0]

	return instrument, nil
}

func (s entityServiceImpl) GetInstrumentsByIds(ctx context.Context, ids []entity.InstrumentID) ([]entity.Instrument, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var instrumentRows = make([]mysql.Instrument, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		instrumentRows, err = s.mySQLQueries.GetInstrumentsByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetInstrumentsByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.Instrument{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteInstrumentsByIds(newCtx, instrumentIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteInstrumentsByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetGrades(ctx context.Context, pagination util.PaginationSpec) (entity.GetGradesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	var gradeRows = make([]mysql.Grade, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		gradeRows, err = s.mySQLQueries.GetGrades(newCtx, mysql.GetGradesParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetGrades(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountGrades(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountGrades(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetGradesResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	grades := NewGradesFromMySQLGrades(gradeRows)

	return entity.GetGradesResult{
		Grades:           grades,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetGradeById(ctx context.Context, id entity.GradeID) (entity.Grade, error) {
	var gradeRow mysql.Grade
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		gradeRow, err = s.mySQLQueries.GetGradeById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetGradeById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.Grade{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	grade := NewGradesFromMySQLGrades([]mysql.Grade{gradeRow})[0]

	return grade, nil
}

func (s entityServiceImpl) GetGradesByIds(ctx context.Context, ids []entity.GradeID) ([]entity.Grade, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var gradeRows = make([]mysql.Grade, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		gradeRows, err = s.mySQLQueries.GetGradesByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetGradesByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.Grade{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteGradesByIds(newCtx, gradeIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteGradesByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetCourses(ctx context.Context, pagination util.PaginationSpec) (entity.GetCoursesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	var courseRows = make([]mysql.GetCoursesRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		courseRows, err = s.mySQLQueries.GetCourses(newCtx, mysql.GetCoursesParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetCourses(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountCourses(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountCourses(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetCoursesResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	courses := NewCoursesFromGetCoursesRow(courseRows)

	return entity.GetCoursesResult{
		Courses:          courses,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetCourseById(ctx context.Context, id entity.CourseID) (entity.Course, error) {
	var courseRow mysql.GetCourseByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		courseRow, err = s.mySQLQueries.GetCourseById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetCourseById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.Course{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	course := NewCoursesFromGetCoursesRow([]mysql.GetCoursesRow{courseRow.ToGetCoursesRow()})[0]

	return course, nil
}

func (s entityServiceImpl) GetCoursesByIds(ctx context.Context, ids []entity.CourseID) ([]entity.Course, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var courseRows = make([]mysql.GetCoursesByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		courseRows, err = s.mySQLQueries.GetCoursesByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetCoursesByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.Course{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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
				GradeID:               int64(spec.GradeID),
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteCoursesByIds(newCtx, courseIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteCourseByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	var classRows = make([]mysql.GetClassesRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		classRows, err = s.mySQLQueries.GetClasses(newCtx, mysql.GetClassesParams{
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
			return fmt.Errorf("mySQLQueries.GetClasses(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountClasses(newCtx, mysql.CountClassesParams{
			IsDeactivateds:   isDeactivatedFilters,
			TeacherID:        teacherID,
			UseTeacherFilter: useTeacherFilter,
			StudentID:        studentID,
			UseStudentFilter: useStudentFilter,
			CourseID:         courseID,
			UseCourseFilter:  useCourseFilter,
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountClasses(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetClassesResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	classes := NewClassesFromGetClassesRow(classRows)

	return entity.GetClassesResult{
		Classes:          classes,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetClassById(ctx context.Context, id entity.ClassID) (entity.Class, error) {
	var classRows = make([]mysql.GetClassByIdRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		classRows, err = s.mySQLQueries.GetClassById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetClassById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.Class{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	var classRows = make([]mysql.GetClassesByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		classRows, err = s.mySQLQueries.GetClassesByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetClassesByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.Class{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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
				TransportFee:           spec.TransportFee,
				TeacherID:              sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != entity.TeacherID_None},
				CourseID:               int64(spec.CourseID),
				AutoOweAttendanceToken: util.BoolToInt32(true),
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
			err := qtx.UpdateClass(newCtx, mysql.UpdateClassParams{
				TransportFee:           spec.TransportFee,
				TeacherID:              sql.NullInt64{Int64: int64(spec.TeacherID), Valid: spec.TeacherID != entity.TeacherID_None},
				CourseID:               int64(spec.CourseID),
				AutoOweAttendanceToken: util.BoolToInt32(spec.AutoOweAttendanceToken),
				IsDeactivated:          util.BoolToInt32(spec.IsDeactivated),
				ID:                     classId,
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

	var studentEnrollmentRows = make([]mysql.GetStudentEnrollmentsRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentEnrollmentRows, err = s.mySQLQueries.GetStudentEnrollments(newCtx, mysql.GetStudentEnrollmentsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentEnrollments(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountStudentEnrollments(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountStudentEnrollments(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetStudentEnrollmentsResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	studentEnrollments := NewStudentEnrollmentsFromGetStudentEnrollmentsRow(studentEnrollmentRows)

	return entity.GetStudentEnrollmentsResult{
		StudentEnrollments: studentEnrollments,
		PaginationResult:   *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetStudentEnrollmentById(ctx context.Context, id entity.StudentEnrollmentID) (entity.StudentEnrollment, error) {
	var studentEnrollmentRow mysql.GetStudentEnrollmentByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentEnrollmentRow, err = s.mySQLQueries.GetStudentEnrollmentById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentEnrollmentById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.StudentEnrollment{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	studentEnrollment := NewStudentEnrollmentsFromGetStudentEnrollmentsRow([]mysql.GetStudentEnrollmentsRow{studentEnrollmentRow.ToGetStudentEnrollmentsRow()})[0]

	return studentEnrollment, nil
}

func (s entityServiceImpl) GetStudentEnrollmentsByClassId(ctx context.Context, classId entity.ClassID) ([]entity.StudentEnrollment, error) {
	var studentEnrollmentRows = make([]mysql.GetStudentEnrollmentsByClassIdRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentEnrollmentRows, err = qtx.GetStudentEnrollmentsByClassId(newCtx, int64(classId))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentEnrollmentsByClassId(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.StudentEnrollment{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	studentEnrollmentRowsConverted := make([]mysql.GetStudentEnrollmentsRow, 0, len(studentEnrollmentRows))
	for _, row := range studentEnrollmentRows {
		studentEnrollmentRowsConverted = append(studentEnrollmentRowsConverted, row.ToGetStudentEnrollmentsRow())
	}

	studentEnrollments := NewStudentEnrollmentsFromGetStudentEnrollmentsRow(studentEnrollmentRowsConverted)

	return studentEnrollments, nil
}

func (s entityServiceImpl) GetTeacherSpecialFees(ctx context.Context, pagination util.PaginationSpec) (entity.GetTeacherSpecialFeesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	var teacherSpecialFeeRows = make([]mysql.GetTeacherSpecialFeesRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherSpecialFeeRows, err = s.mySQLQueries.GetTeacherSpecialFees(newCtx, mysql.GetTeacherSpecialFeesParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherSpecialFees(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountTeacherSpecialFees(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountTeacherSpecialFees(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetTeacherSpecialFeesResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teacherSpecialFees := NewTeacherSpecialFeesFromGetTeacherSpecialFeesRow(teacherSpecialFeeRows)

	return entity.GetTeacherSpecialFeesResult{
		TeacherSpecialFees: teacherSpecialFees,
		PaginationResult:   *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetTeacherSpecialFeeById(ctx context.Context, id entity.TeacherSpecialFeeID) (entity.TeacherSpecialFee, error) {
	var teacherSpecialFeeRow mysql.GetTeacherSpecialFeeByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherSpecialFeeRow, err = s.mySQLQueries.GetTeacherSpecialFeeById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherSpecialFeeById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.TeacherSpecialFee{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teacherSpecialFee := NewTeacherSpecialFeesFromGetTeacherSpecialFeesRow([]mysql.GetTeacherSpecialFeesRow{teacherSpecialFeeRow.ToGetTeacherSpecialFeesRow()})[0]

	return teacherSpecialFee, nil
}

func (s entityServiceImpl) GetTeacherSpecialFeesByIds(ctx context.Context, ids []entity.TeacherSpecialFeeID) ([]entity.TeacherSpecialFee, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var teacherSpecialFeeRows = make([]mysql.GetTeacherSpecialFeesByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherSpecialFeeRows, err = s.mySQLQueries.GetTeacherSpecialFeesByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherSpecialFeesByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherSpecialFee{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteTeacherSpecialFeesByIds(newCtx, teacherSpecialFeeIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteTeacherSpecialFeeByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetEnrollmentPayments(ctx context.Context, pagination util.PaginationSpec, timeFilter util.TimeSpec, sortRecent bool) (entity.GetEnrollmentPaymentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	timeFilter.SetDefaultForZeroValues()

	var enrollmentPaymentRows = make([]mysql.GetEnrollmentPaymentsRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		if !sortRecent { // TODO: find alternative: currently sqlc dynamic query is so bad that we need to do this :(
			enrollmentPaymentRows, err = s.mySQLQueries.GetEnrollmentPayments(newCtx, mysql.GetEnrollmentPaymentsParams{
				StartDate: timeFilter.StartDatetime,
				EndDate:   timeFilter.EndDatetime,
				Limit:     int32(limit),
				Offset:    int32(offset),
			})
			if err != nil {
				return fmt.Errorf("mySQLQueries.GetEnrollmentPayments(): %w", err)
			}

		} else {
			enrollmentPaymentDescendingDateRows, err := s.mySQLQueries.GetEnrollmentPaymentsDescendingDate(newCtx, mysql.GetEnrollmentPaymentsDescendingDateParams{
				StartDate: timeFilter.StartDatetime,
				EndDate:   timeFilter.EndDatetime,
				Limit:     int32(limit),
				Offset:    int32(offset),
			})
			if err != nil {
				return fmt.Errorf("mySQLQueries.GetEnrollmentPaymentsDescendingDate(): %w", err)
			}

			for _, row := range enrollmentPaymentDescendingDateRows {
				enrollmentPaymentRows = append(enrollmentPaymentRows, row.ToGetEnrollmentPaymentsRow())
			}
		}

		totalResults, err = s.mySQLQueries.CountEnrollmentPayments(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountEnrollmentPayments(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetEnrollmentPaymentsResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	enrollmentPayments := NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow(enrollmentPaymentRows)

	return entity.GetEnrollmentPaymentsResult{
		EnrollmentPayments: enrollmentPayments,
		PaginationResult:   *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetEnrollmentPaymentById(ctx context.Context, id entity.EnrollmentPaymentID) (entity.EnrollmentPayment, error) {
	var enrollmentPaymentRow mysql.GetEnrollmentPaymentByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		enrollmentPaymentRow, err = s.mySQLQueries.GetEnrollmentPaymentById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetEnrollmentPaymentById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.EnrollmentPayment{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	enrollmentPayment := NewEnrollmentPaymentsFromGetEnrollmentPaymentsRow([]mysql.GetEnrollmentPaymentsRow{enrollmentPaymentRow.ToGetEnrollmentPaymentsRow()})[0]

	return enrollmentPayment, nil
}

func (s entityServiceImpl) GetEnrollmentPaymentsByIds(ctx context.Context, ids []entity.EnrollmentPaymentID) ([]entity.EnrollmentPayment, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var enrollmentPaymentRows = make([]mysql.GetEnrollmentPaymentsByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		enrollmentPaymentRows, err = s.mySQLQueries.GetEnrollmentPaymentsByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetEnrollmentPaymentsByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.EnrollmentPayment{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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
				BalanceBonus:      spec.BalanceBonus,
				CourseFeeValue:    spec.CourseFeeValue,
				TransportFeeValue: spec.TransportFeeValue,
				PenaltyFeeValue:   spec.PenaltyFeeValue,
				DiscountFeeValue:  spec.DiscountFeeValue,
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
				BalanceBonus:      spec.BalanceBonus,
				CourseFeeValue:    spec.CourseFeeValue,
				TransportFeeValue: spec.TransportFeeValue,
				PenaltyFeeValue:   spec.PenaltyFeeValue,
				DiscountFeeValue:  spec.DiscountFeeValue,
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteEnrollmentPaymentsByIds(newCtx, enrollmentPaymentIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteEnrollmentPaymentByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetStudentLearningTokens(ctx context.Context, pagination util.PaginationSpec) (entity.GetStudentLearningTokensResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	var studentLearningTokenRows = make([]mysql.GetStudentLearningTokensRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentLearningTokenRows, err = s.mySQLQueries.GetStudentLearningTokens(newCtx, mysql.GetStudentLearningTokensParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentLearningTokens(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountStudentLearningTokens(newCtx)
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountStudentLearningTokens(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetStudentLearningTokensResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	studentLearningTokens := NewStudentLearningTokensFromGetStudentLearningTokensRow(studentLearningTokenRows)

	return entity.GetStudentLearningTokensResult{
		StudentLearningTokens: studentLearningTokens,
		PaginationResult:      *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetStudentLearningTokenById(ctx context.Context, id entity.StudentLearningTokenID) (entity.StudentLearningToken, error) {
	var studentLearningTokenRow mysql.GetStudentLearningTokenByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentLearningTokenRow, err = s.mySQLQueries.GetStudentLearningTokenById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentLearningTokenById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.StudentLearningToken{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	studentLearningToken := NewStudentLearningTokensFromGetStudentLearningTokensRow([]mysql.GetStudentLearningTokensRow{studentLearningTokenRow.ToGetStudentLearningTokensRow()})[0]

	return studentLearningToken, nil
}

func (s entityServiceImpl) GetStudentLearningTokensByIds(ctx context.Context, ids []entity.StudentLearningTokenID) ([]entity.StudentLearningToken, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var studentLearningTokenRows = make([]mysql.GetStudentLearningTokensByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		studentLearningTokenRows, err = s.mySQLQueries.GetStudentLearningTokensByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetStudentLearningTokensByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.StudentLearningToken{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
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
				Quota:                    spec.Quota,
				CourseFeeQuarterValue:    spec.CourseFeeQuarterValue,
				TransportFeeQuarterValue: spec.TransportFeeQuarterValue,
				CreatedAt:                time.Now().UTC(),
				LastUpdatedAt:            time.Now().UTC(),
				EnrollmentID:             int64(spec.StudentEnrollmentID),
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
				Quota:                    spec.Quota,
				CourseFeeQuarterValue:    spec.CourseFeeQuarterValue,
				TransportFeeQuarterValue: spec.TransportFeeQuarterValue,
				LastUpdatedAt:            time.Now().UTC(),
				ID:                       int64(spec.StudentLearningTokenID),
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

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteStudentLearningTokensByIds(newCtx, studentLearningTokenIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteStudentLearningTokenByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetAttendances(ctx context.Context, pagination util.PaginationSpec, spec entity.GetAttendancesSpec, sortRecent bool) (entity.GetAttendancesResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	timeFilter := spec.TimeSpec
	timeFilter.SetDefaultForZeroValues()

	classID := int64(spec.ClassID)
	useClassFilter := spec.ClassID != entity.ClassID_None

	studentID := int64(spec.StudentID)
	useStudentFilter := spec.StudentID != entity.StudentID_None

	useUnpaidFilter := spec.UnpaidOnly

	var attendanceRows = make([]mysql.GetAttendancesRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		if !sortRecent { // TODO: find alternative: currently sqlc dynamic query is so bad that we need to do this :(
			attendanceRows, err = s.mySQLQueries.GetAttendances(newCtx, mysql.GetAttendancesParams{
				StartDate:        timeFilter.StartDatetime,
				EndDate:          timeFilter.EndDatetime,
				ClassID:          classID,
				UseClassFilter:   useClassFilter,
				StudentID:        studentID,
				UseStudentFilter: useStudentFilter,
				UseUnpaidFilter:  useUnpaidFilter,
				Limit:            int32(limit),
				Offset:           int32(offset),
			})
			if err != nil {
				return fmt.Errorf("mySQLQueries.GetAttendances(): %w", err)
			}
		} else {
			attendanceDescendingdateRows, err := s.mySQLQueries.GetAttendancesDescendingDate(newCtx, mysql.GetAttendancesDescendingDateParams{
				StartDate:        timeFilter.StartDatetime,
				EndDate:          timeFilter.EndDatetime,
				ClassID:          classID,
				UseClassFilter:   useClassFilter,
				StudentID:        studentID,
				UseStudentFilter: useStudentFilter,
				UseUnpaidFilter:  useUnpaidFilter,
				Limit:            int32(limit),
				Offset:           int32(offset),
			})
			if err != nil {
				return fmt.Errorf("mySQLQueries.GetAttendancesDescendingDate(): %w", err)
			}

			for _, row := range attendanceDescendingdateRows {
				attendanceRows = append(attendanceRows, row.ToGetAttendancesRow())
			}
		}

		totalResults, err = s.mySQLQueries.CountAttendances(newCtx, mysql.CountAttendancesParams{
			StartDate:        timeFilter.StartDatetime,
			EndDate:          timeFilter.EndDatetime,
			ClassID:          classID,
			UseClassFilter:   useClassFilter,
			StudentID:        studentID,
			UseStudentFilter: useStudentFilter,
			UseUnpaidFilter:  useUnpaidFilter,
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountAttendances(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetAttendancesResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}
	attendances := NewAttendancesFromGetAttendancesRow(attendanceRows)

	return entity.GetAttendancesResult{
		Attendances:      attendances,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetUnpaidAttendancesByTeacherId(ctx context.Context, spec entity.GetUnpaidAttendancesByTeacherIdSpec) ([]entity.Attendance, error) {
	timeFilter := spec.TimeSpec
	timeFilter.SetDefaultForZeroValues()

	var attendanceRows = make([]mysql.GetUnpaidAttendancesByTeacherIdRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		attendanceRows, err = s.mySQLQueries.GetUnpaidAttendancesByTeacherId(newCtx, mysql.GetUnpaidAttendancesByTeacherIdParams{
			StartDate: timeFilter.StartDatetime,
			EndDate:   timeFilter.EndDatetime,
			TeacherID: int64(spec.TeacherID),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetUnpaidAttendancesByTeacherId(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.Attendance{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	attendanceRowsConverted := make([]mysql.GetAttendancesRow, 0, len(attendanceRows))
	for _, row := range attendanceRows {
		attendanceRowsConverted = append(attendanceRowsConverted, row.ToGetAttendancesRow())
	}

	attendances := NewAttendancesFromGetAttendancesRow(attendanceRowsConverted)

	return attendances, nil
}

func (s entityServiceImpl) GetAttendanceById(ctx context.Context, id entity.AttendanceID) (entity.Attendance, error) {
	var attendanceRow mysql.GetAttendanceByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		attendanceRow, err = s.mySQLQueries.GetAttendanceById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetAttendanceById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.Attendance{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	attendance := NewAttendancesFromGetAttendancesRow([]mysql.GetAttendancesRow{attendanceRow.ToGetAttendancesRow()})[0]

	return attendance, nil
}

func (s entityServiceImpl) GetAttendancesByIds(ctx context.Context, ids []entity.AttendanceID) ([]entity.Attendance, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var attendanceRows = make([]mysql.GetAttendancesByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		attendanceRows, err = s.mySQLQueries.GetAttendancesByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetAttendancesByIds(): %w", err)
		}

		return nil
	})
	if err != nil {
		return []entity.Attendance{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	attendanceRowsConverted := make([]mysql.GetAttendancesRow, 0, len(attendanceRows))
	for _, row := range attendanceRows {
		attendanceRowsConverted = append(attendanceRowsConverted, row.ToGetAttendancesRow())
	}

	attendances := NewAttendancesFromGetAttendancesRow(attendanceRowsConverted)

	return attendances, nil
}

func (s entityServiceImpl) InsertAttendances(ctx context.Context, specs []entity.InsertAttendanceSpec) ([]entity.AttendanceID, error) {
	attendanceIDs := make([]entity.AttendanceID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			attendanceID, err := qtx.InsertAttendance(newCtx, mysql.InsertAttendanceParams{
				Date:                  spec.Date,
				UsedStudentTokenQuota: spec.UsedStudentTokenQuota,
				Duration:              spec.Duration,
				Note:                  spec.Note,
				ClassID:               int64(spec.ClassID),
				TeacherID:             int64(spec.TeacherID),
				StudentID:             int64(spec.StudentID),
				TokenID: sql.NullInt64{
					Int64: int64(spec.StudentLearningTokenID),
					Valid: spec.StudentLearningTokenID != entity.StudentLearningTokenID_None,
				},
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertAttendance(): %w", err)
			}
			attendanceIDs = append(attendanceIDs, entity.AttendanceID(attendanceID))
		}
		return nil
	})
	if err != nil {
		return []entity.AttendanceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}
	return attendanceIDs, nil
}

func (s entityServiceImpl) UpdateAttendances(ctx context.Context, specs []entity.UpdateAttendanceSpec) ([]entity.AttendanceID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountAttendancesByIds)
	if errV != nil {
		return []entity.AttendanceID{}, errV
	}

	attendanceIDs := make([]entity.AttendanceID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateAttendance(newCtx, mysql.UpdateAttendanceParams{
				Date:                  spec.Date,
				UsedStudentTokenQuota: spec.UsedStudentTokenQuota,
				Duration:              spec.Duration,
				Note:                  spec.Note,
				IsPaid:                util.BoolToInt32(spec.IsPaid),
				ClassID:               int64(spec.ClassID),
				TeacherID:             int64(spec.TeacherID),
				StudentID:             int64(spec.StudentID),
				TokenID:               sql.NullInt64{Int64: int64(spec.StudentLearningTokenID), Valid: true},
				ID:                    int64(spec.AttendanceID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateAttendance(): %w", err)
			}
			attendanceIDs = append(attendanceIDs, spec.AttendanceID)
		}
		return nil
	})
	if err != nil {
		return []entity.AttendanceID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return attendanceIDs, nil
}

func (s entityServiceImpl) DeleteAttendances(ctx context.Context, ids []entity.AttendanceID) error {
	attendanceIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		attendanceIdsInt64 = append(attendanceIdsInt64, int64(id))
	}

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteAttendancesByIds(newCtx, attendanceIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteAttendanceByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}

func (s entityServiceImpl) GetTeacherPayments(ctx context.Context, pagination util.PaginationSpec, spec entity.GetTeacherPaymentsSpec) (entity.GetTeacherPaymentsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	timeFilter := spec.TimeSpec
	timeFilter.SetDefaultForZeroValues()

	teacherID := int64(spec.TeacherID)
	useTeacherFilter := spec.TeacherID != entity.TeacherID_None

	var teacherPaymentRows = make([]mysql.GetTeacherPaymentsRow, 0)
	var totalResults int64 = 0
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherPaymentRows, err = s.mySQLQueries.GetTeacherPayments(newCtx, mysql.GetTeacherPaymentsParams{
			StartDate:        timeFilter.StartDatetime,
			EndDate:          timeFilter.EndDatetime,
			TeacherID:        teacherID,
			UseTeacherFilter: useTeacherFilter,
			Limit:            int32(limit),
			Offset:           int32(offset),
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherPayments(): %w", err)
		}

		totalResults, err = s.mySQLQueries.CountTeacherPayments(newCtx, mysql.CountTeacherPaymentsParams{
			TeacherID:        teacherID,
			UseTeacherFilter: useTeacherFilter,
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.CountTeacherPayments(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.GetTeacherPaymentsResult{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teacherPayments := NewTeacherPaymentsFromGetTeacherPaymentsRow(teacherPaymentRows)

	return entity.GetTeacherPaymentsResult{
		TeacherPayments:  teacherPayments,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s entityServiceImpl) GetTeacherPaymentsByTeacherId(ctx context.Context, spec entity.GetTeacherPaymentsByTeacherIdSpec) ([]entity.TeacherPayment, error) {
	timeFilter := spec.TimeSpec
	timeFilter.SetDefaultForZeroValues()

	teacherID := int64(spec.TeacherID)

	var teacherPaymentRows = make([]mysql.GetTeacherPaymentsByTeacherIdRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherPaymentRows, err = s.mySQLQueries.GetTeacherPaymentsByTeacherId(newCtx, mysql.GetTeacherPaymentsByTeacherIdParams{
			StartDate: timeFilter.StartDatetime,
			EndDate:   timeFilter.EndDatetime,
			TeacherID: teacherID,
		})
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherPaymentsByTeacherId(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherPayment{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teacherPaymentRowsConverted := make([]mysql.GetTeacherPaymentsRow, 0, len(teacherPaymentRows))
	for _, row := range teacherPaymentRows {
		teacherPaymentRowsConverted = append(teacherPaymentRowsConverted, row.ToGetTeacherPaymentsRow())
	}

	teacherPayments := NewTeacherPaymentsFromGetTeacherPaymentsRow(teacherPaymentRowsConverted)

	return teacherPayments, nil
}

func (s entityServiceImpl) GetTeacherPaymentById(ctx context.Context, id entity.TeacherPaymentID) (entity.TeacherPayment, error) {
	var teacherPaymentRow mysql.GetTeacherPaymentByIdRow
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherPaymentRow, err = s.mySQLQueries.GetTeacherPaymentById(newCtx, int64(id))
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherPaymentById(): %w", err)
		}
		return nil
	})
	if err != nil {
		return entity.TeacherPayment{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teacherPayment := NewTeacherPaymentsFromGetTeacherPaymentsRow([]mysql.GetTeacherPaymentsRow{teacherPaymentRow.ToGetTeacherPaymentsRow()})[0]

	return teacherPayment, nil
}

func (s entityServiceImpl) GetTeacherPaymentsByIds(ctx context.Context, ids []entity.TeacherPaymentID) ([]entity.TeacherPayment, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	var teacherPaymentRows = make([]mysql.GetTeacherPaymentsByIdsRow, 0)
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		var err error
		teacherPaymentRows, err = s.mySQLQueries.GetTeacherPaymentsByIds(newCtx, idsInt)
		if err != nil {
			return fmt.Errorf("mySQLQueries.GetTeacherPaymentsByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherPayment{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	teacherPaymentRowsConverted := make([]mysql.GetTeacherPaymentsRow, 0, len(teacherPaymentRows))
	for _, row := range teacherPaymentRows {
		teacherPaymentRowsConverted = append(teacherPaymentRowsConverted, row.ToGetTeacherPaymentsRow())
	}

	teacherPayments := NewTeacherPaymentsFromGetTeacherPaymentsRow(teacherPaymentRowsConverted)

	return teacherPayments, nil
}

func (s entityServiceImpl) InsertTeacherPayments(ctx context.Context, specs []entity.InsertTeacherPaymentSpec) ([]entity.TeacherPaymentID, error) {
	teacherPaymentIDs := make([]entity.TeacherPaymentID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			teacherPaymentID, err := qtx.InsertTeacherPayment(newCtx, mysql.InsertTeacherPaymentParams{
				AttendanceID:          int64(spec.AttendanceID),
				PaidCourseFeeValue:    spec.PaidCourseFeeValue,
				PaidTransportFeeValue: spec.PaidTransportFeeValue,
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertTeacherPayment(): %w", err)
			}
			teacherPaymentIDs = append(teacherPaymentIDs, entity.TeacherPaymentID(teacherPaymentID))
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherPaymentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teacherPaymentIDs, nil
}

func (s entityServiceImpl) UpdateTeacherPayments(ctx context.Context, specs []entity.UpdateTeacherPaymentSpec) ([]entity.TeacherPaymentID, error) {
	errV := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountTeacherPaymentsByIds)
	if errV != nil {
		return []entity.TeacherPaymentID{}, errV
	}

	teacherPaymentIDs := make([]entity.TeacherPaymentID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			err := qtx.UpdateTeacherPayment(newCtx, mysql.UpdateTeacherPaymentParams{
				AttendanceID:          int64(spec.AttendanceID),
				PaidCourseFeeValue:    spec.PaidCourseFeeValue,
				PaidTransportFeeValue: spec.PaidTransportFeeValue,
				AddedAt:               spec.AddedAt,
				ID:                    int64(spec.TeacherPaymentID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateTeacherPayment(): %w", err)
			}
			teacherPaymentIDs = append(teacherPaymentIDs, spec.TeacherPaymentID)
		}
		return nil
	})
	if err != nil {
		return []entity.TeacherPaymentID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return teacherPaymentIDs, nil
}

func (s entityServiceImpl) DeleteTeacherPayments(ctx context.Context, ids []entity.TeacherPaymentID) error {
	teacherPaymentIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		teacherPaymentIdsInt64 = append(teacherPaymentIdsInt64, int64(id))
	}

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		err := s.mySQLQueries.DeleteTeacherPaymentsByIds(newCtx, teacherPaymentIdsInt64)
		if err != nil {
			return fmt.Errorf("mySQLQueries.DeleteTeacherPaymentsByIds(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return nil
}
