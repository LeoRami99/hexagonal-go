package application

import (
	"context"
	"fmt"

	entclient "hexagonal-go-backend/internal/ent"
	entpermission "hexagonal-go-backend/internal/ent/permission"
	entprojectmembership "hexagonal-go-backend/internal/ent/projectmembership"
	entprojectrole "hexagonal-go-backend/internal/ent/projectrole"
	enttenantmembership "hexagonal-go-backend/internal/ent/tenantmembership"
	entuser "hexagonal-go-backend/internal/ent/user"
)

type Authorizer interface {
	Can(ctx context.Context, userID, projectID, permissionCode string) (bool, error)
	IsSuperAdmin(ctx context.Context, userID string) (bool, error)
}

type AuthzService struct {
	client *entclient.Client
}

func NewAuthzService(client *entclient.Client) *AuthzService {
	return &AuthzService{client: client}
}

// IsSuperAdmin checks if a user has platform-level super administrator privileges.
func (s *AuthzService) IsSuperAdmin(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	u, err := s.client.User.Query().
		Where(entuser.IDEQ(userID)).
		Select(entuser.FieldIsSuperAdmin).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("check super admin for user %s: %w", userID, err)
	}
	return u.IsSuperAdmin, nil
}

// Can evaluates if a user is authorized to perform an atomic action within a project.
// Follows the rules defined in README_LMS_SAAS:
// 1. Superadmin has platform-wide bypass.
// 2. Evaluates active project memberships and their assigned roles' permissions.
func (s *AuthzService) Can(ctx context.Context, userID, projectID, permissionCode string) (bool, error) {
	if userID == "" || permissionCode == "" {
		return false, nil
	}

	// 1. Superadmin tiene acceso global irrestricto
	isSuper, err := s.IsSuperAdmin(ctx, userID)
	if err != nil {
		return false, err
	}
	if isSuper {
		return true, nil
	}

	// 2. Si no es superadmin, se requiere un project_id para contextualizar
	if projectID == "" {
		return false, nil
	}

	// 3. Consultar si el usuario tiene membresía activa en el proyecto
	// y al menos un rol habilitado que contenga el permiso atómico
	hasPermission, err := s.client.ProjectMembership.Query().
		Where(
			entprojectmembership.ProjectIDEQ(projectID),
			entprojectmembership.StatusEQ(entprojectmembership.StatusACTIVE),
			entprojectmembership.HasTenantMembershipWith(
				enttenantmembership.UserIDEQ(userID),
				enttenantmembership.StatusEQ(enttenantmembership.StatusACTIVE),
			),
			entprojectmembership.HasRolesWith(
				entprojectrole.IsEnabledEQ(true),
				entprojectrole.HasPermissionsWith(
					entpermission.CodeEQ(permissionCode),
				),
			),
		).
		Exist(ctx)

	if err != nil {
		return false, fmt.Errorf("evaluate permission %s for user %s in project %s: %w", permissionCode, userID, projectID, err)
	}

	return hasPermission, nil
}
