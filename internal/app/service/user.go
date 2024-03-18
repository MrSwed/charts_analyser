package service

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"database/sql"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"regexp"
)

func NewUserService(r *repository.Repository, conf *config.JWT, log *zap.Logger) *UserService {
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

	return &UserService{r: r, validate: validate, conf: conf}
}

type UserService struct {
	r        *repository.Repository
	validate *validator.Validate
	conf     *config.JWT
}

func (s *UserService) Login(ctx context.Context, userLogin domain.LoginForm) (token string, err error) {
	var user *domain.UserDB
	if user, err = s.r.User.GetUserByLogin(ctx, userLogin.Login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrLogin
		}
		return
	}
	if !user.Hash.IsValidPassword(userLogin.Password) {
		err = myErr.ErrLogin
		return
	}
	token, err = domain.NewClaimUser(s.conf, user.ID, user.Login, user.Role).Token()
	return
}

func (s *UserService) GetUser(ctx context.Context, login domain.UserLogin) (user *domain.UserDB, err error) {
	user, err = s.r.User.GetUserByLogin(ctx, login)
	if errors.Is(err, sql.ErrNoRows) {
		err = myErr.ErrNotExist
	}
	return
}

func (s *UserService) AddUser(ctx context.Context, user *domain.UserChange) (id domain.UserID, err error) {
	if err = s.validate.Struct(user); err != nil {
		return
	}
	var userDB *domain.UserDB
	if userDB, err = domain.NewUserDB(0, user.Login, user.Password, user.Role); err != nil {
		return
	}
	if id, err = s.r.User.AddUser(ctx, userDB); err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			err = myErr.ErrDuplicateRecord
		} else if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == "23505" {
			err = myErr.ErrDuplicateRecord
		}
	}

	return
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.UserChange) (err error) {
	if err = s.validate.VarCtx(ctx, user.ID, "required,gt=0"); err != nil {
		return fmt.Errorf("field 'id' required%w", validator.ValidationErrors{})
	}
	if err = s.validate.Struct(user); err != nil {
		return
	}
	var userDB *domain.UserDB
	if userDB, err = domain.NewUserDB(*user.ID, user.Login, user.Password, user.Role); err != nil {
		return
	}

	if err = s.r.User.UpdateUser(ctx, userDB); err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" {
			err = myErr.ErrDuplicateRecord
		} else if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
	}
	return
}

func (s *UserService) SetDeletedUser(ctx context.Context, delete bool, userIDs ...domain.UserID) (err error) {
	return s.r.User.SetDeletedUser(ctx, delete, userIDs...)
}
