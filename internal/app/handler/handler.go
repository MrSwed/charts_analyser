package handler

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"
)

type Handler struct {
	app *fiber.App
	s   *service.Service
	log *zap.Logger
}

func NewHandler(app *fiber.App, s *service.Service, log *zap.Logger) *Handler {
	return &Handler{s: s, log: log, app: app}
}

// Handler init routes
func (h *Handler) Handler() *Handler {

	h.app.Use(logger.New(logger.Config{}))

	//h.app.Use(ginzap.Ginzap(h.log, time.RFC3339, true))

	// TODO maybe:
	//h.app.Use(h.getUserID())

	api := h.app.Group(constant.RouteApi)
	api.Get(constant.RouteZones, h.Zones())
	api.Get(constant.RouteVessels, h.Vessels())

	monitor := api.Group(constant.RouteMonitor)
	monitor.Get("/state", h.VesselState())
	monitor.Get("", h.MonitoredList())
	monitor.Post("", h.SetControl())
	monitor.Delete("", h.DelControl())

	track := api.Group(constant.RouteTrack)
	track.Post(constant.RouteID, h.Track())
	track.Get(constant.RouteID, h.GetTrack())

	return h
}
