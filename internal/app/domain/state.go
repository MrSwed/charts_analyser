package domain

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

type LocationPoint struct {
	Location Point `json:"location" db:"location"`
}

type Track struct {
	Timestamp time.Time `json:"timestamp"`
	LocationPoint
	Vessel
}

type CurrentZone struct {
	Zones  []ZoneName `json:"zones" db:"zones"`
	TimeIn time.Time  `json:"timeIn" db:"time_in"`
}

func (v CurrentZone) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

type VesselState struct {
	Timestamp    time.Time `json:"timestamp"`
	Location     Point     `json:"location"`
	Vessel       Vessel    `json:"vessel"`
	CurrentZone  `json:"currentZone"`
	ZoneDuration string `json:"zoneDuration"`
}

func (v VesselState) MarshalBinary() ([]byte, error) {
	v.ZoneTimeSet()
	return json.Marshal(v)
}
func (v VesselState) Value() (driver.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (v *VesselState) Scan(src interface{}) error {
	var source []byte
	if reflect.ValueOf(src).Kind() == reflect.String {
		source = []byte(src.(string))
	} else {
		source = src.([]byte)
	}
	err := json.Unmarshal(source, v)
	if err != nil {
		return err
	}
	v.ZoneTimeSet()
	return nil
}

func (v *VesselState) SetFromMap(m map[string]string) (err error) {
	// refactor to key-to-key if it will be slow
	var b []byte
	if b, err = json.Marshal(m); err != nil {
		return
	}
	err = json.Unmarshal(b, &v)
	return
}

func (v *VesselState) ZoneTimeSet() {

	v.ZoneDuration = v.Timestamp.Sub(v.CurrentZone.TimeIn).String()

}

// Point 	(0 - lon, 1 - ltd)
type Point [2]float64

func (v Point) Value() (driver.Value, error) {
	strLon := strconv.FormatFloat(v[0], 'g', -1, 64)
	strLtd := strconv.FormatFloat(v[1], 'g', -1, 64)
	b := "SRID=4326;POINT(" + strLon + " " + strLtd + ")"
	return b, nil
}

func (v *Point) Scan(src interface{}) error {
	var source []byte
	if reflect.ValueOf(src).Kind() == reflect.String {
		source = []byte(src.(string))
	} else {
		source = src.([]byte)
	}
	err := json.Unmarshal(source, v)
	if err != nil {
		return err
	}
	return nil
}

func (v Point) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}
