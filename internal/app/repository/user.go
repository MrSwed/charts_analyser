package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"time"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetUser(ctx context.Context, userID domain.UserID) (user domain.User, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("id", "login", "hash", "created_it", "is_deleted", "role").
		From(constant.DBUsers).
		Where("id = $1 and is_deleted is not true", userID).
		ToSql(); err != nil {
		return
	}

	err = r.db.GetContext(ctx, &user, sqlStr, args...)
	return
}

func (r *UserRepo) AddUser(ctx context.Context, user domain.UserDB) (id domain.UserID, err error) {
	var (
		sqlStr string
		args   []interface{}
	)

	createdAt := time.Now()
	sqBuild := sq.Insert(constant.DBUsers).
		Columns("login",
			"role",
			"hash",
			"created_at",
			"modified_at",
		).
		Values(
			user.Login,
			user.Role,
			user.Hash,
			createdAt,
			createdAt,
		).
		Suffix(" returning id ")

	if sqlStr, args, err = sqBuild.ToSql(); err != nil {
		return
	}

	err = r.db.GetContext(ctx, &id, sqlStr, args...)
	return
}

func (r *UserRepo) UpdateUser(ctx context.Context, user domain.UserDB) (err error) {
	var (
		sqlStr string
		args   []interface{}
	)

	modifiedAt := time.Now()
	if sqlStr, args, err = sq.Update(constant.DBUsers).
		SetMap(map[string]interface{}{
			"login":       user.Login,
			"role":        user.Role,
			"hash":        user.Hash,
			"modified_at": modifiedAt,
		}).ToSql(); err != nil {
		return
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return
}

func (r *UserRepo) SetDeletedUser(ctx context.Context, delete bool, userIDs ...domain.UserID) (err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Update(constant.DBUsers).
		Set("is_deleted", delete).
		Where(sqrl.Eq{"id": userIDs}).
		ToSql(); err != nil {
		return
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return
}
