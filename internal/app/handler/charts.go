package handler

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

// Zones
// @Tags        Chart
// @Summary     список морских карт
// @Description которые пересекались заданными в запросе судами в заданный временной промежуток.
// @Accept      json
// @Param       {object} query     domain.InputVessels false "Входные параметры: идентификаторы судов, стартовая дата, конечная дата."
// @Produce     json
// @Success     200         {object} []string
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /vessel [get]
func (h *Handler) Zones() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			query domain.InputVessels
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
