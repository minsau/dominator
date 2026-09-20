package registrar

import "time"

type Route struct {
	ProjectSlug  string    `json:"project_slug"`
	Slug         string    `json:"slug,omitempty"`
	Host         string    `json:"host"`
	Backend      string    `json:"backend"`
	AuthRequired bool      `json:"auth_required"`
	InternalOnly bool      `json:"internal_only"`
	UpdatedAt    time.Time `json:"updated_at"`
}

