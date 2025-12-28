// Package adapters contains repository implementations.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ingvar/aiaggregator/infrastructure/postgres/db"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// PostgresJobRepository implements JobRepository using PostgreSQL.
type PostgresJobRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

var _ usecases.JobRepository = (*PostgresJobRepository)(nil)

// NewPostgresJobRepository creates a new PostgresJobRepository.
func NewPostgresJobRepository(pool *pgxpool.Pool) *PostgresJobRepository {
	return &PostgresJobRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *PostgresJobRepository) Create(ctx context.Context, job *domain.Job) error {
	created, err := r.queries.CreateJob(ctx, db.CreateJobParams{
		TenantID: uuidToPgtype(job.TenantID),
		Type:     string(job.Type),
		Input:    job.Input,
		Status:   string(job.Status),
	})
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}

	// Update job with generated values
	job.ID = pgtypeToUUID(created.ID)
	job.CreatedAt = created.CreatedAt.Time
	job.UpdatedAt = created.UpdatedAt.Time
	return nil
}

func (r *PostgresJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	dbJob, err := r.queries.GetJob(ctx, uuidToPgtype(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrJobNotFound
		}
		return nil, fmt.Errorf("get job: %w", err)
	}
	return dbJobToDomain(&dbJob), nil
}

func (r *PostgresJobRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error) {
	dbJobs, err := r.queries.ListJobsByTenant(ctx, db.ListJobsByTenantParams{
		TenantID: uuidToPgtype(tenantID),
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	jobs := make([]*domain.Job, len(dbJobs))
	for i, j := range dbJobs {
		jobs[i] = dbJobToDomain(&j)
	}
	return jobs, nil
}

func (r *PostgresJobRepository) Update(ctx context.Context, job *domain.Job) error {
	_, err := r.queries.UpdateJob(ctx, db.UpdateJobParams{
		ID:         uuidToPgtype(job.ID),
		Status:     string(job.Status),
		Result:     textToPgtype(job.Result),
		Error:      textToPgtype(job.Error),
		Provider:   textToPgtype(job.Provider),
		TokensIn:   intToPgtype(job.TokensIn),
		TokensOut:  intToPgtype(job.TokensOut),
		Cost:       floatToNumeric(job.Cost),
		StartedAt:  timeToPgtype(job.StartedAt),
		FinishedAt: timeToPgtype(job.FinishedAt),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrJobNotFound
		}
		return fmt.Errorf("update job: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteJob(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	count, err := r.queries.CountJobsByTenant(ctx, uuidToPgtype(tenantID))
	if err != nil {
		return 0, fmt.Errorf("count jobs: %w", err)
	}
	return int(count), nil
}

// PostgresTenantRepository implements TenantRepository using PostgreSQL.
type PostgresTenantRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

var _ usecases.TenantRepository = (*PostgresTenantRepository)(nil)

// NewPostgresTenantRepository creates a new PostgresTenantRepository.
func NewPostgresTenantRepository(pool *pgxpool.Pool) *PostgresTenantRepository {
	return &PostgresTenantRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *PostgresTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	created, err := r.queries.CreateTenant(ctx, db.CreateTenantParams{
		Name:   tenant.Name,
		ApiKey: tenant.APIKey,
		Active: pgtype.Bool{Bool: tenant.Active, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create tenant: %w", err)
	}

	tenant.ID = pgtypeToUUID(created.ID)
	tenant.CreatedAt = created.CreatedAt.Time
	tenant.UpdatedAt = created.UpdatedAt.Time
	return nil
}

func (r *PostgresTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	dbTenant, err := r.queries.GetTenant(ctx, uuidToPgtype(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return dbTenantToDomain(&dbTenant), nil
}

func (r *PostgresTenantRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error) {
	dbTenant, err := r.queries.GetTenantByAPIKey(ctx, apiKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("get tenant by api key: %w", err)
	}
	return dbTenantToDomain(&dbTenant), nil
}

func (r *PostgresTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(map[string]interface{}{
		"darkMode": tenant.Settings.DarkMode,
		"notifications": map[string]bool{
			"jobCompleted":    tenant.Settings.Notifications.JobCompleted,
			"jobFailed":       tenant.Settings.Notifications.JobFailed,
			"providerOffline": tenant.Settings.Notifications.ProviderOffline,
			"usageThreshold":  tenant.Settings.Notifications.UsageThreshold,
			"weeklySummary":   tenant.Settings.Notifications.WeeklySummary,
			"marketingEmails": tenant.Settings.Notifications.MarketingEmails,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	_, err = r.queries.UpdateTenant(ctx, db.UpdateTenantParams{
		ID:              uuidToPgtype(tenant.ID),
		Name:            tenant.Name,
		Active:          pgtype.Bool{Bool: tenant.Active, Valid: true},
		DefaultProvider: pgtype.Text{String: tenant.DefaultProvider, Valid: tenant.DefaultProvider != ""},
		Settings:        settingsJSON,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrTenantNotFound
		}
		return fmt.Errorf("update tenant: %w", err)
	}
	return nil
}

func (r *PostgresTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteTenant(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	return nil
}

func (r *PostgresTenantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	dbTenants, err := r.queries.ListTenants(ctx, db.ListTenantsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}

	tenants := make([]*domain.Tenant, len(dbTenants))
	for i, t := range dbTenants {
		tenants[i] = dbTenantToDomain(&t)
	}
	return tenants, nil
}

// PostgresUsageRepository implements UsageRepository using PostgreSQL.
type PostgresUsageRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

var _ usecases.UsageRepository = (*PostgresUsageRepository)(nil)

// NewPostgresUsageRepository creates a new PostgresUsageRepository.
func NewPostgresUsageRepository(pool *pgxpool.Pool) *PostgresUsageRepository {
	return &PostgresUsageRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *PostgresUsageRepository) Record(ctx context.Context, usage *domain.Usage) error {
	created, err := r.queries.CreateUsage(ctx, db.CreateUsageParams{
		TenantID:  uuidToPgtype(usage.TenantID),
		JobID:     uuidToPgtype(usage.JobID),
		Provider:  usage.Provider,
		Model:     pgtype.Text{String: usage.Model, Valid: usage.Model != ""},
		TokensIn:  intToPgtype(usage.TokensIn),
		TokensOut: intToPgtype(usage.TokensOut),
		Cost:      floatToNumeric(usage.Cost),
	})
	if err != nil {
		return fmt.Errorf("record usage: %w", err)
	}

	usage.ID = pgtypeToUUID(created.ID)
	usage.CreatedAt = created.CreatedAt.Time
	return nil
}

func (r *PostgresUsageRepository) Create(ctx context.Context, usage *domain.Usage) error {
	return r.Record(ctx, usage)
}

func (r *PostgresUsageRepository) GetByJobID(ctx context.Context, jobID uuid.UUID) (*domain.Usage, error) {
	dbUsage, err := r.queries.GetUsageByJobID(ctx, uuidToPgtype(jobID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get usage by job id: %w", err)
	}
	return dbUsageToDomain(&dbUsage), nil
}

func (r *PostgresUsageRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Usage, error) {
	dbUsages, err := r.queries.ListUsageByTenant(ctx, db.ListUsageByTenantParams{
		TenantID: uuidToPgtype(tenantID),
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list usage: %w", err)
	}

	usages := make([]*domain.Usage, len(dbUsages))
	for i, u := range dbUsages {
		usages[i] = dbUsageToDomain(&u)
	}
	return usages, nil
}

func (r *PostgresUsageRepository) GetSummaryByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error) {
	dbSummaries, err := r.queries.GetUsageSummaryByTenant(ctx, uuidToPgtype(tenantID))
	if err != nil {
		return nil, fmt.Errorf("get usage summary: %w", err)
	}

	summaries := make([]*domain.UsageSummary, len(dbSummaries))
	for i, s := range dbSummaries {
		summaries[i] = &domain.UsageSummary{
			Provider:       s.Provider,
			TotalTokensIn:  int(s.TotalTokensIn),
			TotalTokensOut: int(s.TotalTokensOut),
			TotalCost:      numericToFloat(s.TotalCost),
			JobCount:       int(s.JobCount),
		}
	}
	return summaries, nil
}

func (r *PostgresUsageRepository) GetSummary(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error) {
	return r.GetSummaryByTenant(ctx, tenantID)
}

// Helper conversion functions

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

func textToPgtype(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgtypeToText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func intToPgtype(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

func pgtypeToInt(i pgtype.Int4) int {
	if !i.Valid {
		return 0
	}
	return int(i.Int32)
}

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	n.Scan(fmt.Sprintf("%.6f", f))
	return n
}

func floatPtrToNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	return floatToNumeric(*f)
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}

func timeToPgtype(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// Domain conversion functions

func dbJobToDomain(j *db.Job) *domain.Job {
	job := &domain.Job{
		ID:        pgtypeToUUID(j.ID),
		TenantID:  pgtypeToUUID(j.TenantID),
		Type:      domain.JobType(j.Type),
		Input:     j.Input,
		Status:    domain.JobStatus(j.Status),
		Result:    pgtypeToText(j.Result),
		Error:     pgtypeToText(j.Error),
		Provider:  pgtypeToText(j.Provider),
		TokensIn:  pgtypeToInt(j.TokensIn),
		TokensOut: pgtypeToInt(j.TokensOut),
		Cost:      numericToFloat(j.Cost),
		CreatedAt: j.CreatedAt.Time,
		UpdatedAt: j.UpdatedAt.Time,
	}
	if j.StartedAt.Valid {
		job.StartedAt = &j.StartedAt.Time
	}
	if j.FinishedAt.Valid {
		job.FinishedAt = &j.FinishedAt.Time
	}
	return job
}

func dbTenantToDomain(t *db.Tenant) *domain.Tenant {
	tenant := &domain.Tenant{
		ID:        pgtypeToUUID(t.ID),
		Name:      t.Name,
		APIKey:    t.ApiKey,
		Active:    t.Active.Bool,
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
		Settings:  domain.DefaultTenantSettings(),
	}

	if t.DefaultProvider.Valid {
		tenant.DefaultProvider = t.DefaultProvider.String
	}

	// Parse settings from JSONB
	if len(t.Settings) > 0 {
		var settings struct {
			DarkMode      bool `json:"darkMode"`
			Notifications struct {
				JobCompleted    bool `json:"jobCompleted"`
				JobFailed       bool `json:"jobFailed"`
				ProviderOffline bool `json:"providerOffline"`
				UsageThreshold  bool `json:"usageThreshold"`
				WeeklySummary   bool `json:"weeklySummary"`
				MarketingEmails bool `json:"marketingEmails"`
			} `json:"notifications"`
		}
		if err := json.Unmarshal(t.Settings, &settings); err == nil {
			tenant.Settings.DarkMode = settings.DarkMode
			tenant.Settings.Notifications.JobCompleted = settings.Notifications.JobCompleted
			tenant.Settings.Notifications.JobFailed = settings.Notifications.JobFailed
			tenant.Settings.Notifications.ProviderOffline = settings.Notifications.ProviderOffline
			tenant.Settings.Notifications.UsageThreshold = settings.Notifications.UsageThreshold
			tenant.Settings.Notifications.WeeklySummary = settings.Notifications.WeeklySummary
			tenant.Settings.Notifications.MarketingEmails = settings.Notifications.MarketingEmails
		}
	}

	return tenant
}

func dbUsageToDomain(u *db.Usage) *domain.Usage {
	model := ""
	if u.Model.Valid {
		model = u.Model.String
	}
	return &domain.Usage{
		ID:        pgtypeToUUID(u.ID),
		TenantID:  pgtypeToUUID(u.TenantID),
		JobID:     pgtypeToUUID(u.JobID),
		Provider:  u.Provider,
		Model:     model,
		TokensIn:  pgtypeToInt(u.TokensIn),
		TokensOut: pgtypeToInt(u.TokensOut),
		Cost:      numericToFloat(u.Cost),
		CreatedAt: u.CreatedAt.Time,
	}
}

// PostgresPricingRepository implements PricingRepository using PostgreSQL.
type PostgresPricingRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

var _ usecases.PricingRepository = (*PostgresPricingRepository)(nil)

// NewPostgresPricingRepository creates a new PostgresPricingRepository.
func NewPostgresPricingRepository(pool *pgxpool.Pool) *PostgresPricingRepository {
	return &PostgresPricingRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *PostgresPricingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProviderPricing, error) {
	pricing, err := r.queries.GetPricing(ctx, uuidToPgtype(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get pricing: %w", err)
	}
	return dbPricingToDomain(&pricing), nil
}

func (r *PostgresPricingRepository) GetByProviderModel(ctx context.Context, provider, model string) (*domain.ProviderPricing, error) {
	pricing, err := r.queries.GetPricingByProviderModel(ctx, db.GetPricingByProviderModelParams{
		Provider: provider,
		Model:    model,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get pricing by provider/model: %w", err)
	}
	return dbPricingToDomain(&pricing), nil
}

func (r *PostgresPricingRepository) GetDefaultByProvider(ctx context.Context, provider string) (*domain.ProviderPricing, error) {
	pricing, err := r.queries.GetDefaultPricingByProvider(ctx, provider)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get default pricing: %w", err)
	}
	return dbPricingToDomain(&pricing), nil
}

func (r *PostgresPricingRepository) List(ctx context.Context) ([]*domain.ProviderPricing, error) {
	pricings, err := r.queries.ListPricing(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pricing: %w", err)
	}

	result := make([]*domain.ProviderPricing, len(pricings))
	for i, p := range pricings {
		result[i] = dbPricingToDomain(&p)
	}
	return result, nil
}

func (r *PostgresPricingRepository) ListByProvider(ctx context.Context, provider string) ([]*domain.ProviderPricing, error) {
	pricings, err := r.queries.ListPricingByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("list pricing by provider: %w", err)
	}

	result := make([]*domain.ProviderPricing, len(pricings))
	for i, p := range pricings {
		result[i] = dbPricingToDomain(&p)
	}
	return result, nil
}

func (r *PostgresPricingRepository) Create(ctx context.Context, pricing *domain.ProviderPricing) error {
	created, err := r.queries.CreatePricing(ctx, db.CreatePricingParams{
		Provider:              pricing.Provider,
		Model:                 pricing.Model,
		InputPricePerMillion:  floatToNumeric(pricing.InputPricePerMillion),
		OutputPricePerMillion: floatToNumeric(pricing.OutputPricePerMillion),
		ImagePrice:            floatPtrToNumeric(pricing.ImagePrice),
		IsDefault:             pricing.IsDefault,
	})
	if err != nil {
		return fmt.Errorf("create pricing: %w", err)
	}

	pricing.ID = pgtypeToUUID(created.ID)
	pricing.CreatedAt = created.CreatedAt.Time
	pricing.UpdatedAt = created.UpdatedAt.Time
	return nil
}

func (r *PostgresPricingRepository) Update(ctx context.Context, pricing *domain.ProviderPricing) error {
	_, err := r.queries.UpdatePricing(ctx, db.UpdatePricingParams{
		ID:                    uuidToPgtype(pricing.ID),
		InputPricePerMillion:  floatToNumeric(pricing.InputPricePerMillion),
		OutputPricePerMillion: floatToNumeric(pricing.OutputPricePerMillion),
		ImagePrice:            floatPtrToNumeric(pricing.ImagePrice),
		IsDefault:             pricing.IsDefault,
	})
	if err != nil {
		return fmt.Errorf("update pricing: %w", err)
	}
	return nil
}

func (r *PostgresPricingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeletePricing(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete pricing: %w", err)
	}
	return nil
}

func dbPricingToDomain(p *db.ProviderPricing) *domain.ProviderPricing {
	var imagePrice *float64
	if p.ImagePrice.Valid {
		val := numericToFloat(p.ImagePrice)
		imagePrice = &val
	}

	return &domain.ProviderPricing{
		ID:                    pgtypeToUUID(p.ID),
		Provider:              p.Provider,
		Model:                 p.Model,
		InputPricePerMillion:  numericToFloat(p.InputPricePerMillion),
		OutputPricePerMillion: numericToFloat(p.OutputPricePerMillion),
		ImagePrice:            imagePrice,
		IsDefault:             p.IsDefault,
		CreatedAt:             p.CreatedAt.Time,
		UpdatedAt:             p.UpdatedAt.Time,
	}
}
