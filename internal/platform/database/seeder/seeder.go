package seeder

import (
	"context"
	"fmt"
	"log/slog"

	entclient "hexagonal-go-backend/internal/ent"
	entfeature "hexagonal-go-backend/internal/ent/feature"
	entpermission "hexagonal-go-backend/internal/ent/permission"
	entprojectrole "hexagonal-go-backend/internal/ent/projectrole"
	"hexagonal-go-backend/internal/platform/identifier"
)

type PermissionDef struct {
	Code        string
	Module      string
	Description string
}

type FeatureDef struct {
	Code        string
	Name        string
	Description string
	IsCore      bool
}

// SeedPlatform seeds global atomic permissions and feature catalog into the database.
func SeedPlatform(ctx context.Context, client *entclient.Client, log *slog.Logger) error {
	log.Info("seeding platform catalog (permissions & features)...")

	// 1. Catálogo global de permisos atómicos estables
	permissions := []PermissionDef{
		// Tenants & Projects
		{Code: "tenant.view", Module: "tenants", Description: "Ver información de la organización"},
		{Code: "tenant.update", Module: "tenants", Description: "Editar configuración de la organización"},
		{Code: "project.view", Module: "projects", Description: "Ver información del proyecto"},
		{Code: "project.settings.update", Module: "projects", Description: "Modificar configuración y branding del proyecto"},
		{Code: "project.members.manage", Module: "projects", Description: "Gestionar miembros del proyecto"},
		{Code: "project.roles.manage", Module: "projects", Description: "Administrar roles del proyecto"},

		// Catálogo y Contenido
		{Code: "course.view", Module: "courses", Description: "Ver cursos y lecciones"},
		{Code: "course.create", Module: "courses", Description: "Crear nuevos cursos"},
		{Code: "course.update", Module: "courses", Description: "Editar contenidos y versiones de cursos"},
		{Code: "course.publish", Module: "courses", Description: "Publicar versiones de cursos"},
		{Code: "course.delete", Module: "courses", Description: "Eliminar o archivar cursos"},
		{Code: "category.manage", Module: "catalog", Description: "Gestionar categorías de cursos"},
		{Code: "learning_path.manage", Module: "catalog", Description: "Gestionar rutas de aprendizaje"},

		// Operación Académica
		{Code: "instance.manage", Module: "delivery", Description: "Crear y gestionar instancias/cohortes de cursos"},
		{Code: "instance.staff.assign", Module: "delivery", Description: "Asignar docentes y monitores a instancias"},
		{Code: "enrollment.create", Module: "enrollment", Description: "Inscribir estudiantes en cursos"},
		{Code: "enrollment.view", Module: "enrollment", Description: "Ver inscripciones de estudiantes"},
		{Code: "enrollment.cancel", Module: "enrollment", Description: "Cancelar o retirar inscripciones"},
		{Code: "progress.view", Module: "progress", Description: "Consultar progreso de estudiantes"},
		{Code: "progress.record", Module: "progress", Description: "Registrar avance en lecciones"},

		// Evaluaciones y Tareas
		{Code: "quiz.manage", Module: "assessment", Description: "Crear y configurar quizzes"},
		{Code: "quiz.review", Module: "assessment", Description: "Revisar intentos y respuestas de quizzes"},
		{Code: "quiz.attempt", Module: "assessment", Description: "Presentar intentos de quizzes"},
		{Code: "assignment.manage", Module: "assignment", Description: "Configurar actividades y tareas"},
		{Code: "assignment.review", Module: "assignment", Description: "Calificar y retroalimentar entregas"},
		{Code: "assignment.submit", Module: "assignment", Description: "Enviar entregas de tareas"},

		// Certificados y Comunidad
		{Code: "certificate.template.manage", Module: "certificate", Description: "Gestionar plantillas de certificados"},
		{Code: "certificate.issue", Module: "certificate", Description: "Emitir certificados a estudiantes"},
		{Code: "certificate.view", Module: "certificate", Description: "Ver y descargar certificados"},
		{Code: "community.moderate", Module: "community", Description: "Moderar preguntas y respuestas en comunidad"},
		{Code: "community.post", Module: "community", Description: "Publicar preguntas y respuestas"},
		{Code: "community.view", Module: "community", Description: "Ver comunidad y foros"},
	}

	for _, pDef := range permissions {
		exists, err := client.Permission.Query().Where(entpermission.CodeEQ(pDef.Code)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("check permission %s: %w", pDef.Code, err)
		}
		if !exists {
			_, err = client.Permission.Create().
				SetID(identifier.New()).
				SetCode(pDef.Code).
				SetModule(pDef.Module).
				SetDescription(pDef.Description).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create permission %s: %w", pDef.Code, err)
			}
		}
	}

	// 2. Catálogo global de Features del SaaS
	features := []FeatureDef{
		{Code: "courses", Name: "Cursos y Contenido", Description: "Módulo central de cursos, versiones y lecciones", IsCore: true},
		{Code: "learning_paths", Name: "Rutas de Aprendizaje", Description: "Trayectorias ordenadas de cursos", IsCore: false},
		{Code: "quizzes", Name: "Cuestionarios y Quizzes", Description: "Evaluaciones automatizadas con puntaje mínimo", IsCore: false},
		{Code: "certificates", Name: "Certificados Verificables", Description: "Generación y verificación pública de certificados", IsCore: false},
		{Code: "assignments", Name: "Actividades y Tareas", Description: "Entregas con calificación de docentes", IsCore: false},
		{Code: "community", Name: "Comunidad y Foros", Description: "Preguntas y respuestas contextualizadas en instancias", IsCore: false},
	}

	for _, fDef := range features {
		exists, err := client.Feature.Query().Where(entfeature.CodeEQ(fDef.Code)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("check feature %s: %w", fDef.Code, err)
		}
		if !exists {
			_, err = client.Feature.Create().
				SetID(identifier.New()).
				SetCode(fDef.Code).
				SetName(fDef.Name).
				SetDescription(fDef.Description).
				SetIsCore(fDef.IsCore).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create feature %s: %w", fDef.Code, err)
			}
		}
	}

	log.Info("platform catalog seeded successfully")
	return nil
}

