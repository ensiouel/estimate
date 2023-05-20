package entity

import "time"

type Website struct {
	URL         string        `db:"url" json:"url"`
	LastCheckAt time.Time     `db:"last_check_at" json:"last_check_at"`
	AccessTime  time.Duration `db:"access_time" json:"access_time"`
	Available   bool          `db:"available" json:"available"`
}
