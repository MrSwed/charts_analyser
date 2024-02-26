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

// MonitoredList
// @Tags        MonitoredList
// @Summary     Список судов
// @Description поставленных на мониторинг
// @Accept      json
// @Produce     json
// @Success     200         {object} []domain.Vessel "Ok"
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /monitor/ [get]
func (h *Handler) MonitoredList() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.Monitor.MonitoredVessels(ctx)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("Error monitored list", zap.Error(err))
		}
		status := http.StatusOK
		c.JSON(status, result)
	}
}

// VesselState
// @Tags        VesselState
// @Summary     Текущие данные
// @Description для выбранных судов, стоящих на мониторинге
// @Accept      json
// @Produce     json
// @Param       vessel_id     query  {array}  domain.VesselID true "ID Судна"
// @Param       RequestBody   body   []domain.VesselID true "список ID Суден"
// @Success     200         {object} []domain.VesselState "Ok"
// @Failure     400
// @Failure     404           "no data yet"
// @Failure     500
// @Failure     403          :todo
// @Router      /monitor/state [get]
func (h *Handler) VesselState() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			VesselIDs []domain.VesselID
		)
		err := c.ShouldBindJSON(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if len(VesselIDs) == 0 {
			var query domain.InputVessels
			if err = c.BindQuery(&query); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				h.log.Error("SetControl,query", zap.Error(err))
				return
			}
			VesselIDs = query.VesselIDs
		}

		if len(VesselIDs) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		result, err := h.s.Monitor.GetStates(ctx, VesselIDs...)
		if err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("Error get states", zap.Error(err), zap.Any("ids", VesselIDs))
		}
		status := http.StatusOK
		c.JSON(status, result)
	}
}

// SetControl
// @Tags        SetControl
// @Summary     Поставить судно на контроль
// @Description
// на мониторинг (снять с мониторинга)
// @Accept      json
// @Param       vessel_id     query  {array}  domain.VesselID true "ID Судна"
// @Param       RequestBody   body   []domain.VesselID true "список ID Суден"
// @Produce     json
// @Success     200         {string} string "Ok"
// @Failure     400
// @Failure     500
// @Failure     403          :todo
// @Router      /monitor/ [post]
func (h *Handler) SetControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			VesselIDs []domain.VesselID
		)
		err := c.ShouldBindJSON(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if len(VesselIDs) == 0 {
			var query domain.InputVessels
			if err = c.BindQuery(&query); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				h.log.Error("SetControl,query", zap.Error(err))
				return
			}
			VesselIDs = query.VesselIDs
		}

		if len(VesselIDs) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.SetControl(ctx, true, VesselIDs...)
		if err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("ids", VesselIDs))
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
			VesselIDs []domain.VesselID
		)
		err := c.ShouldBindJSON(&VesselIDs)
		if err != nil && !errors.Is(err, io.EOF) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if len(VesselIDs) == 0 {
			var query domain.InputVessels
			if err = c.BindQuery(&query); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				h.log.Error("SetControl,query", zap.Error(err))
				return
			}
			VesselIDs = query.VesselIDs
		}

		if len(VesselIDs) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout)
		defer cancel()

		err = h.s.SetControl(ctx, false, VesselIDs...)
		if err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.Error("SetControl", zap.Error(err), zap.Any("ids", VesselIDs))
			return
		}
		status := http.StatusOK
		c.String(status, "ok")
	}
}
