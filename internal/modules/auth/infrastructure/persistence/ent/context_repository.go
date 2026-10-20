package ent

import (
	"context"
	"fmt"

	entclient "hexagonal-go-backend/internal/ent"
	entprojectmembership "hexagonal-go-backend/internal/ent/projectmembership"
	enttenant "hexagonal-go-backend/internal/ent/tenants"
	enttenantmembership "hexagonal-go-backend/internal/ent/tenantmembership"
	entuser "hexagonal-go-backend/internal/ent/user"
	authapp "hexagonal-go-backend/internal/modules/auth/application"
)

type AuthContextRepository struct {
	client *entclient.Client
}

func NewAuthContextRepository(client *entclient.Client) *AuthContextRepository {
	return &AuthContextRepository{client: client}
}

func (r *AuthContextRepository) GetSessionContext(ctx context.Context, userID string) (*authapp.SessionContext, error) {
	u, err := r.client.User.Query().
		Where(entuser.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	avatar := ""
	if u.AvatarURL != nil {
		avatar = *u.AvatarURL
	}

	userSummary := authapp.UserInfo{
		ID:           u.ID,
		AuthID:       u.AuthID,
		Name:         u.Name,
		Email:        u.Email,
		AvatarURL:    avatar,
		IsSuperAdmin: u.IsSuperAdmin,
	}

	// 1. Tenancy resolution
	var tenantList []authapp.TenantContext
	if u.IsSuperAdmin {
		allTenants, err := r.client.Tenants.Query().
			Where(enttenant.StatusEQ("ACTIVE")).
			All(ctx)
		if err == nil {
			tenantList = make([]authapp.TenantContext, len(allTenants))
			for i, t := range allTenants {
				tenantList[i] = authapp.TenantContext{
					ID:     t.ID,
					Name:   t.Name,
					Slug:   t.Slug,
					Status: string(t.Status),
				}
			}
		}
	} else {
		tms, err := r.client.TenantMembership.Query().
			Where(
				enttenantmembership.UserIDEQ(userID),
				enttenantmembership.StatusEQ("ACTIVE"),
			).
			WithTenant().
			All(ctx)
		if err == nil {
			tenantList = make([]authapp.TenantContext, 0, len(tms))
			for _, tm := range tms {
				if tm.Edges.Tenant != nil {
					t := tm.Edges.Tenant
					tenantList = append(tenantList, authapp.TenantContext{
						ID:     t.ID,
						Name:   t.Name,
						Slug:   t.Slug,
						Status: string(t.Status),
					})
				}
			}
		}
	}

	// 2. Project/Campus memberships with roles
	pms, err := r.client.ProjectMembership.Query().
		Where(
			entprojectmembership.HasTenantMembershipWith(
				enttenantmembership.UserIDEQ(userID),
			),
			entprojectmembership.StatusEQ("ACTIVE"),
		).
		WithProject().
		WithRoles().
		All(ctx)

	var projectList []authapp.ProjectMembershipContext
	if err == nil {
		projectList = make([]authapp.ProjectMembershipContext, 0, len(pms))
		for _, pm := range pms {
			if pm.Edges.Project != nil {
				proj := pm.Edges.Project
				roles := make([]string, len(pm.Edges.Roles))
				for i, r := range pm.Edges.Roles {
					roles[i] = r.Code
				}
				projectList = append(projectList, authapp.ProjectMembershipContext{
					ProjectID:   proj.ID,
					ProjectName: proj.Name,
					ProjectSlug: proj.Slug,
					TenantID:    proj.TenantID,
					Roles:       roles,
				})
			}
		}
	}

	return &authapp.SessionContext{
		User:     userSummary,
		Tenants:  tenantList,
		Projects: projectList,
	}, nil
}
