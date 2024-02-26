package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"errors"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"sync"
)

type MonitorCache struct {
	rds *redis.Client
	mu  sync.Mutex
}

func NewMonitorRepository(rds *redis.Client) *MonitorCache {
	return &MonitorCache{rds: rds}
}

func redisKeys(vesselIDs ...domain.VesselID) []string {
	s := make([]string, 0, len(vesselIDs))
	for _, v := range vesselIDs {
		s = append(s, constant.RedisVeselPrefix+v.String())
	}
	return s
}

func (r *MonitorCache) IsMonitored(ctx context.Context, vesselId domain.VesselID) (state bool, err error) {
	state, err = r.rds.SIsMember(ctx, constant.RedisControlIds, vesselId.String()).Result()
	return
}

func (r *MonitorCache) SetControl(ctx context.Context, control bool, vesselsItems ...*domain.Vessel) (err error) {
	if len(vesselsItems) == 0 {
		return errors.New("no vessels for control")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	vessels := domain.Vessels(vesselsItems)
	veselIDs := vessels.InterfacesIDs()
	veselNames := vessels.StringAr()
	if control {
		if _, err = r.rds.SAdd(ctx, constant.RedisControlIds, veselIDs...).Result(); err != nil {
			return
		}
		if _, err = r.rds.SAdd(ctx, constant.RedisControlVessels, veselNames...).Result(); err != nil {
			return
		}
	} else {
		if _, err = r.rds.SRem(ctx, constant.RedisControlIds, veselIDs...).Result(); err != nil {
			return
		}
		if _, err = r.rds.SRem(ctx, constant.RedisControlVessels, veselNames...).Result(); err != nil {
			return
		}
	}
	return
}

func (r *MonitorCache) GetState(ctx context.Context, vesselId domain.VesselID) (state *domain.VesselState, err error) {
	var m string
	if m, err = r.rds.Get(ctx, redisKeys(vesselId)[0]).Result(); err != nil {
		return
	}
	if len(m) > 0 {
		state = new(domain.VesselState)
		err = json.Unmarshal([]byte(m), &state)
	}
	return
}

func (r *MonitorCache) UpdateState(ctx context.Context, vesselId domain.VesselID, v *domain.VesselState) (err error) {
	if v == nil {
		err = errors.New("updateState: input data nil")
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err = r.rds.Set(ctx, redisKeys(vesselId)[0], v, 0).Result()
	return
}

func (r *MonitorCache) MonitoredVessels(ctx context.Context) (vessels domain.Vessels, err error) {
	var strIDs []string
	if strIDs, err = r.rds.SMembers(ctx, constant.RedisControlVessels).Result(); err != nil {
		return
	}
	err = vessels.FromStrings(strIDs...)
	return
}
