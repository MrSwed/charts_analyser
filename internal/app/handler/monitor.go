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

// MonitoredList
// @Tags        MonitoredList
// @Summary     Список судов
// @Description поставленных на мониторинг
// @Accept      json
// @Produce     json
// @Success     200         {object} []domain.Vessel "Ok"
// @Failure     400
// @Failure     500
// @Router      /monitor/ [get]
// @Security    BearerAuth
func (h *Handler) MonitoredList() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()
		result, err := h.s.Monitor.MonitoredVessels(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error monitored list", zap.Error(err))
			return nil
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// VesselState
// @Tags        VesselState
// @Summary     Текущие данные
// @Description для выбранных судов, стоящих на мониторинге
// @Accept      json
// @Produce     json
// @Param       VesselIDs   body     []domain.VesselID    true "список ID Судов"
// @Success     200         {object} []domain.VesselState "Ok"
// @Failure     400
// @Failure     404           "no data yet"
// @Failure     500
// @Router      /monitor/state [get]
// @Security    BearerAuth
func (h *Handler) VesselState() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			VesselIDs []domain.VesselID
		)
		err = c.BodyParser(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		if len(VesselIDs) == 0 {
			var query domain.InputVessels
			if err = c.QueryParser(&query); err != nil {
				c.Status(http.StatusBadRequest)
				h.log.Error("SetControl,query", zap.Error(err))
				return nil
			}
			VesselIDs = query.VesselIDs
		}

		if len(VesselIDs) == 0 {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.Monitor.GetStates(ctx, VesselIDs...)
		if err != nil && !errors.Is(err, myErr.ErrNotExist) {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error get states", zap.Error(err), zap.Any("ids", VesselIDs))
			return nil
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// SetControl
// @Tags        SetControl
// @Summary     Поставить судно на контроль
// @Description
// на мониторинг (снять с мониторинга)
// @Accept      json
// @Param       VesselIDs   body   []domain.VesselID true "список ID Судов"
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Router      /monitor/ [post]
// @Security    BearerAuth
func (h *Handler) SetControl() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			VesselIDs []domain.VesselID
		)
		err = c.BodyParser(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		if len(VesselIDs) == 0 {
			var query domain.InputVessels
			if err = c.QueryParser(&query); err != nil {
				c.Status(http.StatusBadRequest)
				h.log.Error("SetControl,query", zap.Error(err))
				return nil
			}
			VesselIDs = query.VesselIDs
		}

		if len(VesselIDs) == 0 {
			c.Status(http.StatusBadRequest)
			return nil
		}
		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.SetControl(ctx, true, VesselIDs...)
		if err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.Status(http.StatusNotFound)
				return nil
			}
			c.Status(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("ids", VesselIDs))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString("ok")
		return
	}
}

// DelControl
// @Tags        DelControl
// @Summary     Снять судно с контроля
// @Description
// @Accept      json
// @Param       VesselIDs   body   []domain.VesselID true "список ID Судов"
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Router      /monitor/{id} [delete]
// @Security    BearerAuth
func (h *Handler) DelControl() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			VesselIDs []domain.VesselID
		)
		err = c.BodyParser(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}

		if len(VesselIDs) == 0 {
			c.Status(http.StatusBadRequest)
			return nil
		}
		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.SetControl(ctx, false, VesselIDs...)
		if err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.Status(http.StatusNotFound)
				return nil
			}
			c.Status(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("ids", VesselIDs))
			return nil
		}
		_, err = c.Status(http.StatusOK).WriteString("ok")
		return
	}
}
