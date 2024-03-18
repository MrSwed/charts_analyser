package handler_test

import (
	"bytes"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAddUser(t *testing.T) {

	timeID := time.Now().Format(time.RFC3339Nano)
	newUser := domain.UserChange{
		Login:    domain.UserLogin("Test1_user_" + timeID),
		Password: &[]domain.Password{"Pa$$w0rd"}[0],
		Role:     2,
	}
	existUser := domain.UserChange{
		Login:    domain.UserLogin("Test2_user_" + timeID),
		Password: &[]domain.Password{"Pa$$w0rd1"}[0],
		Role:     2,
	}
	ctx := context.Background()
	id, err := serv.AddUser(ctx, &existUser)
	require.NoError(t, err)
	existUser.ID = &id

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
			name: "Add user. No jwt",
			args: args{
				method: http.MethodPost,
				body:   newUser,
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Add user. Wrong role in jwt",
			args: args{
				method: http.MethodPost,
				body:   newUser,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Add user. OK",
			args: args{
				method: http.MethodPost,
				body:   newUser,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code:        http.StatusCreated,
				responseLen: &[]bool{true}[0],
				contentType: "application/json",
			},
		},
		{
			name: "Add user. No body data",
			args: args{
				method: http.MethodPost,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Add user. duplicate",
			args: args{
				method: http.MethodPost,
				body: domain.UserChange{
					Login:    existUser.Login,
					Password: existUser.Password,
					Role:     2,
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code: http.StatusConflict,
			},
		},
		{
			name: "Add user. Bad body data",
			args: args{
				method: http.MethodPost,
				body:   []interface{}{12.12, "user name"},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Add user. Validate. Bad password",
			args: args{
				method: http.MethodPost,
				body: domain.UserChange{
					Login:    domain.UserLogin("Test2_user_bad_pass" + timeID),
					Password: &[]domain.Password{"password"}[0],
					Role:     2,
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code:            http.StatusBadRequest,
				responseContain: "Password",
				contentType:     "text/plain",
			},
		}, {
			name: "Add user. Validate. Bad role",
			args: args{
				method: http.MethodPost,
				body: domain.UserChange{
					Login:    domain.UserLogin("Test3_user_bad_role" + timeID),
					Password: &[]domain.Password{"password"}[0],
					Role:     3,
				},

				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code:            http.StatusBadRequest,
				responseContain: "Role",
				contentType:     "text/plain",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteUser, bytes.NewReader(bodyJSON))
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

func TestDeleteUser(t *testing.T) {

	timeID := time.Now().Format(time.RFC3339Nano)
	existUser := domain.UserChange{
		Login:    domain.UserLogin("Test1_delete_user_" + timeID),
		Password: &[]domain.Password{"Pa$$w0rd"}[0],
		Role:     2,
	}
	ctx := context.Background()
	id, err := serv.AddUser(ctx, &existUser)
	require.NoError(t, err)
	existUser.ID = &id

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
			name: "Delete user. No jwt",
			args: args{
				method: http.MethodDelete,
				body:   []domain.UserID{*existUser.ID},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Delete user. Wrong role in jwt",
			args: args{
				method: http.MethodDelete,
				body:   []domain.UserID{*existUser.ID},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Delete user. OK",
			args: args{
				method: http.MethodDelete,
				body:   []domain.UserID{*existUser.ID},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Delete user. No body data",
			args: args{
				method: http.MethodDelete,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Delete user. Bad body data.",
			args: args{
				method: http.MethodDelete,
				body:   []interface{}{12.12, "user name", 1555444},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
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
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteUser, bytes.NewReader(bodyJSON))
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

func TestRestoreUser(t *testing.T) {

	timeID := time.Now().Format(time.RFC3339Nano)
	existUser := domain.UserChange{
		Login:    domain.UserLogin("Test1_delete_user_" + timeID),
		Password: &[]domain.Password{"Pa$$w0rd"}[0],
		Role:     2,
	}
	ctx := context.Background()
	id, err := serv.AddUser(ctx, &existUser)
	require.NoError(t, err)
	existUser.ID = &id
	err = serv.User.SetDeletedUser(ctx, true, id)
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
			name: "Restore user. No jwt",
			args: args{
				method: http.MethodPatch,
				body:   []domain.UserID{*existUser.ID},
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Restore user. Wrong role in jwt",
			args: args{
				method: http.MethodPatch,
				body:   []domain.UserID{*existUser.ID},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Restore user. OK",
			args: args{
				method: http.MethodPatch,
				body:   []domain.UserID{*existUser.ID},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Restore user. No body data",
			args: args{
				method: http.MethodPatch,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Restore user. Bad body data.",
			args: args{
				method: http.MethodPatch,
				body:   []interface{}{12.12, "user name", 1555444},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
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
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteUser, bytes.NewReader(bodyJSON))
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

func TestUpdateUser(t *testing.T) {

	timeID := time.Now().Format(time.RFC3339Nano)
	existUser := domain.UserChange{
		Login:    domain.UserLogin("Test2_user_" + timeID),
		Password: &[]domain.Password{"Pa$$w0rd1"}[0],
		Role:     2,
	}
	ctx := context.Background()
	id, err := serv.AddUser(ctx, &existUser)
	require.NoError(t, err)
	existUser.ID = &id

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
			name: "Update user. No jwt",
			args: args{
				method: http.MethodPut,
				body:   existUser,
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    &[]string{"Missing or malformed JWT"}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Update user. Wrong role in jwt",
			args: args{
				method: http.MethodPut,
				body:   existUser,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtOperator,
				},
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
		{
			name: "Update user. OK",
			args: args{
				method: http.MethodPut,
				body:   existUser,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Update user. Without password, OK",
			args: args{
				method: http.MethodPut,
				body: domain.UserChange{
					ID:    existUser.ID,
					Login: existUser.Login,
					Role:  2,
				},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: &[]bool{true}[0],
				contentType: "text/plain",
			},
		},
		{
			name: "Update user. No body data",
			args: args{
				method: http.MethodPut,
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Update user. Bad body data.",
			args: args{
				method: http.MethodPut,
				body:   map[string]interface{}{"userIDs": []interface{}{12.12, "user name", 1555444}},
				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Update user. Not exist.",
			args: args{
				method: http.MethodPut,
				body: domain.UserChange{
					Login:    domain.UserLogin("SomeNotExistUserLogin"),
					Password: &[]domain.Password{"Pa$$w0rd12"}[0],
					Role:     2,
					ID:       &[]domain.UserID{1005000000}[0],
				},

				headers: map[string]string{
					"Authorization": "Bearer " + conf.jwtAdmin,
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
			bodyJSON, _ := json.Marshal(test.args.body)
			request, err := http.NewRequest(test.args.method, constant.RouteAPI+constant.RouteUser, bytes.NewReader(bodyJSON))
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
