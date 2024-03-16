package handler

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// Login
// @Tags        User
// @Summary     Идентификация
// @Description Получение токена
// @Accept      json
// @Produce     json
// @Param       UserAuth   body     domain.LoginForm    true "Логин, пароль"
// @Success     200        {string} string "JWT token"
// @Failure     400
// @Failure     401
// @Failure     403
// @Failure     500
// @Router      /login [post]
// @Security    BearerAuth
func (h *Handler) Login() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			login domain.LoginForm
		)
		err = c.BodyParser(&login)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.User.Login(ctx, login)
		if err != nil {
			if errors.Is(err, myErr.ErrLogin) {
				_, err = c.Status(http.StatusUnauthorized).WriteString(err.Error())
				return
			}

			c.Status(http.StatusInternalServerError)
			h.log.Error("Error add users", zap.Error(err), zap.Any("login", login))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString(result)
		return
	}
}

// AddUser
// @Tags        User
// @Summary     Добавление оператора
// @Description
// @Accept      json
// @Produce     json
// @Param       UserData   body     domain.UserChange    true "данные пользователя"
// @Success     201        {integer} domain.UserID
// @Failure     400
// @Failure     401
// @Failure     403
// @Failure     500
// @Router      /user [post]
// @Security    BearerAuth
func (h *Handler) AddUser() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			user domain.UserChange
		)
		err = c.BodyParser(&user)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.User.AddUser(ctx, user)
		if err != nil {
			if errors.Is(err, myErr.ErrDuplicateRecord) {
				_, err = c.Status(http.StatusConflict).WriteString(err.Error())
				return
			}
			if errors.As(err, &validator.ValidationErrors{}) {
				_, err = c.Status(http.StatusBadRequest).WriteString(err.Error())
				return
			}

			c.Status(http.StatusInternalServerError)
			h.log.Error("Error add users", zap.Error(err), zap.Any("user", user))
			return nil
		}
		return c.Status(http.StatusCreated).JSON(result)
	}
}

// UpdateUser
// @Tags        User
// @Summary     Изменение оператора
// @Description Смена названия оператора, для не удаленных
// @Accept      json
// @Produce     json
// @Param       UserData   body     domain.UserChange    true "данные пользователя"
// @Success     200           {string} string "Ok"
// @Failure     400
// @Failure     401
// @Failure     403
// @Failure     500
// @Router      /user [put]
// @Security    BearerAuth
func (h *Handler) UpdateUser() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			user domain.UserChange
		)
		err = c.BodyParser(&user)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.User.UpdateUser(ctx, user)
		if err != nil {
			if errors.Is(err, myErr.ErrDuplicateRecord) {
				_, err = c.Status(http.StatusConflict).WriteString(err.Error())
				return
			}
			if errors.As(err, &validator.ValidationErrors{}) {
				_, err = c.Status(http.StatusBadRequest).WriteString(err.Error())
				return
			}
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error add user", zap.Error(err), zap.Any("user", user))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString("Ok")
		return
	}
}

// DeleteUsers
// @Tags        User
// @Summary     Удаление операторов
// @Description
// @Accept      json
// @Produce     json
// @Param       UserNames   body     []domain.UserID    true "список ID операторов"
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     401
// @Failure     403
// @Failure     500
// @Router      /user [delete]
// @Security    BearerAuth
func (h *Handler) DeleteUsers() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			UserIDs []domain.UserID
		)
		err = c.BodyParser(&UserIDs)
		if err != nil && !errors.Is(err, io.EOF) || len(UserIDs) == 0 {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.User.SetDeletedUser(ctx, true, UserIDs...)
		if err != nil && !errors.Is(err, myErr.ErrNotExist) {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error delete users", zap.Error(err), zap.Any("ids", UserIDs))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString("Ok")
		return
	}
}

// RestoreUsers
// @Tags        User
// @Summary     Восстановление оператора
// @Description
// @Accept      json
// @Produce     json
// @Param       UserNames   body     []domain.UserID    true "список ID операторов"
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     401
// @Failure     403
// @Failure     500
// @Router      /user [patch]
// @Security    BearerAuth
func (h *Handler) RestoreUsers() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			UserIDs []domain.UserID
		)
		err = c.BodyParser(&UserIDs)
		if err != nil && !errors.Is(err, io.EOF) || len(UserIDs) == 0 {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.User.SetDeletedUser(ctx, false, UserIDs...)
		if err != nil && !errors.Is(err, myErr.ErrNotExist) {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error restore users", zap.Error(err), zap.Any("ids", UserIDs))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString("Ok")
		return
	}
}
