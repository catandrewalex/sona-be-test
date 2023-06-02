package impl

import (
	"context"
	"encoding/json"
	"fmt"

	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/teaching"
	"sonamusica-backend/app-service/util"
	"sonamusica-backend/errs"
	"sonamusica-backend/network"
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
	tx, err := s.mySQLQueries.DB.Begin()
	if err != nil {
		return []teaching.TeacherID{}, fmt.Errorf("mySQLDB.Begin(): %w", err)
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
	tx, err := s.mySQLQueries.DB.Begin()
	if err != nil {
		return []teaching.StudentID{}, fmt.Errorf("mySQLDB.Begin(): %w", err)
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
