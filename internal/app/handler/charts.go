package handler

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// Zones
// @Tags        Chart
// @Summary     список морских карт
// @Description которые пересекались заданными в запросе судами в заданный временной промежуток.
// @Accept      json
// @Param       {object} query     domain.InputVesselsInterval false "Входные параметры: идентификаторы судов, стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {object} []string
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /vessel [get]
func (h *Handler) Zones() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			query domain.InputVesselsInterval
		)
		// todo : if post, use json
		//if body, err := c.GetRawData(); err != nil || len(body) == 0 {
		//	c.AbortWithStatus(http.StatusBadRequest)
		//	return
		//}
		//err = json.NewDecoder(c.Request.Body).Decode(&query)

		err := c.BindQuery(&query)
		if err != nil || len(query.VesselIDs) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.Chart.Zones(ctx, query)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("Error get zones", zap.Error(err))
		}
		status := http.StatusOK
		c.JSON(status, result)
	}
}

// Vessels
// @Tags        Chart
// @Summary     список судов
// @Description которые пересекали заданные в запросе морские карты в заданный временной промежуток.
// @Accept      json
// @Param       {object} query     domain.InputZone false "Входные параметры: идентификатор карт, стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {object} []uint64
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /zones [get]
func (h *Handler) Vessels() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			query domain.InputZone
		)
		// todo : if post, use json
		//if body, err := c.GetRawData(); err != nil || len(body) == 0 {
		//	c.AbortWithStatus(http.StatusBadRequest)
		//	return
		//}
		//err = json.NewDecoder(c.Request.Body).Decode(&query)

		err := c.Bind(&query)
		if err != nil || query.ZoneName == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.Chart.Vessels(ctx, query)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("Error vessel zones", zap.Error(err))
		}
		status := http.StatusOK
		c.JSON(status, result)
	}
}

// Track
// @Tags        Track
// @Summary     Запись трека судна
// @Description
// @Accept      json
// @Param       vessel_id     query  {array}  domain.VesselID true "ID Судна"
// @Param       RequestBody   body   []domain.VesselID true "список ID Суден"
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /track/:id [post]
func (h *Handler) Track() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			location domain.Point
			id       domain.VesselID
		)
		err := id.SetFromStr(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		err = c.ShouldBindJSON(&location)
		if err != nil && !errors.Is(err, io.EOF) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		if err = h.s.Track(ctx, id, location); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			if errors.Is(err, myErr.ErrLocationOutOfRange) {
				c.String(http.StatusNotFound, myErr.ErrLocationOutOfRange.Error())
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("id", id), zap.Any("location", location))
			return
		}
		status := http.StatusOK
		c.String(status, "ok")
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
// @Router      /track/:id [post]
func (h *Handler) GetTrack() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			id     domain.VesselID
			query  domain.InputVesselsInterval
			result []domain.Track
		)
		err := c.Bind(&query)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		err = id.SetFromStr(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		query.VesselIDs = domain.VesselIDs{id}

		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		if result, err = h.s.GetTrack(ctx, query); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("GetTrack", zap.Error(err), zap.Any("id", id), zap.Any("query", query))
			return
		}
		status := http.StatusOK
		c.JSON(status, result)
	}
}
