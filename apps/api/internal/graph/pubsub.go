// Package graph provides GraphQL resolvers and pub/sub functionality.
package graph

import (
	"sync"

	"github.com/ingvar/aiaggregator/packages/pubsub"
)

// PubSub manages job update subscriptions.
type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan *Job]struct{}
}

// NewPubSub creates a new PubSub instance.
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string]map[chan *Job]struct{}),
	}
}

// Subscribe adds a channel to receive job updates for a tenant.
func (ps *PubSub) Subscribe(tenantID string, ch chan *Job) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.subscribers[tenantID] == nil {
		ps.subscribers[tenantID] = make(map[chan *Job]struct{})
	}
	ps.subscribers[tenantID][ch] = struct{}{}
}

// SubscribeToJob adds a channel to receive updates for a specific job.
func (ps *PubSub) SubscribeToJob(jobID string, ch chan *Job) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	key := "job:" + jobID
	if ps.subscribers[key] == nil {
		ps.subscribers[key] = make(map[chan *Job]struct{})
	}
	ps.subscribers[key][ch] = struct{}{}
}

// Unsubscribe removes a subscription.
func (ps *PubSub) Unsubscribe(tenantID string, ch chan *Job) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.subscribers[tenantID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(ps.subscribers, tenantID)
		}
	}
}

// UnsubscribeFromJob removes a job-specific subscription.
func (ps *PubSub) UnsubscribeFromJob(jobID string, ch chan *Job) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	key := "job:" + jobID
	if subs, ok := ps.subscribers[key]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(ps.subscribers, key)
		}
	}
}

// Publish sends a job update to all subscribers for that tenant and job.
func (ps *PubSub) Publish(tenantID string, job *Job) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Notify tenant subscribers
	if subs, ok := ps.subscribers[tenantID]; ok {
		for ch := range subs {
			select {
			case ch <- job:
			default:
				// Channel full, skip
			}
		}
	}

	// Notify job-specific subscribers
	key := "job:" + job.ID
	if subs, ok := ps.subscribers[key]; ok {
		for ch := range subs {
			select {
			case ch <- job:
			default:
				// Channel full, skip
			}
		}
	}
}

// HandleRedisUpdate converts a Redis update to a GraphQL Job and publishes it.
func (ps *PubSub) HandleRedisUpdate(update *pubsub.JobUpdate) {
	job := &Job{
		ID:         update.JobID,
		TenantID:   update.TenantID,
		Type:       convertJobType(update.Type),
		Input:      update.Input,
		Status:     convertJobStatus(update.Status),
		Result:     update.Result,
		Error:      update.Error,
		Provider:   update.Provider,
		TokensIn:   update.TokensIn,
		TokensOut:  update.TokensOut,
		Cost:       update.Cost,
		CreatedAt:  update.CreatedAt,
		UpdatedAt:  update.UpdatedAt,
		StartedAt:  update.StartedAt,
		FinishedAt: update.FinishedAt,
	}

	ps.Publish(update.TenantID, job)
}

// HandleRedisUpdateWithUsage converts a Redis update to a GraphQL Job and publishes it,
// and also triggers usage update if job is completed.
func (ps *PubSub) HandleRedisUpdateWithUsage(update *pubsub.JobUpdate, getUsage func(tenantID string) []*UsageSummary) {
	ps.HandleRedisUpdate(update)

	// If job completed or failed, notify usage subscribers
	if update.Status == "completed" || update.Status == "failed" {
		if UsagePubSub.HasSubscribers(update.TenantID) {
			if usage := getUsage(update.TenantID); usage != nil {
				UsagePubSub.Publish(update.TenantID, usage)
			}
		}
	}
}

// convertJobStatus converts domain status (lowercase) to GraphQL status (UPPERCASE).
func convertJobStatus(status string) JobStatus {
	switch status {
	case "pending":
		return JobStatusPending
	case "processing":
		return JobStatusProcessing
	case "completed":
		return JobStatusCompleted
	case "failed":
		return JobStatusFailed
	default:
		return JobStatus(status)
	}
}

// convertJobType converts domain type (lowercase) to GraphQL type (UPPERCASE).
func convertJobType(jobType string) JobType {
	switch jobType {
	case "text":
		return JobTypeText
	case "image":
		return JobTypeImage
	default:
		return JobType(jobType)
	}
}

// JobPubSub is the global pub/sub instance for job updates.
var JobPubSub = NewPubSub()

// UsagePubSubType manages usage update subscriptions.
type UsagePubSubType struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []*UsageSummary]struct{}
}

// NewUsagePubSub creates a new UsagePubSub instance.
func NewUsagePubSub() *UsagePubSubType {
	return &UsagePubSubType{
		subscribers: make(map[string]map[chan []*UsageSummary]struct{}),
	}
}

// Subscribe adds a channel to receive usage updates for a tenant.
func (ps *UsagePubSubType) Subscribe(tenantID string, ch chan []*UsageSummary) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.subscribers[tenantID] == nil {
		ps.subscribers[tenantID] = make(map[chan []*UsageSummary]struct{})
	}
	ps.subscribers[tenantID][ch] = struct{}{}
}

// Unsubscribe removes a subscription.
func (ps *UsagePubSubType) Unsubscribe(tenantID string, ch chan []*UsageSummary) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.subscribers[tenantID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(ps.subscribers, tenantID)
		}
	}
}

// Publish sends usage updates to all subscribers for that tenant.
func (ps *UsagePubSubType) Publish(tenantID string, usage []*UsageSummary) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.subscribers[tenantID]; ok {
		for ch := range subs {
			select {
			case ch <- usage:
			default:
				// Channel full, skip
			}
		}
	}
}

// HasSubscribers checks if there are any subscribers for a tenant.
func (ps *UsagePubSubType) HasSubscribers(tenantID string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	subs, ok := ps.subscribers[tenantID]
	return ok && len(subs) > 0
}

// UsagePubSub is the global pub/sub instance for usage updates.
var UsagePubSub = NewUsagePubSub()
