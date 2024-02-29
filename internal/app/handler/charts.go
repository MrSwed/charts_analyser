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
// @Param       InputVesselsInterval        body     domain.InputVesselsInterval true "Входные параметры: идентификаторы судов, стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {object} []string
// @Failure     400
// @Failure     500
// @Router      /zones [get]
// @Security    BearerAuth
func (h *Handler) Zones() fiber.Handler {
	//
	return func(c *fiber.Ctx) (err error) {
		var (
			query domain.InputVesselsInterval
		)
		err = c.BodyParser(&query)
		if err != nil && !errors.Is(err, io.EOF) || len(query.VesselIDs) == 0 {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()
		var result []domain.ZoneName
		result, err = h.s.Chart.Zones(ctx, query)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error get zones", zap.Error(err))
			return nil
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// Vessels
// @Tags        Chart
// @Summary     список судов
// @Description которые пересекали указанные морские карты в заданный временной промежуток.
// @Accept      json
// @Param       InputZones                 body      domain.InputZones            true  "Входные параметры: идентификаторы карт, стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {object} []uint64
// @Failure     400
// @Failure     500
// @Router      /vessel [get]
// @Security    BearerAuth
func (h *Handler) Vessels() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			query domain.InputZones
		)
		err = c.BodyParser(&query)
		if err != nil && !errors.Is(err, io.EOF) || len(query.ZoneNames) == 0 {
			c.Status(http.StatusBadRequest)
			return nil
		}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		var result []domain.VesselID
		result, err = h.s.Chart.Vessels(ctx, query)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			h.log.Error("Error vessel zones", zap.Error(err))
			return nil
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// Track
// @Tags        Track
// @Summary     Запись трека судна
// @Description
// @Accept      json
// @Param       Authorization  header string          true "Bearer: JWT claims must have: id key used as vesselID and role: 1"
// @Param       VesselID       header domain.VesselID true "id field of jwt key"
// @Param       Point          body   domain.Point    true "[lon, lat]"
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Router      /track/ [post]
// @Security    BearerAuth
func (h *Handler) Track() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		var (
			location domain.InputPoint
			id       = GetVesselId(c)
		)
		if id == 0 {
			c.Status(http.StatusForbidden)
			return nil
		}
		err = c.BodyParser(&location)
		if err != nil && !errors.Is(err, io.EOF) {
			c.Status(http.StatusBadRequest)
			return nil
		}
		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		if err = h.s.Track(ctx, id, location); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.Status(http.StatusNotFound)
				return nil
			}
			if errors.Is(err, myErr.ErrLocationOutOfRange) {
				_, err = c.Status(http.StatusBadRequest).WriteString(myErr.ErrLocationOutOfRange.Error())
				return
			}
			c.Status(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("id", id), zap.Any("location", location))
			return nil
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
// @Param       id            path      uint64               true  "ID Судна "
// @Param       DateInterval  query     domain.DateInterval  true  "Входные параметры: стартовая дата, конечная дата."
// @Produce     json
// @Success     200          {string} string "Ok"
// @Failure     400
// @Failure     500
// @Router      /track/{id} [post]
// @Security    BearerAuth
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
			return nil
		}
		err = id.SetFromStr(c.Params("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return nil
		}
		query.VesselIDs = domain.VesselIDs{id}

		ctx, cancel := context.WithTimeout(c.Context(), constant.ServerOperationTimeout)
		defer cancel()

		if result, err = h.s.GetTrack(ctx, query); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.Status(http.StatusNotFound)
				return nil
			}
			c.Status(http.StatusInternalServerError)
			h.log.Error("GetTrack", zap.Error(err), zap.Any("id", id), zap.Any("query", query))
			return nil
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}
