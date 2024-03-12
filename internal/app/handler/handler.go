package handler

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/service"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	app  *fiber.App
	s    *service.Service
	log  *zap.Logger
	conf *config.Config
}

func NewHandler(app *fiber.App, s *service.Service, conf *config.Config, log *zap.Logger) *Handler {
	return &Handler{s: s, log: log, app: app, conf: conf}
}

// Handler init routes
func (h *Handler) Handler() *Handler {

	api := h.app.Group(constant.RouteAPI)
	api.Use(GetAccessWare(&h.conf.JWT))

	opAw := CheckIsRole(constant.RoleOperator)
	veAw := CheckIsRole(constant.RoleVessel)

	chart := api.Group(constant.RouteChart)
	chart.Use(opAw)
	chart.Get(constant.RouteZones, h.ChartZones())
	chart.Get(constant.RouteVessels, h.ChartVessels())

	monitor := api.Group(constant.RouteMonitor)
	monitor.Use(opAw)
	monitor.Get(constant.RouteState, h.VesselState())
	monitor.Get("", h.MonitoredList())
	monitor.Post("", h.SetControl())
	monitor.Delete("", h.DelControl())

	track := api.Group(constant.RouteTrack)
	track.Post("", veAw, h.Track())
	track.Get(constant.RouteID, opAw, h.GetTrack())

	return h
}
