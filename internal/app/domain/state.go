package domain

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

type VesselState struct {
	Timestamp time.Time `json:"timestamp"`
	//Status       string        `json:"status"`
	Location     Point        `json:"location" db:"ST_AsGeoJSON(coordinate)::json->>'coordinates' as location"`
	Vessel       Vessel       `json:"vessel"`
	DateInterval DateInterval `json:"dateInterval"`
}

/*
func (v VesselState) Map() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
*/

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

// Point 	(0 - lon, 1 - ltd)
type Point [2]float64

func (v Point) Value() (driver.Value, error) {
	strLon := strconv.FormatFloat(v[0], 'g', -1, 64)
	strLtd := strconv.FormatFloat(v[1], 'g', -1, 64)
	b := "POINT(" + strLon + " " + strLtd + ")"
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
