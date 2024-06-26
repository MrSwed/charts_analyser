package handler_test

import (
	"bytes"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func (suite *HandlerTestSuite) TestVesselState() {
	t := suite.T()

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
				query:  []int64{int64(suite.cfg.VesselID)},
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
				query:  []int64{int64(suite.cfg.VesselID)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Control vessel. Ok",
			args: args{
				method: http.MethodPost,
				query:  []int64{int64(suite.cfg.VesselID)},
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
			name: "Control vessel. Bad vessel ids",
			args: args{
				method: http.MethodPost,
				query:  []string{strconv.FormatInt(int64(suite.cfg.VesselID), 10)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
			bodyJSON, _ := json.Marshal(test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteMonitor+constant.RouteState, bytes.NewReader(bodyJSON))
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
				var data []domain.VesselState
				err = json.Unmarshal(resBody, &data)
				require.NoError(t, err, string(resBody))
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

func (suite *HandlerTestSuite) TestMonitoredList() {
	t := suite.T()
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Monitor list. User",
			args: args{
				method: http.MethodGet,
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

func (suite *HandlerTestSuite) TestSetControl() {
	t := suite.T()
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
				query:  []int64{int64(suite.cfg.VesselID)},
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
				query:  []int64{int64(suite.cfg.VesselID)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Set control. Ok",
			args: args{
				method: http.MethodPost,
				query:  []int64{int64(suite.cfg.VesselID)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
				query:  []string{strconv.FormatInt(int64(suite.cfg.VesselID), 10)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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

func (suite *HandlerTestSuite) TestDeleteControl() {
	t := suite.T()
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
			name: "Delete control. No jwt",
			args: args{
				method: http.MethodDelete,
				query:  []int64{int64(suite.cfg.VesselID)},
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
				query:  []int64{int64(suite.cfg.VesselID)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Delete control. User",
			args: args{
				method: http.MethodDelete,
				query:  []int64{int64(suite.cfg.VesselID)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
				query:  []string{strconv.FormatInt(int64(suite.cfg.VesselID), 10)},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
