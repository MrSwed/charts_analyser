package handler

// used  env fo real databases

import (
	"bytes"
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"charts_analyser/internal/app/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type testConfig struct {
	config.Config
	jwtOperator string
}

func newEnvConfig() (c *testConfig) {
	err := godotenv.Load(".env.test")
	if err != nil {
		log.Fatal(err)
	}
	c = &testConfig{}
	c.WithEnv().CleanSchemes()
	c.jwtOperator, err = domain.NewClaimOperator(&c.JWT, 12, "Operator for test").Token()
	if err != nil {
		log.Fatal(err)
	}
	return
}

var (
	conf = newEnvConfig()
	db   = func() *sqlx.DB {
		db, err := sqlx.Connect("postgres", conf.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db
	}()
)

func TestChartZones(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	timeStart, err := time.Parse("2006-01-02 03:04:05", `2017-01-08 00:00:00`)
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 03:04:05", `2017-01-09 00:00:00`)
	require.NoError(t, err)
	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		body    map[string]interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get vessel zones. No jwt",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"vesselIDs": []int64{vesselID},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Get vessel zones. Wrong role in jwt",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"vesselIDs": []int64{vesselID},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Get vessel zones. Operator",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"vesselIDs": []int64{vesselID},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "application/json",
			},
		},
		{
			name: "Get vessel zones. Bad vessel ids",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"vesselIDs": []string{strconv.FormatInt(vesselID, 10)},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get vessel zones. unknown",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"vesselIDs": []int64{10000000000},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    &[]string{"[]"}[0],
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteChart+constant.RouteZones, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if strings.Contains(test.want.contentType, "application/json") {
				var data []domain.ZoneName
				err = json.Unmarshal(resBody, &data)
				require.NoError(t, err)
				if test.want.responseLen != nil {
					if *test.want.responseLen {
						assert.Greater(t, len(resBody), 0)
					} else {
						assert.Equal(t, len(resBody), 0)
					}
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestChartVessels(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	zoneName := "zone_40"
	timeStart, err := time.Parse("2006-01-02 03:04:05", `2017-01-08 00:00:00`)
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 03:04:05", `2017-01-09 00:00:00`)
	require.NoError(t, err)

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		query   map[string]interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get zone vessels. No JWT",
			args: args{
				method: http.MethodPost,
				query: map[string]interface{}{
					"zoneNames": []string{zoneName},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Get zone vessels. ok",
			args: args{
				method: http.MethodPost,
				query: map[string]interface{}{
					"zoneNames": []string{zoneName},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "application/json",
			},
		},
		{
			name: "Get zone vessels. Unknown",
			args: args{
				method: http.MethodPost,
				query: map[string]interface{}{
					"zoneNames": []string{"zone_XXX"},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    &[]string{"[]"}[0],
				contentType: "application/json",
			},
		},
		{
			name: "Get zone vessels. bad parameters",
			args: args{
				method: http.MethodPost,
				query: map[string]interface{}{
					"vesselIDs": []string{"wrong keys"},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(&test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteChart+constant.RouteVessels, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if strings.Contains(test.want.contentType, "application/json") {
				var data []domain.VesselID
				err = json.Unmarshal(resBody, &data)
				require.NoError(t, err)
				if test.want.responseLen != nil {
					if *test.want.responseLen {
						assert.Greater(t, len(resBody), 0)
					} else {
						assert.Equal(t, len(resBody), 0)
					}
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestVesselState(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		query   interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Control vessel. No jwt",
			args: args{
				method: http.MethodPost,
				query:  []int64{vesselID},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Control vessel. Wrong role in jwt",
			args: args{
				method: http.MethodPost,
				query:  []int64{vesselID},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Control vessel. Operator",
			args: args{
				method: http.MethodPost,
				query:  []int64{vesselID},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "application/json",
			},
		},
		{
			name: "Control vessel. Bad vessel ids",
			args: args{
				method: http.MethodPost,
				query:  []string{strconv.FormatInt(vesselID, 10)},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Control vessel. unknown",
			args: args{
				method: http.MethodPost,
				query:  []int64{10000000000},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    &[]string{"[]"}[0],
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteMonitor+constant.RouteState, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if strings.Contains(test.want.contentType, "application/json") {
				var data []domain.VesselState
				err = json.Unmarshal(resBody, &data)
				require.NoError(t, err)
				if test.want.responseLen != nil {
					if *test.want.responseLen {
						assert.Greater(t, len(resBody), 0)
					} else {
						assert.Equal(t, len(resBody), 0)
					}
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestMonitoredList(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Monitor list. No jwt",
			args: args{
				method: http.MethodGet,
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Monitor list. Wrong role in jwt",
			args: args{
				method: http.MethodGet,
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Monitor list. Operator",
			args: args{
				method: http.MethodGet,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteMonitor, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if strings.Contains(test.want.contentType, "application/json") {
				var data []domain.Vessel
				err = json.Unmarshal(resBody, &data)
				require.NoError(t, err)
				if test.want.responseLen != nil {
					if *test.want.responseLen {
						assert.Greater(t, len(resBody), 0)
					} else {
						assert.Equal(t, len(resBody), 0)
					}
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestSetControl(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		query   interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Set control. No jwt",
			args: args{
				method: http.MethodPost,
				query:  []int64{vesselID},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Set control. Wrong role in jwt",
			args: args{
				method: http.MethodPost,
				query:  []int64{vesselID},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Set control. Operator",
			args: args{
				method: http.MethodPost,
				query:  []int64{vesselID},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Set control. Bad vessel ids",
			args: args{
				method: http.MethodPost,
				query:  []string{strconv.FormatInt(vesselID, 10)},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Set control. No vessel ids",
			args: args{
				method: http.MethodPost,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Set control. unknown",
			args: args{
				method: http.MethodPost,
				query:  []int64{10000000000},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteMonitor, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestDeleteControl(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		query   interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Set control. No jwt",
			args: args{
				method: http.MethodDelete,
				query:  []int64{vesselID},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Delete control. Wrong role in jwt",
			args: args{
				method: http.MethodDelete,
				query:  []int64{vesselID},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Delete control. Operator",
			args: args{
				method: http.MethodDelete,
				query:  []int64{vesselID},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Delete control. Bad vessel ids",
			args: args{
				method: http.MethodDelete,
				query:  []string{strconv.FormatInt(vesselID, 10)},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Delete control. No vessel ids",
			args: args{
				method: http.MethodDelete,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Delete control. unknown",
			args: args{
				method: http.MethodDelete,
				query:  []int64{10000000000},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteMonitor, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestTrack(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	claimsUnknownVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(10000000000), "")
	jwtUnknownVessel, err := claimsUnknownVessel.Token()
	require.NoError(t, err)

	claimsNoVessel := jwt.NewWithClaims(jwt.SigningMethodHS512, struct {
		jwt.RegisteredClaims
		Role int64 `json:"role"`
	}{
		Role: 1,
	})

	jetVesselWithoutID, err := claimsNoVessel.SigningString()
	require.NoError(t, err)

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		query   interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Track. No jwt",
			args: args{
				method: http.MethodPost,
				query:  []float64{12.12, 12.12},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Track. Wrong role in jwt, operator",
			args: args{
				method: http.MethodPost,
				query:  []float64{12.12, 12.12},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Track. OK",
			args: args{
				method: http.MethodPost,
				query:  []float64{12.12, 12.12},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Track. No track data",
			args: args{
				method: http.MethodPost,
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Track. Bad track data",
			args: args{
				method: http.MethodPost,
				query:  []string{"12.12", "12.12"},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Track. No vessel ids",
			args: args{
				method: http.MethodPost,
				headers: map[string]string{
					"Authorization": "Bearer " + jetVesselWithoutID,
				},
			},
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{
			name: "Track. NO coordinates",
			args: args{
				method: http.MethodPost,
				query:  []int64{},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Track. Unknown vessel",
			args: args{
				method: http.MethodPost,
				query:  []float64{12.12, 12.12},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtUnknownVessel,
				},
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteTrack, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestAddVessel(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []string{"Test1 Vessel_" + timeID, "Test2 Vessel_" + timeID}

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		body    interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Add vessel. No jwt",
			args: args{
				method: http.MethodPost,
				body:   newVessels,
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Add vessel. Wrong role in jwt",
			args: args{
				method: http.MethodPost,
				body:   newVessels,
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Add vessel. OK",
			args: args{
				method: http.MethodPost,
				body:   newVessels,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusCreated,
				responseLen: &[]bool{true}[0],
				contentType: "application/json",
			},
		},
		{
			name: "Add vessel. No body data",
			args: args{
				method: http.MethodPost,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Add vessel. Bad body data. Strings only",
			args: args{
				method: http.MethodPost,
				body:   []interface{}{12.12, "vessel name"},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Add vessel. Repeated names. Re-add vessel - return earlier added",
			args: args{
				method: http.MethodPost,
				body:   append(newVessels, newVessels...),

				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusCreated,
				responseLen: &[]bool{true}[0],
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteVessels, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestDeleteVessel(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := s.Vessel.AddVessel(ctx, newVessels...)
	require.NoError(t, err)
	require.Equal(t, len(newVessels), len(vessels))

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		body    interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Delete vessel. No jwt",
			args: args{
				method: http.MethodDelete,
				body:   vessels.IDs(),
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Delete vessel. Wrong role in jwt",
			args: args{
				method: http.MethodDelete,
				body:   vessels.IDs(),
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Delete vessel. OK",
			args: args{
				method: http.MethodDelete,
				body:   vessels.IDs(),
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Delete vessel. No body data",
			args: args{
				method: http.MethodDelete,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Delete vessel. Bad body data.",
			args: args{
				method: http.MethodDelete,
				body:   []interface{}{12.12, "vessel name", 1555444},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Delete vessel. Repeated ids.",
			args: args{
				method: http.MethodDelete,
				body:   append(vessels.IDs(), vessels.IDs()...),

				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteVessels, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestRestoreVessel(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := s.Vessel.AddVessel(ctx, newVessels...)
	require.NoError(t, err)
	err = s.Vessel.SetDeleted(ctx, true, vessels.IDs()...)
	require.NoError(t, err)

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		body    interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Restore vessel. No jwt",
			args: args{
				method: http.MethodPatch,
				body:   vessels.IDs(),
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Restore vessel. Wrong role in jwt",
			args: args{
				method: http.MethodPatch,
				body:   vessels.IDs(),
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Restore vessel. OK",
			args: args{
				method: http.MethodPatch,
				body:   vessels.IDs(),
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Restore vessel. No body data",
			args: args{
				method: http.MethodPatch,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Restore vessel. Bad body data.",
			args: args{
				method: http.MethodPatch,
				body:   []interface{}{12.12, "vessel name", 1555444},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Restore vessel. Repeated ids.",
			args: args{
				method: http.MethodPatch,
				body:   append(vessels.IDs(), vessels.IDs()...),

				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteVessels, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestGetVessel(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := s.Vessel.AddVessel(ctx, newVessels...)
	require.NoError(t, err)
	require.Equal(t, len(newVessels), len(vessels))
	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		query   map[string]interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get vessel. No jwt",
			args: args{
				method: http.MethodGet,
				query:  map[string]interface{}{"vesselIDs": vessels.IDs()},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Get vessel. Wrong role in jwt",
			args: args{
				method: http.MethodGet,
				query:  map[string]interface{}{"vesselIDs": vessels.IDs()},
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Get vessel. OK",
			args: args{
				method: http.MethodGet,
				query:  map[string]interface{}{"vesselIDs": vessels.IDs()},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:            http.StatusOK,
				responseLen:     &[]bool{true}[0],
				responseContain: vessels.IDs()[0].String(),
				contentType:     "application/json",
			},
		},
		{
			name: "Get vessel. No query data, empty list",
			args: args{
				method: http.MethodGet,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:            http.StatusOK,
				responseLen:     &[]bool{true}[0],
				responseContain: "[]",
				contentType:     "application/json",
			},
		},
		{
			name: "Get vessel. Bad body data.",
			args: args{
				method: http.MethodGet,
				query:  map[string]interface{}{"vesselIDs": []interface{}{12.12, "vessel name", 1555444}},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get vessel. Repeated ids.",
			args: args{
				method: http.MethodGet,
				query:  map[string]interface{}{"vesselIDs": append(vessels.IDs(), vessels.IDs()...)},

				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:            http.StatusOK,
				responseContain: vessels.IDs()[0].String(),
				responseLen:     &[]bool{true}[0],
				contentType:     "application/json",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteVessels, nil)
			require.NoError(t, err)

			queries := url.Values{}
			for k, v := range test.args.query {
				switch vv := v.(type) {
				case []domain.VesselID:
					for _, vs := range vv {
						queries.Add(k, vs.String())
					}
				case []interface{}:
					for _, vs := range vv {
						queries.Add(k, fmt.Sprintf("%v", vs))
					}
				default:
					queries.Add(k, fmt.Sprintf("%v", vv))
				}
			}
			queries.Encode()
			request.URL.RawQuery = queries.Encode()

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}

func TestUpdateVessel(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := domain.NewClaimVessels(&conf.JWT, domain.VesselID(vesselID), "")
	jwtVessel, err := claimsVessel.Token()
	require.NoError(t, err)

	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := s.Vessel.AddVessel(ctx, newVessels...)
	require.NoError(t, err)
	require.Equal(t, len(newVessels), len(vessels))

	type want struct {
		code            int
		responseLen     *bool
		response        *string
		responseContain string
		contentType     string
	}
	type args struct {
		method  string
		body    interface{}
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Update vessels. No jwt",
			args: args{
				method: http.MethodPut,
				body:   vessels,
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Update vessels. Wrong role in jwt",
			args: args{
				method: http.MethodPut,
				body:   vessels,
				headers: map[string]string{
					"Authorization": "Bearer " + jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Update vessels. OK",
			args: args{
				method: http.MethodPut,
				body:   vessels,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:            http.StatusOK,
				responseLen:     &[]bool{true}[0],
				responseContain: vessels.IDs()[0].String(),
				contentType:     "application/json",
			},
		},
		{
			name: "Update vessels. No body data, empty list",
			args: args{
				method: http.MethodPut,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Update vessels. Bad body data.",
			args: args{
				method: http.MethodPut,
				body:   map[string]interface{}{"vesselIDs": []interface{}{12.12, "vessel name", 1555444}},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Update vessels. Not exist.",
			args: args{
				method: http.MethodPut,
				body:   domain.Vessels{{ID: 10050000, Name: "NotExist Vessel"}},

				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code:            http.StatusOK,
				responseContain: "[]",
				responseLen:     &[]bool{true}[0],
				contentType:     "application/json",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteVessels, bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
			require.NoError(t, err)

			var resBody []byte
			assert.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseLen != nil {
				if *test.want.responseLen {
					assert.Greater(t, len(resBody), 0)
				} else {
					assert.Equal(t, len(resBody), 0)
				}
			}

			if test.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			}

			if test.want.responseContain != "" {
				cont := string(resBody)
				assert.Contains(t, cont, test.want.responseContain)
			}

			if test.want.response != nil {
				cont := strings.TrimSpace(string(resBody))
				assert.Equal(t, cont, *test.want.response)
			}
		})
	}
}
