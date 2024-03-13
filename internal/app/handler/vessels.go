package handler

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// AddVessel
// @Tags        Vessel
// @Summary     Добавление судна
// @Description
// @Accept      json
// @Produce     json
// @Param       VesselNames   body     []domain.VesselName    true "список названий Судов"
// @Success     201           {object} []domain.Vessel
// @Failure     400
// @Failure     500
// @Router      /vessels [post]
// @Security    BearerAuth
func (h *Handler) AddVessel() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			VesselNames []domain.VesselName
		)
		err = c.BodyParser(&VesselNames)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.Vessel.AddVessel(ctx, VesselNames...)
		if err != nil && !errors.Is(err, myErr.ErrNotExist) {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error add vessels", zap.Error(err), zap.Any("Names", VesselNames))
			return nil
		}
		return c.Status(http.StatusCreated).JSON(result)
	}
}

// GetVessel
// @Tags        Vessel
// @Summary     Добавление судна
// @Description
// @Accept      json
// @Produce     json
// @Param       VesselNames   body     []domain.VesselIDs    true "список ID Судов"
// @Success     200           {object} []domain.Vessel
// @Failure     400
// @Failure     500
// @Router      /vessels [get]
// @Security    BearerAuth
func (h *Handler) GetVessel() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			VesselIDs []domain.VesselID
		)
		err = c.BodyParser(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.Vessel.GetVessels(ctx, VesselIDs...)
		if err != nil && !errors.Is(err, myErr.ErrNotExist) {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error get vessels", zap.Error(err), zap.Any("Names", VesselIDs))
			return nil
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// DeleteVessel
// @Tags        Vessel
// @Summary     Удаление судна
// @Description
// @Accept      json
// @Produce     json
// @Param       VesselNames   body     []domain.VesselIDs    true "список ID Судов"
// @Success     200           {string} string "Ok"
// @Failure     400
// @Failure     500
// @Router      /vessels [delete]
// @Security    BearerAuth
func (h *Handler) DeleteVessel() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			VesselIDs []domain.VesselID
		)
		err = c.BodyParser(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.Vessel.SetDeleted(ctx, true, VesselIDs...)
		if err != nil && !errors.Is(err, myErr.ErrNotExist) {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error delete vessels", zap.Error(err), zap.Any("ids", VesselIDs))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString("Ok")
		return
	}
}

// RestoreVessel
// @Tags        Vessel
// @Summary     Восстановление судна
// @Description
// @Accept      json
// @Produce     json
// @Param       VesselNames   body     []domain.VesselIDs    true "список ID Судов"
// @Success     200           {string} string "Ok"
// @Failure     400
// @Failure     500
// @Router      /vessels [patch]
// @Security    BearerAuth
func (h *Handler) RestoreVessel() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			VesselIDs []domain.VesselID
		)
		err = c.BodyParser(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.Vessel.SetDeleted(ctx, false, VesselIDs...)
		if err != nil && !errors.Is(err, myErr.ErrNotExist) {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error restore vessels", zap.Error(err), zap.Any("ids", VesselIDs))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString("Ok")
		return
	}
}
