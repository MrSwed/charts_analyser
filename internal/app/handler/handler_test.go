package handler

// used  env fo real databases

import (
	"bytes"
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"charts_analyser/internal/app/service"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v4"
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
}

type ClaimsVessel struct {
	jwt.RegisteredClaims
	*domain.Vessel
	Role int64 `json:"role"`
}

func (c *ClaimsVessel) Token(key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, ClaimsVessel{
		RegisteredClaims: jwt.RegisteredClaims{},
		Vessel:           c.Vessel,
		Role:             1,
	})

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func newEnvConfig() (c *testConfig) {
	err := godotenv.Load(".env.test")
	if err != nil {
		log.Fatal(err)
	}
	c = &testConfig{}
	c.WithEnv().CleanSchemes()
	c.jwtOperator = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEwMCIsIm5hbWUiOiJPcGVyYXRvcjEwMCIsInJvbGUiOjJ9.OSc0cSsEvxcz_waNjenJlJiCA9xcIjs1ZvDTi9RNBuKAvD5hBLDvm7XwFCIg9uv-lK-Yxb-62XJuuiNxA0FlcA"

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

func TestZones(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	vesselID := int64(9110913)
	claimsVessel := ClaimsVessel{Vessel: &domain.Vessel{ID: domain.VesselID(vesselID)}}
	jwtVessel, err := claimsVessel.Token(conf.JWTSigningKey)
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
				method: http.MethodGet,
				query: map[string]interface{}{
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
				method: http.MethodGet,
				query: map[string]interface{}{
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
				method: http.MethodGet,
				query: map[string]interface{}{
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
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteZones, bytes.NewReader(bodyJSON))
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

func TestVessels(t *testing.T) {
	repo := repository.NewRepository(db)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	_ = NewHandler(app, s, &conf.Config, logger).Handler()

	zoneName := "zone_205"
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
			name: "Get zone vessels. ok",
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
			name: "Get zone vessels. Unknown",
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
			name: "Get zone vessels. bad parameters",
			args: args{
				method: http.MethodGet,
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
	claimsVessel := ClaimsVessel{Vessel: &domain.Vessel{ID: domain.VesselID(vesselID)}}
	jwtVessel, err := claimsVessel.Token(conf.JWTSigningKey)
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
				method: http.MethodGet,
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
				method: http.MethodGet,
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
				method: http.MethodGet,
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
				method: http.MethodGet,
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
	claimsVessel := ClaimsVessel{Vessel: &domain.Vessel{ID: domain.VesselID(vesselID)}}
	jwtVessel, err := claimsVessel.Token(conf.JWTSigningKey)
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
	claimsVessel := ClaimsVessel{Vessel: &domain.Vessel{ID: domain.VesselID(vesselID)}}
	jwtVessel, err := claimsVessel.Token(conf.JWTSigningKey)
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
	claimsVessel := ClaimsVessel{Vessel: &domain.Vessel{ID: domain.VesselID(vesselID)}}
	jwtVessel, err := claimsVessel.Token(conf.JWTSigningKey)
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
	claimsVessel := ClaimsVessel{Vessel: &domain.Vessel{ID: domain.VesselID(vesselID)}}
	jwtVessel, err := claimsVessel.Token(conf.JWTSigningKey)
	require.NoError(t, err)

	claimsUnknownVessel := ClaimsVessel{Vessel: &domain.Vessel{ID: domain.VesselID(10000000000)}}
	jwtUnknownVessel, err := claimsUnknownVessel.Token(conf.JWTSigningKey)
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
