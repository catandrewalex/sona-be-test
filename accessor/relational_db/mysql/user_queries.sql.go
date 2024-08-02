// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user_queries.sql

package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
)

const activateUser = `-- name: ActivateUser :exec
UPDATE user SET is_deactivated = 0 WHERE id = ?
`

func (q *Queries) ActivateUser(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, activateUser, id)
	return err
}

const countUsers = `-- name: CountUsers :one
SELECT Count(*) AS total FROM user
WHERE is_deactivated IN (/*SLICE:isDeactivateds*/?)
`

func (q *Queries) CountUsers(ctx context.Context, isdeactivateds []int32) (int64, error) {
	query := countUsers
	var queryParams []interface{}
	if len(isdeactivateds) > 0 {
		for _, v := range isdeactivateds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", strings.Repeat(",?", len(isdeactivateds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", "NULL", 1)
	}
	row := q.db.QueryRowContext(ctx, query, queryParams...)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const countUsersByIds = `-- name: CountUsersByIds :one
SELECT Count(*) AS total FROM user
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) CountUsersByIds(ctx context.Context, ids []int64) (int64, error) {
	query := countUsersByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	row := q.db.QueryRowContext(ctx, query, queryParams...)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const countUsersNotStudent = `-- name: CountUsersNotStudent :one
SELECT Count(*) AS total FROM user
LEFT JOIN student on user.id = student.user_id
WHERE is_deactivated IN (/*SLICE:isDeactivateds*/?) AND student.user_id IS NULL
`

func (q *Queries) CountUsersNotStudent(ctx context.Context, isdeactivateds []int32) (int64, error) {
	query := countUsersNotStudent
	var queryParams []interface{}
	if len(isdeactivateds) > 0 {
		for _, v := range isdeactivateds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", strings.Repeat(",?", len(isdeactivateds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", "NULL", 1)
	}
	row := q.db.QueryRowContext(ctx, query, queryParams...)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const countUsersNotTeacher = `-- name: CountUsersNotTeacher :one
SELECT Count(*) AS total FROM user
LEFT JOIN teacher on user.id = teacher.user_id
WHERE is_deactivated IN (/*SLICE:isDeactivateds*/?) AND teacher.user_id IS NULL
`

func (q *Queries) CountUsersNotTeacher(ctx context.Context, isdeactivateds []int32) (int64, error) {
	query := countUsersNotTeacher
	var queryParams []interface{}
	if len(isdeactivateds) > 0 {
		for _, v := range isdeactivateds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", strings.Repeat(",?", len(isdeactivateds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", "NULL", 1)
	}
	row := q.db.QueryRowContext(ctx, query, queryParams...)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const deactivateUser = `-- name: DeactivateUser :exec
UPDATE user SET is_deactivated = 1 WHERE id = ?
`

func (q *Queries) DeactivateUser(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deactivateUser, id)
	return err
}

const deleteUserById = `-- name: DeleteUserById :exec
DELETE FROM user
WHERE id = ?
`

func (q *Queries) DeleteUserById(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteUserById, id)
	return err
}

const deleteUserCredentialByUserId = `-- name: DeleteUserCredentialByUserId :exec
DELETE FROM user_credential
WHERE user_id = ?
`

func (q *Queries) DeleteUserCredentialByUserId(ctx context.Context, userID int64) error {
	_, err := q.db.ExecContext(ctx, deleteUserCredentialByUserId, userID)
	return err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, user_detail, privilege_type, is_deactivated, created_at FROM user
WHERE email = ? LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email sql.NullString) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.UserDetail,
		&i.PrivilegeType,
		&i.IsDeactivated,
		&i.CreatedAt,
	)
	return i, err
}

const getUserById = `-- name: GetUserById :one
SELECT id, username, email, user_detail, privilege_type, is_deactivated, created_at FROM user
WHERE id = ? LIMIT 1
`

// ============================== USER ==============================
func (q *Queries) GetUserById(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.UserDetail,
		&i.PrivilegeType,
		&i.IsDeactivated,
		&i.CreatedAt,
	)
	return i, err
}

const getUserCredentialByEmail = `-- name: GetUserCredentialByEmail :one
SELECT user_id, username, email, password FROM user_credential WHERE email = ?
`

func (q *Queries) GetUserCredentialByEmail(ctx context.Context, email sql.NullString) (UserCredential, error) {
	row := q.db.QueryRowContext(ctx, getUserCredentialByEmail, email)
	var i UserCredential
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.Email,
		&i.Password,
	)
	return i, err
}

const getUserCredentialById = `-- name: GetUserCredentialById :one
SELECT user_id, username, email, password FROM user_credential WHERE user_id = ?
`

// ============================== USER_CREDENTIAL ==============================
func (q *Queries) GetUserCredentialById(ctx context.Context, userID int64) (UserCredential, error) {
	row := q.db.QueryRowContext(ctx, getUserCredentialById, userID)
	var i UserCredential
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.Email,
		&i.Password,
	)
	return i, err
}

const getUserCredentialByUsername = `-- name: GetUserCredentialByUsername :one
SELECT user_id, username, email, password FROM user_credential WHERE username = ?
`

func (q *Queries) GetUserCredentialByUsername(ctx context.Context, username string) (UserCredential, error) {
	row := q.db.QueryRowContext(ctx, getUserCredentialByUsername, username)
	var i UserCredential
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.Email,
		&i.Password,
	)
	return i, err
}

const getUsers = `-- name: GetUsers :many
SELECT id, username, email, user_detail, privilege_type, is_deactivated, created_at FROM user
WHERE is_deactivated IN (/*SLICE:isDeactivateds*/?)
ORDER BY id
LIMIT ? OFFSET ?
`

type GetUsersParams struct {
	IsDeactivateds []int32
	Limit          int32
	Offset         int32
}

func (q *Queries) GetUsers(ctx context.Context, arg GetUsersParams) ([]User, error) {
	query := getUsers
	var queryParams []interface{}
	if len(arg.IsDeactivateds) > 0 {
		for _, v := range arg.IsDeactivateds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", strings.Repeat(",?", len(arg.IsDeactivateds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", "NULL", 1)
	}
	queryParams = append(queryParams, arg.Limit)
	queryParams = append(queryParams, arg.Offset)
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.UserDetail,
			&i.PrivilegeType,
			&i.IsDeactivated,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUsersByIds = `-- name: GetUsersByIds :many
