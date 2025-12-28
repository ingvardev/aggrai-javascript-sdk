package graph

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/apps/api/internal/middleware"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// Auth helper functions

// getClientIP extracts client IP from context
func getClientIP(ctx context.Context) net.IP {
	if ip := ctx.Value("client_ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			return net.ParseIP(ipStr)
		}
	}
	return nil
}

// getUserAgent extracts user agent from context
func getUserAgent(ctx context.Context) string {
	if ua := ctx.Value("user_agent"); ua != nil {
		if uaStr, ok := ua.(string); ok {
			return uaStr
		}
	}
	return ""
}

// getSessionToken extracts session token from context
func getSessionToken(ctx context.Context) string {
	if token := ctx.Value("session_token"); token != nil {
		if tokenStr, ok := token.(string); ok {
			return tokenStr
		}
	}
	return ""
}

// domainOwnerToGraphQL converts domain.TenantOwner to GraphQL TenantOwner
func domainOwnerToGraphQL(owner *domain.TenantOwner) *TenantOwner {
	if owner == nil {
		return nil
	}

	role := OwnerRoleMember
	switch owner.Role {
	case domain.OwnerRoleOwner:
		role = OwnerRoleOwner
	case domain.OwnerRoleAdmin:
		role = OwnerRoleAdmin
	case domain.OwnerRoleMember:
		role = OwnerRoleMember
	}

	return &TenantOwner{
		ID:            owner.ID.String(),
		TenantID:      owner.TenantID.String(),
		Email:         owner.Email,
		Name:          owner.Name,
		Role:          role,
		Active:        owner.Active,
		EmailVerified: owner.EmailVerified,
		LastLoginAt:   owner.LastLoginAt,
		CreatedAt:     owner.CreatedAt,
		UpdatedAt:     owner.UpdatedAt,
	}
}

// domainSessionToGraphQL converts domain.OwnerSession to GraphQL Session
func domainSessionToGraphQL(session *domain.OwnerSession) *Session {
	if session == nil {
		return nil
	}

	var ipAddr *string
	if session.IPAddress != nil {
		s := session.IPAddress.String()
		ipAddr = &s
	}

	return &Session{
		ID:        session.ID.String(),
		UserAgent: &session.UserAgent,
		IPAddress: ipAddr,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}
}

// graphqlRoleToDomain converts GraphQL OwnerRole to domain.OwnerRole
func graphqlRoleToDomain(role OwnerRole) domain.OwnerRole {
	switch role {
	case OwnerRoleOwner:
		return domain.OwnerRoleOwner
	case OwnerRoleAdmin:
		return domain.OwnerRoleAdmin
	default:
		return domain.OwnerRoleMember
	}
}

// ========================================
// Auth Mutation Resolvers
// ========================================

