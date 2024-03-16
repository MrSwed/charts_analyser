package service

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"database/sql"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

func NewUserService(r *repository.Repository, log *zap.Logger) *UserService {
	validate := validator.New()

	err := validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		if len(v) < constant.PasswordMinLen || len(v) > constant.PasswordMaxLen {
			return false
		}
		for _, rule := range constant.PasswordCheckRules {
			if t, _ := regexp.MatchString(rule, v); !t {
				return false
			}
		}
		return true

	})
	if err != nil {
		log.Error("RegisterValidation", zap.Error(err))
	}

	return &UserService{r: r, validate: validate}
}

type UserService struct {
	r        *repository.Repository
	validate *validator.Validate
}

func (s *UserService) GetUser(ctx context.Context, userID domain.UserID) (user domain.User, err error) {
	user, err = s.r.User.GetUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		err = myErr.ErrNotExist
	}
	return
}

func (s *UserService) AddUser(ctx context.Context, user domain.UserChange) (id domain.UserID, err error) {
	if err = s.validate.Struct(user); err != nil {
		return
	}
	var hash []byte
	if hash, err = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); err != nil {
		return
	}
	userDB := domain.UserDB{
		User: domain.User{Login: user.Login},
		Hash: hash,
		Role: user.Role,
	}

	if id, err = s.r.User.AddUser(ctx, userDB); err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			err = myErr.ErrDuplicateRecord
		}
	}

	return
}

func (s *UserService) UpdateUser(ctx context.Context, user domain.UserChange) (err error) {
	if err = s.validate.VarCtx(ctx, user.ID, "required,gt=0"); err != nil {
		return fmt.Errorf("validation for 'id' failed%w", validator.ValidationErrors{})
	}
	if err = s.validate.Struct(user); err != nil {
		return
	}
	var hash []byte
	if hash, err = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); err != nil {
		return
	}
	userDB := domain.UserDB{
		User: domain.User{Login: user.Login, ID: *user.ID},
		Hash: hash,
		Role: user.Role,
	}
	if err = s.r.User.UpdateUser(ctx, userDB); err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			err = myErr.ErrDuplicateRecord
		}
	}
	return

}

func (s *UserService) SetDeletedUser(ctx context.Context, delete bool, userIDs ...domain.UserID) (err error) {
	return s.r.User.SetDeletedUser(ctx, delete, userIDs...)
}