SELECT id, username, email, user_detail, privilege_type, is_deactivated, created_at from user
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) GetUsersByIds(ctx context.Context, ids []int64) ([]User, error) {
	query := getUsersByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.UserDetail,
			&i.PrivilegeType,
			&i.IsDeactivated,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUsersNotStudent = `-- name: GetUsersNotStudent :many
SELECT user.id, user.username, user.email, user.user_detail, user.privilege_type, user.is_deactivated, user.created_at FROM user
LEFT JOIN student on user.id = student.user_id
WHERE is_deactivated IN (/*SLICE:isDeactivateds*/?) AND student.user_id IS NULL
ORDER BY user.id
LIMIT ? OFFSET ?
`

type GetUsersNotStudentParams struct {
	IsDeactivateds []int32
	Limit          int32
	Offset         int32
}

type GetUsersNotStudentRow struct {
	User User
}

func (q *Queries) GetUsersNotStudent(ctx context.Context, arg GetUsersNotStudentParams) ([]GetUsersNotStudentRow, error) {
	query := getUsersNotStudent
	var queryParams []interface{}
	if len(arg.IsDeactivateds) > 0 {
		for _, v := range arg.IsDeactivateds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", strings.Repeat(",?", len(arg.IsDeactivateds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", "NULL", 1)
	}
	queryParams = append(queryParams, arg.Limit)
	queryParams = append(queryParams, arg.Offset)
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUsersNotStudentRow
	for rows.Next() {
		var i GetUsersNotStudentRow
		if err := rows.Scan(
			&i.User.ID,
			&i.User.Username,
			&i.User.Email,
			&i.User.UserDetail,
			&i.User.PrivilegeType,
			&i.User.IsDeactivated,
			&i.User.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUsersNotTeacher = `-- name: GetUsersNotTeacher :many
SELECT user.id, user.username, user.email, user.user_detail, user.privilege_type, user.is_deactivated, user.created_at FROM user
LEFT JOIN teacher on user.id = teacher.user_id
WHERE is_deactivated IN (/*SLICE:isDeactivateds*/?) AND teacher.user_id IS NULL
ORDER BY user.id
LIMIT ? OFFSET ?
`

type GetUsersNotTeacherParams struct {
	IsDeactivateds []int32
	Limit          int32
	Offset         int32
}

type GetUsersNotTeacherRow struct {
	User User
}

func (q *Queries) GetUsersNotTeacher(ctx context.Context, arg GetUsersNotTeacherParams) ([]GetUsersNotTeacherRow, error) {
	query := getUsersNotTeacher
	var queryParams []interface{}
	if len(arg.IsDeactivateds) > 0 {
		for _, v := range arg.IsDeactivateds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", strings.Repeat(",?", len(arg.IsDeactivateds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:isDeactivateds*/?", "NULL", 1)
	}
	queryParams = append(queryParams, arg.Limit)
	queryParams = append(queryParams, arg.Offset)
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUsersNotTeacherRow
	for rows.Next() {
		var i GetUsersNotTeacherRow
		if err := rows.Scan(
			&i.User.ID,
			&i.User.Username,
			&i.User.Email,
			&i.User.UserDetail,
			&i.User.PrivilegeType,
			&i.User.IsDeactivated,
			&i.User.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertUser = `-- name: InsertUser :execlastid
INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  ?, ?, ?, ?
)
`

type InsertUserParams struct {
	Email         sql.NullString
	Username      string
	UserDetail    json.RawMessage
	PrivilegeType int32
}

func (q *Queries) InsertUser(ctx context.Context, arg InsertUserParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertUser,
		arg.Email,
		arg.Username,
		arg.UserDetail,
		arg.PrivilegeType,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const insertUserCredential = `-- name: InsertUserCredential :execlastid
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  ?, ?, ?, ?
)
`

type InsertUserCredentialParams struct {
	UserID   int64
	Username string
	Email    sql.NullString
	Password string
}

func (q *Queries) InsertUserCredential(ctx context.Context, arg InsertUserCredentialParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertUserCredential,
		arg.UserID,
		arg.Username,
		arg.Email,
		arg.Password,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const isUserExist = `-- name: IsUserExist :one
SELECT EXISTS(SELECT id FROM user WHERE email = ? LIMIT 1)
`

func (q *Queries) IsUserExist(ctx context.Context, email sql.NullString) (bool, error) {
	row := q.db.QueryRowContext(ctx, isUserExist, email)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const updatePasswordByUserId = `-- name: UpdatePasswordByUserId :exec
UPDATE user_credential SET password = ? WHERE user_id = ?
`

type UpdatePasswordByUserIdParams struct {
	Password string
	UserID   int64
}

func (q *Queries) UpdatePasswordByUserId(ctx context.Context, arg UpdatePasswordByUserIdParams) error {
	_, err := q.db.ExecContext(ctx, updatePasswordByUserId, arg.Password, arg.UserID)
	return err
}

const updateUser = `-- name: UpdateUser :exec
UPDATE user SET username = ?, email = ?, user_detail = ?, privilege_type = ?, is_deactivated = ? WHERE id = ?
`

type UpdateUserParams struct {
	Username      string
	Email         sql.NullString
	UserDetail    json.RawMessage
	PrivilegeType int32
	IsDeactivated int32
	ID            int64
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.ExecContext(ctx, updateUser,
		arg.Username,
		arg.Email,
		arg.UserDetail,
		arg.PrivilegeType,
		arg.IsDeactivated,
		arg.ID,
	)
	return err
}

const updateUserByUsername = `-- name: UpdateUserByUsername :execrows
UPDATE user SET email = ?, user_detail = ?, privilege_type = ?, is_deactivated = ? WHERE username = ?
`

type UpdateUserByUsernameParams struct {
	Email         sql.NullString
	UserDetail    json.RawMessage
	PrivilegeType int32
	IsDeactivated int32
	Username      string
}

func (q *Queries) UpdateUserByUsername(ctx context.Context, arg UpdateUserByUsernameParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateUserByUsername,
		arg.Email,
		arg.UserDetail,
		arg.PrivilegeType,
		arg.IsDeactivated,
		arg.Username,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const updateUserCredentialInfoByUserId = `-- name: UpdateUserCredentialInfoByUserId :exec
UPDATE user_credential SET username = ?, email = ? WHERE user_id = ?
`

type UpdateUserCredentialInfoByUserIdParams struct {
	Username string
	Email    sql.NullString
	UserID   int64
}

func (q *Queries) UpdateUserCredentialInfoByUserId(ctx context.Context, arg UpdateUserCredentialInfoByUserIdParams) error {
	_, err := q.db.ExecContext(ctx, updateUserCredentialInfoByUserId, arg.Username, arg.Email, arg.UserID)
	return err
}

const updateUserCredentialInfoByUsername = `-- name: UpdateUserCredentialInfoByUsername :execrows
UPDATE user_credential SET email = ? WHERE username = ?
`

type UpdateUserCredentialInfoByUsernameParams struct {
	Email    sql.NullString
	Username string
}

func (q *Queries) UpdateUserCredentialInfoByUsername(ctx context.Context, arg UpdateUserCredentialInfoByUsernameParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateUserCredentialInfoByUsername, arg.Email, arg.Username)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const updateUserPrivilege = `-- name: UpdateUserPrivilege :exec
UPDATE user SET privilege_type = ? WHERE id = ?
`

type UpdateUserPrivilegeParams struct {
	PrivilegeType int32
	ID            int64
}

func (q *Queries) UpdateUserPrivilege(ctx context.Context, arg UpdateUserPrivilegeParams) error {
	_, err := q.db.ExecContext(ctx, updateUserPrivilege, arg.PrivilegeType, arg.ID)
	return err
}
