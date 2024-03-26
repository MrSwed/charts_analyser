package domain

import "time"

type ControlLog struct {
	*Vessel
	Timestamp time.Time `db:"Timestamp"`
	Control   bool      `db:"control"`
	Comment   *string   `db:"comment"`
}
