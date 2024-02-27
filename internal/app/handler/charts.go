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

// Zones
// @Tags        Chart
// @Summary     список морских карт
// @Description которые пересекались заданными в запросе судами в заданный временной промежуток.
// @Accept      json
// @Param       json {object} body     domain.InputVesselsInterval false "Входные параметры: идентификаторы судов, стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {object} []string
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /vessel [get]
func (h *Handler) Zones() fiber.Handler {
	//
	return func(c *fiber.Ctx) (err error) {
		var (
			query domain.InputVesselsInterval
		)
		err = c.BodyParser(&query)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()
		var result []domain.ZoneName
		result, err = h.s.Chart.Zones(ctx, query)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error get zones", zap.Error(err))

		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// Vessels
// @Tags        Chart
// @Summary     список судов
// @Description которые пересекали заданные в запросе морские карты в заданный временной промежуток.
// @Accept      json
// @Param       {object} query     domain.InputZones false "Входные параметры: идентификатор карт, стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {object} []uint64
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /zones [get]
func (h *Handler) Vessels() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			query domain.InputZones
		)
		err = c.BodyParser(&query)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		var result []domain.VesselID
		result, err = h.s.Chart.Vessels(ctx, query)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error vessel zones", zap.Error(err))
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// Track
// @Tags        Track
// @Summary     Запись трека судна
// @Description
// @Accept      json
// @Param       vessel_id     header  {int64}  domain.VesselID true "ID Судна"
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /track/ [post]
// @Security    ApiKeyAuth
func (h *Handler) Track() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			location domain.Point
			id       = GetVesselId(c)
		)
		//err = id.SetFromStr(c.Params("id"))
		//if err != nil {
		//	c.Status(http.StatusBadRequest)
		//	return
		//}
		if id == 0 {
			c.Status(http.StatusForbidden)
			return
		}
		err = c.BodyParser(&location)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		if err = h.s.Track(ctx, id, location); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.Status(http.StatusNotFound)
				return
			}
			if errors.Is(err, myErr.ErrLocationOutOfRange) {
				_, err = c.Status(http.StatusBadRequest).WriteString(myErr.ErrLocationOutOfRange.Error())
				return
			}
			c.Status(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("id", id), zap.Any("location", location))
			return
		}
		_, err = c.Status(http.StatusOK).WriteString("ok")
		return
	}
}

// GetTrack
// @Tags        GetTrack
// @Summary     Маршрут судна за указанный период
// @Description
// @Accept      json
// @Param       {object} query     domain.DateInterval false "Входные параметры: стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /track/{id} [post]
// @Security    ApiKeyAuth
func (h *Handler) GetTrack() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			id     domain.VesselID
			query  domain.InputVesselsInterval
			result []domain.Track
		)
		err = c.QueryParser(&query)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		err = id.SetFromStr(c.Params("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		query.VesselIDs = domain.VesselIDs{id}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		if result, err = h.s.GetTrack(ctx, query); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.Status(http.StatusNotFound)
				return
			}
			c.Status(http.StatusInternalServerError)
			h.log.Error("GetTrack", zap.Error(err), zap.Any("id", id), zap.Any("query", query))
			return
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}
