// Package domain contains core business entities and value objects.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationSettings represents user notification preferences.
type NotificationSettings struct {
	JobCompleted    bool
	JobFailed       bool
	ProviderOffline bool
	UsageThreshold  bool
	WeeklySummary   bool
	MarketingEmails bool
}

// TenantSettings represents user preferences.
type TenantSettings struct {
	DarkMode      bool
	Notifications NotificationSettings
}

// DefaultTenantSettings returns default settings for a new tenant.
func DefaultTenantSettings() TenantSettings {
	return TenantSettings{
		DarkMode: true,
		Notifications: NotificationSettings{
			JobCompleted:    true,
			JobFailed:       true,
			ProviderOffline: true,
			UsageThreshold:  false,
			WeeklySummary:   false,
			MarketingEmails: false,
		},
	}
}

// Tenant represents an organization or user with API access.
type Tenant struct {
	ID              uuid.UUID
	Name            string
	APIKey          string
	Active          bool
	DefaultProvider string
	Settings        TenantSettings
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewTenant creates a new tenant.
func NewTenant(name, apiKey string) *Tenant {
	now := time.Now()
	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		APIKey:    apiKey,
		Active:    true,
		Settings:  DefaultTenantSettings(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Deactivate marks the tenant as inactive.
func (t *Tenant) Deactivate() {
	t.Active = false
	t.UpdatedAt = time.Now()
}
