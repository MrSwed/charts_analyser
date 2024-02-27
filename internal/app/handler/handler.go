package handler

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	h.app.Use(logger.New(logger.Config{}))

	api := h.app.Group(constant.RouteApi)
	api.Use(GetAccessWare(&h.conf.JWT))

	opAw := CheckIsRole(constant.RoleOperator)
	veAw := CheckIsRole(constant.RoleVessel)
	api.Get(constant.RouteZones, opAw, h.Zones())
	api.Get(constant.RouteVessels, opAw, h.Vessels())

	monitor := api.Group(constant.RouteMonitor)
	monitor.Use(opAw)
	monitor.Get("/state", h.VesselState())
	monitor.Get("", h.MonitoredList())
	monitor.Post("", h.SetControl())
	monitor.Delete("", h.DelControl())

	track := api.Group(constant.RouteTrack)
	track.Use(veAw)
	track.Post(constant.RouteID, h.Track())
	track.Get(constant.RouteID, h.GetTrack())

	return h
}
