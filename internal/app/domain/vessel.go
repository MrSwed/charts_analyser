package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type VesselID int64

func (v *VesselID) String() string {
	return strconv.FormatInt(int64(*v), 10)
}

func (v *VesselID) SetFromStr(s string) (err error) {
	var f int64
	if f, err = strconv.ParseInt(s, 10, 64); err == nil {
		*v = VesselID(f)
	}
	return
}

type VesselIDs []VesselID

type VesselName string

func (v *VesselName) String() string {
	return string(*v)
}

type Vessel struct {
	ID   VesselID   `json:"id" db:"vessel_id"`
	Name VesselName `json:"name" db:"vessel_name"`
}

func (v *Vessel) String() string {
	return v.ID.String() + ":" + v.Name.String()
}

func (v *Vessel) FromString(s string) (err error) {
	sa := strings.SplitN(s, ":", 2)
	v.Name = VesselName(sa[1])
	err = v.ID.SetFromStr(sa[0])
	return
}

type Vessels []*Vessel

func (v *Vessels) FromStrings(s ...string) (err error) {
	*v = make([]*Vessel, 0, len(s))
	for _, vStr := range s {
		vStr := strings.TrimSpace(vStr)
		if len(vStr) == 0 {
			continue
		}
		iv := new(Vessel)
		if er := iv.FromString(vStr); er != nil {
			err = errors.Join(err, er)
			continue
		}
		*v = append(*v, iv)
	}
	return
}

func (v *Vessels) StringAr() []interface{} {
	a := make([]interface{}, 0, len(*v))
	for _, i := range *v {
		a = append(a, i.String())
	}
	return a
}

func (v *Vessels) IDsStringAr() []string {
	a := make([]string, 0, len(*v))
	for _, i := range *v {
		if i.ID == 0 {
			continue
		}
		a = append(a, i.ID.String())
	}
	return a
}

func (v *Vessels) InterfacesIDs() (ids []interface{}) {
	ids = make([]interface{}, 0, len(*v))
	for _, i := range *v {
		if i.ID == 0 {
			continue
		}
		ids = append(ids, int64(i.ID))
	}
	return
}

func (v *Vessels) IDs() (ids []VesselID) {
	ids = make([]VesselID, 0, len(*v))
	for _, i := range *v {
		if i.ID == 0 {
			continue
		}
		ids = append(ids, i.ID)
	}
	return
}

type VesselInfo struct {
	Vessel
	CreatedAt *time.Time `db:"created_at" json:"createdAt"`
}