// ProvisionDefaultProjectRoles ensures a project has its default system roles with their standard permissions.
func ProvisionDefaultProjectRoles(ctx context.Context, client *entclient.Client, projectID string) error {
	// Mapeo de permisos por rol de sistema
	roleMatrix := map[string]struct {
		Name        string
		Description string
		Permissions []string
	}{
		"admin": {
			Name:        "Administrador",
			Description: "Administrador del campus con control operativo completo",
			Permissions: []string{
				"project.view", "project.settings.update", "project.members.manage", "project.roles.manage",
				"course.view", "course.create", "course.update", "course.publish", "course.delete",
				"category.manage", "learning_path.manage",
				"instance.manage", "instance.staff.assign",
				"enrollment.create", "enrollment.view", "enrollment.cancel",
				"progress.view",
				"quiz.manage", "quiz.review",
				"assignment.manage", "assignment.review",
				"certificate.template.manage", "certificate.issue", "certificate.view",
				"community.moderate", "community.post", "community.view",
			},
		},
		"teacher": {
			Name:        "Docente",
			Description: "Profesor con permisos de gestión de contenidos, calificaciones y revisión",
			Permissions: []string{
				"course.view", "course.update",
				"instance.manage",
				"enrollment.view",
				"progress.view",
				"quiz.manage", "quiz.review",
				"assignment.manage", "assignment.review",
				"certificate.view",
				"community.moderate", "community.post", "community.view",
			},
		},
		"monitor": {
			Name:        "Monitor / Auxiliar",
			Description: "Tutor o auxiliar con permisos de soporte, dudas y asistencia en revisiones",
			Permissions: []string{
				"course.view",
				"enrollment.view",
				"progress.view",
				"quiz.review",
				"assignment.review",
				"community.moderate", "community.post", "community.view",
			},
		},
		"student": {
			Name:        "Estudiante",
			Description: "Alumno participante con acceso a sus cursos, tareas y certificados",
			Permissions: []string{
				"course.view",
				"progress.view", "progress.record",
				"quiz.attempt",
				"assignment.submit",
				"certificate.view",
				"community.post", "community.view",
			},
		},
	}

	for code, rDef := range roleMatrix {
		roleEntity, err := client.ProjectRole.Query().
			Where(entprojectrole.ProjectIDEQ(projectID), entprojectrole.CodeEQ(code)).
			Only(ctx)

		if err != nil && !entclient.IsNotFound(err) {
			return fmt.Errorf("check project role %s: %w", code, err)
		}

		if entclient.IsNotFound(err) {
			// Buscar entidades de permisos
			perms, err := client.Permission.Query().
				Where(entpermission.CodeIn(rDef.Permissions...)).
				All(ctx)
			if err != nil {
				return fmt.Errorf("query permissions for role %s: %w", code, err)
			}

			_, err = client.ProjectRole.Create().
				SetID(identifier.New()).
				SetProjectID(projectID).
				SetCode(code).
				SetName(rDef.Name).
				SetDescription(rDef.Description).
				SetIsSystem(true).
				SetIsEnabled(true).
				AddPermissions(perms...).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create project role %s: %w", code, err)
			}
		} else {
			// Asegurar permisos
			perms, err := client.Permission.Query().
				Where(entpermission.CodeIn(rDef.Permissions...)).
				All(ctx)
			if err != nil {
				return fmt.Errorf("query permissions for role %s: %w", code, err)
			}
			err = roleEntity.Update().
				AddPermissions(perms...).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("update permissions for project role %s: %w", code, err)
			}
		}
	}

	return nil
}
