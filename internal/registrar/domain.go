package registrar

import "time"

type Route struct {
	ProjectSlug  string    `json:"project_slug"`
	Host         string    `json:"host"`
	Backend      string    `json:"backend"`
	AuthRequired bool      `json:"auth_required"`
	UpdatedAt    time.Time `json:"updated_at"`
}
