// to test with real db set env DATABASE_DSN before run with created, but empty tables
// to test with file - set env FILE_STORAGE_PATH
package handler

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
	"testing"
	"time"

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

func TestZones(t *testing.T) {
	repo := repository.NewRepository(db, redisCli)
	logger, _ := zap.NewDevelopment()
	s := service.NewService(repo, logger)
	app := fiber.New()
	app.Use(recover.New())

	h := NewHandler(app, s, conf, logger)
	api := app.Group(constant.RouteApi)
	api.Get(constant.RouteZones, h.Zones())

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
		query  map[string]interface{}
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
				query: map[string]interface{}{
					"vesselIDs": []int64{vesselId},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: true,
				contentType: "application/json",
			},
		},
		{
			name: "Get vessel 'Federal Saguenay' zones (id as string), bad Request",
			args: args{
				method: http.MethodGet,
				query: map[string]interface{}{
					"vesselIDs": []string{strconv.FormatInt(vesselId, 10)},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get unknown vessels zones",
			args: args{
				method: http.MethodGet,
				query: map[string]interface{}{
					"vesselIDs": []int64{10000000000},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: false,
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJson, _ := json.Marshal(&test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteApi+constant.RouteZones, bytes.NewReader(bodyJson))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			res, err := app.Test(request)
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
	app := fiber.New()
	app.Use(recover.New())

	h := NewHandler(app, s, conf, logger)
	api := app.Group(constant.RouteApi)
	api.Get(constant.RouteVessels, h.Vessels())

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
		query  map[string]interface{}
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
				query: map[string]interface{}{
					"zoneNames": []string{zoneName},
					"start":     timeStart.Format(time.RFC3339),
					"finish":    timeEnd.Format(time.RFC3339),
				},
			},
			want: want{
				code:        http.StatusOK,
				responseLen: true,
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
			},
			want: want{
				code:        http.StatusOK,
				responseLen: false,
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
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bodyJson, _ := json.Marshal(&test.args.query)
			request, err := http.NewRequest(test.args.method, constant.RouteApi+constant.RouteVessels, bytes.NewReader(bodyJson))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			res, err := app.Test(request)
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
