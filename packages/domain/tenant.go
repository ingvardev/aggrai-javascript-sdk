// Package domain contains core business entities and value objects.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents an organization or user with API access.
type Tenant struct {
	ID        uuid.UUID
	Name      string
	APIKey    string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTenant creates a new tenant.
func NewTenant(name, apiKey string) *Tenant {
	now := time.Now()
	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		APIKey:    apiKey,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Deactivate marks the tenant as inactive.
func (t *Tenant) Deactivate() {
	t.Active = false
	t.UpdatedAt = time.Now()
}
