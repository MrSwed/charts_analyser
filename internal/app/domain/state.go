package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Control struct {
	State        bool       `json:"control" db:"control"`
	ControlStart *time.Time `json:"controlStart" db:"control_start"`
	ControlEnd   *time.Time `json:"controlEnd" db:"control_end"`
}

type Track struct {
	Timestamp time.Time `json:"timestamp" db:"time"`
	Location  Point     `json:"location" db:"location"`
	Vessel
}

type CurrentZone struct {
	Zones  []ZoneName `json:"zones" db:"zones"`
	TimeIn time.Time  `json:"timeIn" db:"time_in"`
}

func (v CurrentZone) Value() (driver.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (v *CurrentZone) Scan(src interface{}) error {
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

type VesselState struct {
	Control
	Vessel
	Timestamp    *time.Time   `json:"timestamp" db:"timestamp"`
	Location     *Point       `json:"location" db:"location"`
	CurrentZone  *CurrentZone `json:"currentZone" db:"current_zone"`
	ZoneDuration *Duration    `json:"zoneDuration" db:"zone_duration"`
}
type Duration time.Duration

func (d *Duration) Scan(raw interface{}) error {
	if raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case string:
		vv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		*d = Duration(time.Duration(vv) * time.Second)
	case float64:
		*d = Duration(time.Duration(v) * time.Second)
	case int64:
		*d = Duration(time.Duration(v) * time.Second)
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Duration from: %#v", v)
	}

	return nil
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d Duration) MarshalJSON() (b []byte, err error) {
	return json.Marshal(d.String())
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
	if src == nil {
		return nil
	}
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