// loginImpl implements the login mutation
func (r *mutationResolver) loginImpl(ctx context.Context, input LoginInput) (*AuthPayload, error) {
	if r.webAuthService == nil {
		return &AuthPayload{
			Success: false,
			Error:   strPtr("Authentication service not configured"),
		}, nil
	}

	result := r.webAuthService.Login(ctx, &usecases.LoginRequest{
		Email:     input.Email,
		Password:  input.Password,
		UserAgent: getUserAgent(ctx),
		IP:        getClientIP(ctx),
	})

	if !result.Success {
		errMsg := "Invalid credentials"
		if result.Error != nil {
			errMsg = result.Error.Error()
		}
		return &AuthPayload{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	return &AuthPayload{
		Success:      true,
		SessionToken: &result.SessionToken,
		Owner:        domainOwnerToGraphQL(result.Owner),
		Tenant:       domainTenantToGraphQL(result.Tenant),
	}, nil
}

// logoutImpl implements the logout mutation
func (r *mutationResolver) logoutImpl(ctx context.Context) (bool, error) {
	if r.webAuthService == nil {
		return false, nil
	}

	sessionToken := getSessionToken(ctx)
	if sessionToken == "" {
		return true, nil // Already logged out
	}

	err := r.webAuthService.Logout(ctx, sessionToken)
	if err != nil {
		return false, err
	}
	return true, nil
}

// logoutAllImpl implements the logoutAll mutation
func (r *mutationResolver) logoutAllImpl(ctx context.Context) (bool, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return false, nil
	}

	err := r.webAuthService.LogoutAll(ctx, webCtx.OwnerID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// registerImpl implements the register mutation
func (r *mutationResolver) registerImpl(ctx context.Context, input RegisterInput) (*AuthPayload, error) {
	if r.webAuthService == nil {
		return &AuthPayload{
			Success: false,
			Error:   strPtr("Authentication service not configured"),
		}, nil
	}

	// Create tenant first
	tenant := &domain.Tenant{
		ID:     uuid.New(),
		Name:   input.TenantName,
		Active: true,
	}
	if err := r.tenantRepo.Create(ctx, tenant); err != nil {
		return &AuthPayload{
			Success: false,
			Error:   strPtr("Failed to create tenant"),
		}, nil
	}

	// Create owner
	owner, err := r.webAuthService.CreateOwner(ctx, tenant.ID, input.Email, input.Password, input.Name, domain.OwnerRoleOwner)
	if err != nil {
		// Clean up tenant
		_ = r.tenantRepo.Delete(ctx, tenant.ID)
		errMsg := err.Error()
		return &AuthPayload{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	// Auto-login
	loginResult := r.webAuthService.Login(ctx, &usecases.LoginRequest{
		Email:     input.Email,
		Password:  input.Password,
		UserAgent: getUserAgent(ctx),
		IP:        getClientIP(ctx),
	})

	if !loginResult.Success {
		return &AuthPayload{
			Success: true, // Registration succeeded, but auto-login failed
			Owner:   domainOwnerToGraphQL(owner),
			Tenant:  domainTenantToGraphQL(tenant),
		}, nil
	}

	return &AuthPayload{
		Success:      true,
		SessionToken: &loginResult.SessionToken,
		Owner:        domainOwnerToGraphQL(owner),
		Tenant:       domainTenantToGraphQL(tenant),
	}, nil
}

// changePasswordImpl implements the changePassword mutation
func (r *mutationResolver) changePasswordImpl(ctx context.Context, input ChangePasswordInput) (bool, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return false, nil
	}

	// Verify current password first
	owner, err := r.webAuthService.GetOwnerByID(ctx, webCtx.OwnerID)
	if err != nil {
		return false, err
	}

	if !owner.CheckPassword(input.CurrentPassword) {
		return false, nil
	}

	// Change password
	err = r.webAuthService.ChangePassword(ctx, webCtx.OwnerID, input.NewPassword)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ========================================
// Owner Management Resolvers
// ========================================

// createOwnerImpl implements the createOwner mutation
func (r *mutationResolver) createOwnerImpl(ctx context.Context, input CreateOwnerInput) (*TenantOwner, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return nil, errUnauthorized
	}

	// Only owners and admins can create owners
	if !webCtx.HasPermission(domain.OwnerRoleAdmin) {
		return nil, errForbidden
	}

	role := graphqlRoleToDomain(input.Role)
	owner, err := r.webAuthService.CreateOwner(ctx, webCtx.TenantID, input.Email, input.Password, input.Name, role)
	if err != nil {
		return nil, err
	}

	return domainOwnerToGraphQL(owner), nil
}

// updateOwnerImpl implements the updateOwner mutation
func (r *mutationResolver) updateOwnerImpl(ctx context.Context, id string, input UpdateOwnerInput) (*TenantOwner, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return nil, errUnauthorized
	}

	ownerID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	owner, err := r.webAuthService.GetOwnerByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	// Check tenant isolation
	if owner.TenantID != webCtx.TenantID {
		return nil, errNotFound
	}

	// Only owners can update other owners; admins can update members
	if owner.Role == domain.OwnerRoleOwner && webCtx.Role != domain.OwnerRoleOwner {
		return nil, errForbidden
	}

	// Apply updates
	if input.Name != nil {
		owner.Name = *input.Name
	}
	if input.Role != nil {
		owner.Role = graphqlRoleToDomain(*input.Role)
	}
	if input.Active != nil {
		owner.Active = *input.Active
	}

	if err := r.webAuthService.UpdateOwner(ctx, owner); err != nil {
		return nil, err
	}

	return domainOwnerToGraphQL(owner), nil
}

// deleteOwnerImpl implements the deleteOwner mutation
func (r *mutationResolver) deleteOwnerImpl(ctx context.Context, id string) (bool, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return false, errUnauthorized
	}

	// Only owners can delete owners
	if webCtx.Role != domain.OwnerRoleOwner {
		return false, errForbidden
	}

	ownerID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Can't delete yourself
	if ownerID == webCtx.OwnerID {
		return false, errCannotDeleteSelf
	}

	// Check tenant isolation
	owner, err := r.webAuthService.GetOwnerByID(ctx, ownerID)
	if err != nil {
		return false, err
	}
	if owner.TenantID != webCtx.TenantID {
		return false, errNotFound
	}

	// Deactivate instead of hard delete
	owner.Active = false
	if err := r.webAuthService.UpdateOwner(ctx, owner); err != nil {
		return false, err
	}

	return true, nil
}

// ========================================
// Query Resolvers
// ========================================

// currentOwnerImpl implements the currentOwner query
func (r *queryResolver) currentOwnerImpl(ctx context.Context) (*TenantOwner, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return nil, nil // Not logged in
	}

	owner, err := r.webAuthService.GetOwnerByID(ctx, webCtx.OwnerID)
	if err != nil {
		return nil, err
	}

	return domainOwnerToGraphQL(owner), nil
}

// mySessionsImpl implements the mySessions query
func (r *queryResolver) mySessionsImpl(ctx context.Context) ([]*Session, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return []*Session{}, nil
	}

	// Note: Need to add GetSessionsByOwnerID to WebAuthService
	// For now return empty list
	return []*Session{}, nil
}

// tenantOwnersImpl implements the tenantOwners query
func (r *queryResolver) tenantOwnersImpl(ctx context.Context) ([]*TenantOwner, error) {
	webCtx := middleware.WebAuthContextFromContext(ctx)
	if webCtx == nil {
		return nil, errUnauthorized
	}

	// Only admins and owners can list owners
	if !webCtx.HasPermission(domain.OwnerRoleAdmin) {
		return nil, errForbidden
	}

	owners, err := r.webAuthService.GetOwnersByTenantID(ctx, webCtx.TenantID)
	if err != nil {
		return nil, err
	}

	result := make([]*TenantOwner, len(owners))
	for i, o := range owners {
		result[i] = domainOwnerToGraphQL(o)
	}

	return result, nil
}

// ========================================
// Helper functions
// ========================================

func strPtr(s string) *string {
	return &s
}

var (
	errUnauthorized      = &GraphQLError{Message: "unauthorized", Code: "UNAUTHORIZED"}
	errForbidden         = &GraphQLError{Message: "forbidden", Code: "FORBIDDEN"}
	errNotFound          = &GraphQLError{Message: "not found", Code: "NOT_FOUND"}
	errCannotDeleteSelf  = &GraphQLError{Message: "cannot delete yourself", Code: "CANNOT_DELETE_SELF"}
)

// GraphQLError is a custom error type for GraphQL
type GraphQLError struct {
	Message string
	Code    string
}

func (e *GraphQLError) Error() string {
	return e.Message
}
