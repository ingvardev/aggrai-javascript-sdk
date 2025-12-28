package graph

import (
	"encoding/base64"

	"github.com/ingvar/aiaggregator/packages/domain"
)

// domainJobToGraphQL converts a domain Job to a GraphQL Job
func domainJobToGraphQL(job *domain.Job) *Job {
	if job == nil {
		return nil
	}

	var status JobStatus
	switch job.Status {
	case domain.JobStatusPending:
		status = JobStatusPending
	case domain.JobStatusProcessing:
		status = JobStatusProcessing
	case domain.JobStatusCompleted:
		status = JobStatusCompleted
	case domain.JobStatusFailed:
		status = JobStatusFailed
	default:
		status = JobStatusPending
	}

	var jobType JobType
	switch job.Type {
	case domain.JobTypeText:
		jobType = JobTypeText
	case domain.JobTypeImage:
		jobType = JobTypeImage
	default:
		jobType = JobTypeText
	}

	gqlJob := &Job{
		ID:         job.ID.String(),
		TenantID:   job.TenantID.String(),
		Type:       jobType,
		Input:      job.Input,
		Status:     status,
		Result:     job.Result,
		Error:      job.Error,
		Provider:   job.Provider,
		TokensIn:   job.TokensIn,
		TokensOut:  job.TokensOut,
		Cost:       job.Cost,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
		StartedAt:  job.StartedAt,
		FinishedAt: job.FinishedAt,
	}

	return gqlJob
}

// encodeCursor creates a cursor string from job ID
func encodeCursor(jobID string) string {
	return base64.StdEncoding.EncodeToString([]byte(jobID))
}

// domainTenantToGraphQL converts a domain Tenant to a GraphQL Tenant
func domainTenantToGraphQL(tenant *domain.Tenant) *Tenant {
	if tenant == nil {
		return nil
	}

	gqlTenant := &Tenant{
		ID:        tenant.ID.String(),
		Name:      tenant.Name,
		Active:    tenant.Active,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	if tenant.DefaultProvider != "" {
		gqlTenant.DefaultProvider = &tenant.DefaultProvider
	}

	// Convert settings
	gqlTenant.Settings = &TenantSettings{
		DarkMode: tenant.Settings.DarkMode,
		Notifications: &NotificationSettings{
			JobCompleted:    tenant.Settings.Notifications.JobCompleted,
			JobFailed:       tenant.Settings.Notifications.JobFailed,
			ProviderOffline: tenant.Settings.Notifications.ProviderOffline,
			UsageThreshold:  tenant.Settings.Notifications.UsageThreshold,
			WeeklySummary:   tenant.Settings.Notifications.WeeklySummary,
			MarketingEmails: tenant.Settings.Notifications.MarketingEmails,
		},
	}

	return gqlTenant
}

// domainPricingToGraphQL converts a domain ProviderPricing to a GraphQL ProviderPricing
func domainPricingToGraphQL(pricing *domain.ProviderPricing) *ProviderPricing {
	if pricing == nil {
		return nil
	}

	return &ProviderPricing{
		ID:                    pricing.ID.String(),
		Provider:              pricing.Provider,
		Model:                 pricing.Model,
		InputPricePerMillion:  pricing.InputPricePerMillion,
		OutputPricePerMillion: pricing.OutputPricePerMillion,
		ImagePrice:            pricing.ImagePrice,
		IsDefault:             pricing.IsDefault,
		CreatedAt:             pricing.CreatedAt,
		UpdatedAt:             pricing.UpdatedAt,
	}
}
