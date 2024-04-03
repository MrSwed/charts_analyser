package handler_test

import (
	"bytes"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func (suite *HandlerTestSuite) TestAddVessel() {
	t := suite.T()
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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

func (suite *HandlerTestSuite) TestDeleteVessel() {
	t := suite.T()
	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := suite.srv.Vessel.AddVessel(ctx, newVessels...)
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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
			name: "Delete vessel. No body data",
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
			name: "Delete vessel. Bad body data.",
			args: args{
				method: http.MethodDelete,
				body:   []interface{}{12.12, "vessel name", 1555444},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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

func (suite *HandlerTestSuite) TestRestoreVessel() {
	t := suite.T()
	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := suite.srv.Vessel.AddVessel(ctx, newVessels...)
	require.NoError(t, err)
	err = suite.srv.Vessel.SetDeleteVessels(ctx, true, vessels.IDs()...)
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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
			name: "Restore vessel. No body data",
			args: args{
				method: http.MethodPatch,
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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

func (suite *HandlerTestSuite) TestGetVessel() {
	t := suite.T()
	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := suite.srv.Vessel.AddVessel(ctx, newVessels...)
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Get vessel. testdata",
			args: args{
				method: http.MethodGet,
				query:  map[string]interface{}{"vesselIDs": suite.cfg.VesselID},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
				},
			},
			want: want{
				code:            http.StatusOK,
				responseLen:     &[]bool{true}[0],
				responseContain: suite.cfg.VesselID.String(),
				contentType:     "application/json",
			},
		},
		{
			name: "Get vessel. created",
			args: args{
				method: http.MethodGet,
				query:  map[string]interface{}{"vesselIDs": vessels.IDs()},
				headers: map[string]string{
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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

func (suite *HandlerTestSuite) TestUpdateVessel() {
	t := suite.T()
	timeID := time.Now().Format(time.RFC3339Nano)
	newVessels := []domain.VesselName{domain.VesselName("Test1 Vessel_" + timeID), domain.VesselName("Test2 Vessel_" + timeID)}

	ctx := context.Background()
	vessels, err := suite.srv.Vessel.AddVessel(ctx, newVessels...)
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
					"Authorization": "Bearer " + suite.cfg.jwtVessel,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
					"Authorization": "Bearer " + suite.cfg.jwtOperator,
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
