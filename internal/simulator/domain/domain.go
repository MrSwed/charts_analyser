package domain

import (
	"charts_analyser/internal/app/domain"
	"time"
)

type Track struct {
	Timestamp time.Time    `json:"timestamp" db:"time"`
	Location  domain.Point `json:"location" db:"location"`
}

type VesselItem struct {
	domain.Vessel
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	historyCache []Track
}

func (v *VesselItem) Tracks() []Track {
	return v.historyCache
}

func (v *VesselItem) TrackShift() (t Track) {
	if len(v.historyCache) > 0 {
		t = v.historyCache[0]
		v.historyCache = v.historyCache[1:]
	}
	return
}

func (v *VesselItem) TrackAdd(t ...Track) {
	v.historyCache = append(v.historyCache, t...)
}

type HistoryDate time.Time

func (d HistoryDate) Now() (c time.Time) {
	nowTime := time.Now()
	c = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(),
		time.Time(d).Hour(), time.Time(d).Minute(), time.Time(d).Second(), 0, time.UTC)
	return
}
