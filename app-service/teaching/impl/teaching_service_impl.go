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
)

type teachingServiceImpl struct {
	mySQLQueries *relational_db.MySQLQueries
}

var _ teaching.TeachingService = (*teachingServiceImpl)(nil)

func NewTeachingServiceImpl(mySQLQueries *relational_db.MySQLQueries) *teachingServiceImpl {
	return &teachingServiceImpl{
		mySQLQueries: mySQLQueries,
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

func (s teachingServiceImpl) GetTeacherByUserID(ctx context.Context, userID identity.UserID) (teaching.Teacher, error) {
	teacher, err := s.mySQLQueries.GetTeacherByUserId(ctx, int64(userID))
	if err != nil {
		return teaching.Teacher{}, fmt.Errorf("mySQLQueries.GetUserById(): %w", err)
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

	return teaching.GetStudentsResult{
		Students:         students,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s teachingServiceImpl) GetStudentByUserID(ctx context.Context, userID identity.UserID) (teaching.Student, error) {
	student, err := s.mySQLQueries.GetStudentByUserId(ctx, int64(userID))
	if err != nil {
		return teaching.Student{}, fmt.Errorf("mySQLQueries.GetUserById(): %w", err)
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
