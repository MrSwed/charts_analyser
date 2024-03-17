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

func (r *UserRepo) GetUser(ctx context.Context, login domain.UserLogin) (user *domain.UserDB, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	user = new(domain.UserDB)
	if sqlStr, args, err = sq.Select("id", "login", "hash", "created_at", "is_deleted", "role").
		From(constant.DBUsers).
		Where("login = $1 and is_deleted is not true", login).
		ToSql(); err != nil {
		return
	}

	err = r.db.GetContext(ctx, user, sqlStr, args...)
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
		}).
		Where(sqrl.Eq{"id": user.ID}).
		Suffix("returning id").
		ToSql(); err != nil {
		return
	}

	var updatedID int
	err = r.db.GetContext(ctx, &updatedID, sqlStr, args...)

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
