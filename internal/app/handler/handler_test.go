// to test with real db set env DATABASE_DSN before run with created, but empty tables
// to test with file - set env FILE_STORAGE_PATH
package handler

import (
	"charts_analyser/internal/app/config"
	"context"
	"github.com/redis/go-redis/v9"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	conf     = config.NewConfig().WithEnv()
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

/* todo: need refactor tests for fiber * /
func TestZones(t *testing.T) {
	repo := repository.NewRepository(db, redisCli)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
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
		responseLen bool
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
				code:        http.StatusOK,
				responseLen: true,
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
				responseLen: false,
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
				var data []domain.ZoneName
				err = json.Unmarshal(resBody, &data)
				require.NoError(t, err)
				require.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
				if test.want.responseLen {
					assert.Greater(t, len(data), 0)
				} else {
					assert.Equal(t, len(data), 0)
				}
			}
		})
	}
}

func TestVessels(t *testing.T) {
	repo := repository.NewRepository(db, redisCli)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
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
		responseLen bool
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
				code:        http.StatusOK,
				responseLen: true,
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
				responseLen: false,
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
				var data []domain.VesselID
				err = json.Unmarshal(resBody, &data)
				require.NoError(t, err)
				require.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
				if test.want.responseLen {
					assert.Greater(t, len(data), 0)
				} else {
					assert.Equal(t, len(data), 0)
				}
			}
		})
	}
}
/**/
/*todo tests
MonitoredList
SetControl
DelControl
VesselState
Track
*/

/**/
