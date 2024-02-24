// to test with real db set env DATABASE_DSN before run with created, but empty tables
// to test with file - set env FILE_STORAGE_PATH
package handler

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/repository"
	"charts_analyser/internal/app/service"
	"encoding/json"

	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	conf = config.NewConfig().WithEnv()
)

func TestZones(t *testing.T) {
	db, err := sqlx.Connect("postgres", conf.DatabaseDSN)
	if err != nil {
		require.NoError(t, err)
	}
	repo := repository.NewRepository(db)
	s := service.NewService(repo)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	vesselId := int64(9110913)
	timeStart, err := time.Parse("2006-01-02 03:04:05", `2017-01-08 00:00:00`)
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 03:04:05", `2020-10-09 00:00:00`)
	require.NoError(t, err)

	type want struct {
		code        int
		response    []string
		contentType string
	}
	type args struct {
		method string
		query  map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get vessel 'Federal Saguenay' zones",
			args: args{
				method: http.MethodGet,
				query: map[string]string{
					"vessel_id": strconv.FormatInt(vesselId, 10),
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code: http.StatusOK,
				response: []string{
					"zone_176",
					"zone_194",
					"zone_197",
					"zone_199",
					"zone_205",
					"zone_219",
					"zone_221",
					"zone_222",
					"zone_310",
					"zone_312",
					"zone_316",
					"zone_342",
				},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Get unknown vessels zones",
			args: args{
				method: http.MethodGet,
				query: map[string]string{
					"vessel_id": strconv.FormatInt(10000000000, 10),
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    []string{},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Get zones with bad parameters",
			args: args{
				method: http.MethodGet,
				query: map[string]string{
					"vessel_id": "string instead number",
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//b := new(bytes.Buffer)
			//err := json.NewEncoder(b).Encode(test.args.query)
			//require.NoError(t, err)

			req, err := http.NewRequest(test.args.method, ts.URL+constant.RouteApi+constant.RouteZones, nil)
			require.NoError(t, err)
			q := req.URL.Query()
			for k, v := range test.args.query {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			res, err := http.DefaultClient.Do(req)
			var resBody []byte

			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				var data []string
				err = json.Unmarshal(resBody, &data)
				assert.NoError(t, err)
				assert.Equal(t, test.want.response, data)
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestVessels(t *testing.T) {
	db, err := sqlx.Connect("postgres", conf.DatabaseDSN)
	if err != nil {
		require.NoError(t, err)
	}
	repo := repository.NewRepository(db)
	s := service.NewService(repo)
	logger, _ := zap.NewDevelopment()
	h := NewHandler(s, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	zoneName := "zone_205"
	timeStart, err := time.Parse("2006-01-02 03:04:05", `2017-01-08 00:00:00`)
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 03:04:05", `2020-10-09 00:00:00`)
	require.NoError(t, err)

	type want struct {
		code        int
		response    []int64
		contentType string
	}
	type args struct {
		method string
		query  map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get zone 'zone_205' vessels",
			args: args{
				method: http.MethodGet,
				query: map[string]string{
					"zone_name": zoneName,
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code: http.StatusOK,
				response: []int64{
					8130875,
					8130899,
					8512279,
					8617938,
					8902955,
					8902967,
					8918289,
					9046368,
					9108128,
					9108130,
					9110913,
					9110925,
					9112296,
					9117739,
					9118147,
					9143568,
					9164615,
					9174751,
					9186326,
					9186338,
					9192428,
					9200330,
					9200419,
					9200445,
					9205885,
					9205897,
					9205902,
					9205926,
					9205938,
					9218404,
					9218416,
					9222302,
					9229972,
					9229984,
					9229996,
					9230000,
					9233466,
					9235555,
					9238272,
					9241463,
					9243409,
					9244257,
					9256470,
					9261451,
					9267209,
					9271511,
					9271614,
					9272383,
					9274563,
					9278791,
					9280586,
					9282106,
					9282479,
					9282481,
					9283681,
					9284702,
				},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Get zone 'zone_XXX' vessels",
			args: args{
				method: http.MethodGet,
				query: map[string]string{
					"zone_name": "zone_XXX",
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    []int64{},
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "Get zones with bad parameters",
			args: args{
				method: http.MethodGet,
				query: map[string]string{
					"vessel_id": "string instead number",
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//b := new(bytes.Buffer)
			//err := json.NewEncoder(b).Encode(test.args.query)
			//require.NoError(t, err)

			req, err := http.NewRequest(test.args.method, ts.URL+constant.RouteApi+constant.RouteVessels, nil)
			require.NoError(t, err)
			q := req.URL.Query()
			for k, v := range test.args.query {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			res, err := http.DefaultClient.Do(req)
			var resBody []byte

			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.code == http.StatusOK {
				var data []int64
				err = json.Unmarshal(resBody, &data)
				assert.NoError(t, err)
				assert.Equal(t, test.want.response, data)
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
