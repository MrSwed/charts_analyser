package service

import (
	"bytes"
	appDomain "charts_analyser/internal/app/domain"
	"charts_analyser/internal/simulator/config"
	"charts_analyser/internal/simulator/constant"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

type RequestService struct {
	c config.Out
	l *zap.Logger
}

func NewRequest(c config.Out, l *zap.Logger) *RequestService {
	return &RequestService{c: c, l: l}
}

func (s *RequestService) SendTrack(ctx context.Context, vesselID appDomain.VesselID, location appDomain.Point) {
	var err error

	var body []byte
	if body, err = json.Marshal(location); err != nil || len(body) == 0 {
		s.l.Error("unmarshal error or empty body!", zap.Error(err), zap.Any("location", location))
		return
	}

	urlStr := s.c.ServerAddress + constant.RouteTrack + "/" + strconv.FormatInt(int64(vesselID), 10)
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewBuffer(body)); err != nil {
		s.l.Error("http.NewRequest", zap.Error(err), zap.Any("url", urlStr), zap.String("body", string(body)))
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		s.l.Error("http.DefaultClient.Do(req)", zap.Error(err))
		return
	}
	var resultBody []byte
	if resultBody, err = io.ReadAll(res.Body); err != nil {
		s.l.Error("io.ReadAll(res.Body)", zap.Error(err))
	}
	if err = res.Body.Close(); err != nil {
		s.l.Error("res.Body.Close", zap.Error(err))
	}
	s.l.Info("SendTrack done", zap.Any("data", []interface{}{
		req.URL.String(), string(body), req.Method, res.StatusCode, string(resultBody)}))

	return
}
func (s *RequestService) SetControl(ctx context.Context, vesselID appDomain.VesselID) {
	var err error

	urlStr := s.c.ServerAddress + constant.RouteMonitor
	var req *http.Request
	if req, err = http.NewRequest("POST", urlStr, nil); err != nil {
		s.l.Error("http.NewRequest", zap.Error(err), zap.Any("url", urlStr))
		return
	}
	q := req.URL.Query()
	q.Add("vessel_id", vesselID.String())
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		s.l.Error("http.DefaultClient.Do(req)", zap.Error(err))
		return
	}
	var resultBody []byte
	if resultBody, err = io.ReadAll(res.Body); err != nil {
		s.l.Error("io.ReadAll(res.Body)", zap.Error(err))
	}
	if err = res.Body.Close(); err != nil {
		s.l.Error("res.Body.Close", zap.Error(err))
	}
	s.l.Info("SetControl done", zap.Any("data", []interface{}{
		req.URL.String(), req.Method, res.StatusCode, string(resultBody)}))

	return
}