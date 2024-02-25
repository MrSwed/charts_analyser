package handler

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/service"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Handler struct {
	s   *service.Service
	r   *gin.Engine
	log *zap.Logger
}

func NewHandler(s *service.Service, log *zap.Logger) *Handler {
	return &Handler{s: s, log: log}
}

// Handler init routes
func (h *Handler) Handler() http.Handler {

	h.r = gin.New()
	h.r.Use(ginzap.Ginzap(h.log, time.RFC3339, true))

	// TODO maybe:
	//h.r.Use(middleware.Compress(gzip.DefaultCompression))
	//h.r.Use(middleware.Decompress())
	//h.r.Use(h.getUserID())

	api := h.r.Group(constant.RouteApi)
	api.GET(constant.RouteZones, h.Zones())
	api.GET(constant.RouteVessels, h.Vessels())

	monitor := api.Group(constant.RouteMonitor)
	monitor.GET("", h.MonitoredList())
	monitor.POST("", h.SetControl())
	monitor.DELETE("", h.DelControl())

	return h.r
}
