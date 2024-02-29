package handler

// to test with real db set env DATABASE_DSN before run with created, but empty tables

import (
	"bytes"
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"charts_analyser/internal/app/service"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
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
	jwtVessel   string
}

func newEnvConfig() (c *testConfig) {
	err := godotenv.Load(".env.test")
	if err != nil {
		log.Fatal(err)
	}
	c = &testConfig{}
	c.WithEnv().CleanSchemes()
	c.jwtOperator = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEwMCIsIm5hbWUiOiJPcGVyYXRvcjEwMCIsInJvbGUiOjJ9.OSc0cSsEvxcz_waNjenJlJiCA9xcIjs1ZvDTi9RNBuKAvD5hBLDvm7XwFCIg9uv-lK-Yxb-62XJuuiNxA0FlcA"
	c.jwtVessel = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjkyMzM0NjYiLCJuYW1lIjoiU2FnYSBWaWtpbmciLCJyb2xlIjoxfQ.7gbx8YufboNZo3uqGXcFRHIIoIWZt8tQdtNFiuun9Z3_e5a9IWOFw88Wx2rQuhkJwgB6SBPvP_rpqqmK36yL0w"

	return
}

var (
	conf     = newEnvConfig()
	redisCli = func() *redis.Client {
		client := redis.NewClient(&redis.Options{
			Addr:     conf.RedisAddress,
			Password: conf.RedisPass,
			DB:       0,
		})
		if _, err := client.Ping(context.TODO()).Result(); err != nil {
			log.Fatal("cannot connect redis", err)
		}
		return client
	}()
	db = func() *sqlx.DB {
		db, err := sqlx.Connect("postgres", conf.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db
	}()
)

func TestZones(t *testing.T) {
	repo := repository.NewRepository(db, redisCli)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselId := int64(9110913)
	timeStart, err := time.Parse("2006-01-02 03:04:05", `2017-01-08 00:00:00`)
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 03:04:05", `2020-10-09 00:00:00`)
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
			name: "Get vessel zones. No jwt",
			args: args{
				method: http.MethodGet,
				query: map[string]interface{}{
					"vesselIDs": []int64{vesselId},
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
			name: "Get vessel zones, Wrong role in jwt",
			args: args{
				method: http.MethodGet,
				query: map[string]interface{}{
					"vesselIDs": []int64{vesselId},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Get vessel zones, Operator",
			args: args{
				method: http.MethodGet,
				query: map[string]interface{}{
					"vesselIDs": []int64{vesselId},
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
				method: http.MethodGet,
				query: map[string]interface{}{
					"vesselIDs": []string{strconv.FormatInt(vesselId, 10)},
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
			name: "Get vessels zones, unknown",
			args: args{
				method: http.MethodGet,
				query: map[string]interface{}{
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
		t.Run(test.name, func(t *testing.T) {

			bodyJson, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteApi+constant.RouteZones, bytes.NewReader(bodyJson))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
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

func TestVessels(t *testing.T) {
	repo := repository.NewRepository(db, redisCli)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	zoneName := "zone_205"
	timeStart, err := time.Parse("2006-01-02 03:04:05", `2017-01-08 00:00:00`)
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 03:04:05", `2020-10-09 00:00:00`)
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
			name: "Get zone 'zone_205' vessels. No JWT",
			args: args{
				method: http.MethodGet,
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
			name: "Get zone 'zone_205' vessels",
			args: args{
				method: http.MethodGet,
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
			name: "Get zone 'zone_XXX' vessels",
			args: args{
				method: http.MethodGet,
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
			name: "Get zones with bad parameters",
			args: args{
				method: http.MethodGet,
				query: map[string]interface{}{
					"vesselIDs": []string{"fring keys"},
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
		t.Run(test.name, func(t *testing.T) {

			bodyJson, _ := json.Marshal(&test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteApi+constant.RouteVessels, bytes.NewReader(bodyJson))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
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
	repo := repository.NewRepository(db, redisCli)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselId := int64(9110913)

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
				method: http.MethodGet,
				query:  []int64{vesselId},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Control vessel, Wrong role in jwt",
			args: args{
				method: http.MethodGet,
				query:  []int64{vesselId},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Control vessel, Operator",
			args: args{
				method: http.MethodGet,
				query:  []int64{vesselId},
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
				method: http.MethodGet,
				query:  []string{strconv.FormatInt(vesselId, 10)},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get vessels, unknown",
			args: args{
				method: http.MethodGet,
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
		t.Run(test.name, func(t *testing.T) {

			bodyJson, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteApi+constant.RouteMonitor+constant.RouteState, bytes.NewReader(bodyJson))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			if len(test.args.headers) > 0 {
				for k, v := range test.args.headers {
					request.Header.Set(k, v)
				}
			}

			res, err := app.Test(request)
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

/**/
/*todo tests
MonitoredList
SetControl
DelControl
Track
*/

/**/
