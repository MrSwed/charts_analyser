package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"github.com/redis/go-redis/v9"
)

type MonitorCache struct {
	rds *redis.Client
}

func NewMonitorRepository(rds *redis.Client) *MonitorCache {
	return &MonitorCache{rds: rds}
}

func redisKey(vesselId domain.VesselID) string {
	return constant.RedisVeselPrefix + vesselId.String()
}

func (r *MonitorCache) IsMonitored(ctx context.Context, vesselId domain.VesselID) (state bool, err error) {
	state, err = r.rds.SIsMember(ctx, constant.RedisControlIds, vesselId.String()).Result()
	return
}

func (r *MonitorCache) SetControl(ctx context.Context, vessel domain.Vessel, control bool) (err error) {
	if control {
		if _, err = r.rds.SAdd(ctx, constant.RedisControlIds, vessel.ID.String()).Result(); err != nil {
			return
		}
		if _, err = r.rds.SAdd(ctx, constant.RedisControlVessels, vessel.String()).Result(); err != nil {
			return
		}
		//_, err = r.rds.HSet(ctx, redisKey(vessel.ID), domain.VesselState{
		//	Status:       constant.StatusControl,
		//}).Result()
	} else {
		if _, err = r.rds.SRem(ctx, constant.RedisControlIds, vessel.ID.String()).Result(); err != nil {
			return
		}
		if _, err = r.rds.SRem(ctx, constant.RedisControlVessels, vessel.String()).Result(); err != nil {
			return
		}
		_, err = r.rds.Del(ctx, redisKey(vessel.ID)).Result()
	}
	return
}

func (r *MonitorCache) GetState(ctx context.Context, vesselId domain.VesselID) (state *domain.VesselState, err error) {
	m := make(map[string]string)
	if m, err = r.rds.HGetAll(ctx, redisKey(vesselId)).Result(); err != nil {
		return
	}
	err = state.SetFromMap(m)
	return
}

func (r *MonitorCache) UpdateState(ctx context.Context, vesselId domain.VesselID, v domain.VesselState) (err error) {
	_, err = r.rds.HSet(ctx, redisKey(vesselId), v).Result()
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
