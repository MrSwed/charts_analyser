package domain

import (
	"errors"
	"strconv"
	"strings"
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

type VesselName string

func (v *VesselName) String() string {
	return string(*v)
}

type Vessel struct {
	ID   VesselID
	Name VesselName
}

func (v *Vessel) String() string {
	return v.ID.String() + ":" + v.Name.String()
}

func (v *Vessel) FromString(s string) (err error) {
	sa := strings.SplitN(s, ":", 2)
	v.Name = VesselName(s[1])
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
