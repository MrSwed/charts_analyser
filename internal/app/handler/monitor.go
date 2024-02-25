package handler

import (
	"charts_analyser/internal/app/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

// SetControl
// @Tags        SetControl
// @Summary     Поставить судно на контроль
// @Description
// на мониторинг (снять с мониторинга)
// @Accept      json
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /monitor/:id [post]
func (h *Handler) SetControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			vessel = domain.Vessel{}
			id     = c.Param("id")
			query  domain.InputVessel
		)
		err := vessel.ID.SetFromStr(id)
		if err != nil {
			h.log.Error("SetControl id error", zap.Error(err), zap.Any("id", id))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		if err = c.BindQuery(&query); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if query.VesselName != "" {
			vessel.Name = query.VesselName
		}
		err = h.s.SetControl(c, vessel, true)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("id", id))
			return
		}
		status := http.StatusOK
		c.String(status, "ok")
	}
}

// DelControl
// @Tags        DelControl
// @Summary     Снять судно с контроля
// @Description
// @Accept      json
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /monitor/:id [delete]
func (h *Handler) DelControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			vessel = domain.Vessel{}
			id     = c.Param("id")
		)
		err := vessel.ID.SetFromStr(id)
		if err != nil {
			h.log.Error("DelControl id error", zap.Error(err), zap.Any("id", id))
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		err = h.s.SetControl(c, vessel, false)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("DelControl", zap.Error(err), zap.Any("id", id))
		}
		status := http.StatusOK
		c.String(status, "ok")
	}
}
