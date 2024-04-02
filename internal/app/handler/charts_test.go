package handler_test

import (
	"bytes"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func (suite *HandlerTestSuite) TestChartZones() {
	t := suite.T()
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
					"vesselIDs": int64(suite.cfg.VesselID),
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
					"vesselIDs": int64(suite.cfg.VesselID),
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Get vessel zones. User",
			args: args{
				method: http.MethodPost,
				body: map[string]interface{}{
					"vesselIDs": []domain.VesselID{suite.cfg.VesselID},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"vesselIDs": []string{strconv.FormatInt(int64(suite.cfg.VesselID), 10)},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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

			res, err := suite.app.Test(request)
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

func (suite *HandlerTestSuite) TestChartVessels() {
	t := suite.T()
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
					"zoneNames": []string{string(suite.cfg.ZoneName)},
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
					"zoneNames": []string{string(suite.cfg.ZoneName)},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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

			res, err := suite.app.Test(request)
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

func (suite *HandlerTestSuite) TestTrack() {
	t := suite.T()
	claimsUnknownVessel := domain.NewClaimVessels(&suite.cfg.JWT, domain.VesselID(10000000000), "")
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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

			res, err := suite.app.Test(request)
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
