// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: user_queries.sql

package mysql

import (
	"context"
	"encoding/json"
)

const activateUser = `-- name: ActivateUser :exec
UPDATE user SET is_deactivated = 0 WHERE id = ?
`

func (q *Queries) ActivateUser(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, activateUser, id)
	return err
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

const getUser = `-- name: GetUser :many
SELECT id, username, email, user_detail, privilege_type, is_deactivated, created_at FROM user
ORDER BY name
`

func (q *Queries) GetUser(ctx context.Context) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, getUser)
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

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, user_detail, privilege_type, is_deactivated, created_at FROM user
WHERE email = ? LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
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
SELECT user_id, email, password FROM user_credential WHERE email = ?
`

func (q *Queries) GetUserCredentialByEmail(ctx context.Context, email string) (UserCredential, error) {
	row := q.db.QueryRowContext(ctx, getUserCredentialByEmail, email)
	var i UserCredential
	err := row.Scan(&i.UserID, &i.Email, &i.Password)
	return i, err
}

const getUserCredentialById = `-- name: GetUserCredentialById :one
SELECT user_id, email, password FROM user_credential WHERE user_id = ?
`

// ============================== USER_CREDENTIAL ==============================
func (q *Queries) GetUserCredentialById(ctx context.Context, userID int64) (UserCredential, error) {
	row := q.db.QueryRowContext(ctx, getUserCredentialById, userID)
	var i UserCredential
	err := row.Scan(&i.UserID, &i.Email, &i.Password)
	return i, err
}

const insertUser = `-- name: InsertUser :execlastid
INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  ?, ?, ?, ?
)
`

type InsertUserParams struct {
	Email         string
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
  user_id, email, password
) VALUES (
  ?, ?, ?
)
`

type InsertUserCredentialParams struct {
	UserID   int64
	Email    string
	Password string
}

func (q *Queries) InsertUserCredential(ctx context.Context, arg InsertUserCredentialParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertUserCredential, arg.UserID, arg.Email, arg.Password)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const isUserExist = `-- name: IsUserExist :one
SELECT EXISTS(SELECT id FROM user WHERE email = ? LIMIT 1)
`

func (q *Queries) IsUserExist(ctx context.Context, email string) (bool, error) {
	row := q.db.QueryRowContext(ctx, isUserExist, email)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const updateEmailByUserId = `-- name: UpdateEmailByUserId :exec
UPDATE user_credential SET email = ? WHERE user_id = ?
`

type UpdateEmailByUserIdParams struct {
	Email  string
	UserID int64
}

func (q *Queries) UpdateEmailByUserId(ctx context.Context, arg UpdateEmailByUserIdParams) error {
	_, err := q.db.ExecContext(ctx, updateEmailByUserId, arg.Email, arg.UserID)
	return err
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

const updateUserInfo = `-- name: UpdateUserInfo :exec
UPDATE user SET email = ?, username = ?, user_detail = ? WHERE id = ?
`

type UpdateUserInfoParams struct {
	Email      string
	Username   string
	UserDetail json.RawMessage
	ID         int64
}

func (q *Queries) UpdateUserInfo(ctx context.Context, arg UpdateUserInfoParams) error {
	_, err := q.db.ExecContext(ctx, updateUserInfo,
		arg.Email,
		arg.Username,
		arg.UserDetail,
		arg.ID,
	)
	return err
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
