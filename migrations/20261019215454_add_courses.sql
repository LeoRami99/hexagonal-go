-- Create index "features_code_key" to table: "features"
CREATE UNIQUE INDEX "features_code_key" ON "features" ("code");
-- Create "courses" table
CREATE TABLE "courses" ("id" character varying NOT NULL, "tenant_id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "project_id" character varying NOT NULL, "title" character varying NOT NULL, "slug" character varying NOT NULL, "summary" text NOT NULL, "description" text NOT NULL, "level" character varying NOT NULL, "thumbnail_url" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'draft', PRIMARY KEY ("id"), CONSTRAINT "courses_projects_courses" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "courses_level" to table: "courses"
CREATE INDEX "courses_level" ON "courses" ("level");
-- Create index "courses_project_id_slug" to table: "courses"
CREATE UNIQUE INDEX "courses_project_id_slug" ON "courses" ("project_id", "slug");
-- Create index "courses_status" to table: "courses"
CREATE INDEX "courses_status" ON "courses" ("status");
-- Create index "courses_tenant_id" to table: "courses"
CREATE INDEX "courses_tenant_id" ON "courses" ("tenant_id");
-- Drop index "project_tenant_slug" from table: "projects"
DROP INDEX "project_tenant_slug";
-- Modify "projects" table
ALTER TABLE "projects" DROP CONSTRAINT "fk_project_tenant", ADD CONSTRAINT "projects_tenants_projects" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "project_tenant_id_slug" to table: "projects"
CREATE UNIQUE INDEX "project_tenant_id_slug" ON "projects" ("tenant_id", "slug");
-- Drop index "project_feature_is_enabled" from table: "project_features"
DROP INDEX "project_feature_is_enabled";
-- Drop index "project_feature_project_feature" from table: "project_features"
DROP INDEX "project_feature_project_feature";
-- Modify "project_features" table
ALTER TABLE "project_features" DROP CONSTRAINT "fk_project_feature_feature", DROP CONSTRAINT "fk_project_feature_project", ADD CONSTRAINT "project_features_features_project_features" FOREIGN KEY ("feature_id") REFERENCES "features" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "project_features_projects_features" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "projectfeature_is_enabled" to table: "project_features"
CREATE INDEX "projectfeature_is_enabled" ON "project_features" ("is_enabled");
-- Create index "projectfeature_project_id_feature_id" to table: "project_features"
CREATE UNIQUE INDEX "projectfeature_project_id_feature_id" ON "project_features" ("project_id", "feature_id");
-- Create index "users_auth_id_key" to table: "users"
CREATE UNIQUE INDEX "users_auth_id_key" ON "users" ("auth_id");
-- Drop index "tenant_membership_status" from table: "tenant_memberships"
DROP INDEX "tenant_membership_status";
-- Drop index "tenant_membership_tenant_user" from table: "tenant_memberships"
DROP INDEX "tenant_membership_tenant_user";
-- Modify "tenant_memberships" table
ALTER TABLE "tenant_memberships" DROP CONSTRAINT "fk_tenant_membership_tenant", DROP CONSTRAINT "fk_tenant_membership_user", ADD CONSTRAINT "tenant_memberships_tenants_memberships" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "tenant_memberships_users_tenant_memberships" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "tenantmembership_status" to table: "tenant_memberships"
CREATE INDEX "tenantmembership_status" ON "tenant_memberships" ("status");
-- Create index "tenantmembership_tenant_id_user_id" to table: "tenant_memberships"
CREATE UNIQUE INDEX "tenantmembership_tenant_id_user_id" ON "tenant_memberships" ("tenant_id", "user_id");
-- Drop index "project_membership_project_tenant_member" from table: "project_memberships"
DROP INDEX "project_membership_project_tenant_member";
-- Drop index "project_membership_status" from table: "project_memberships"
DROP INDEX "project_membership_status";
-- Modify "project_memberships" table
ALTER TABLE "project_memberships" DROP CONSTRAINT "fk_project_membership_project", DROP CONSTRAINT "fk_project_membership_tenant_member", ADD CONSTRAINT "project_memberships_projects_memberships" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "project_memberships_tenant_memberships_project_memberships" FOREIGN KEY ("tenant_membership_id") REFERENCES "tenant_memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "projectmembership_project_id_tenant_membership_id" to table: "project_memberships"
CREATE UNIQUE INDEX "projectmembership_project_id_tenant_membership_id" ON "project_memberships" ("project_id", "tenant_membership_id");
-- Create index "projectmembership_status" to table: "project_memberships"
CREATE INDEX "projectmembership_status" ON "project_memberships" ("status");
-- Drop index "project_role_is_enabled" from table: "project_roles"
DROP INDEX "project_role_is_enabled";
-- Drop index "project_role_is_system" from table: "project_roles"
DROP INDEX "project_role_is_system";
-- Drop index "project_role_project_code" from table: "project_roles"
DROP INDEX "project_role_project_code";
-- Modify "project_roles" table
ALTER TABLE "project_roles" DROP CONSTRAINT "fk_project_role_project", ADD CONSTRAINT "project_roles_projects_roles" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "projectrole_is_enabled" to table: "project_roles"
CREATE INDEX "projectrole_is_enabled" ON "project_roles" ("is_enabled");
-- Create index "projectrole_is_system" to table: "project_roles"
CREATE INDEX "projectrole_is_system" ON "project_roles" ("is_system");
-- Create index "projectrole_project_id_code" to table: "project_roles"
CREATE UNIQUE INDEX "projectrole_project_id_code" ON "project_roles" ("project_id", "code");
-- Modify "project_membership_roles" table
ALTER TABLE "project_membership_roles" DROP CONSTRAINT "fk_project_membership_roles_member", DROP CONSTRAINT "fk_project_membership_roles_role", ADD CONSTRAINT "project_membership_roles_project_membership_id" FOREIGN KEY ("project_membership_id") REFERENCES "project_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "project_membership_roles_project_role_id" FOREIGN KEY ("project_role_id") REFERENCES "project_roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "permissions" table
ALTER TABLE "permissions" ALTER COLUMN "description" TYPE character varying;
-- Create index "permissions_code_key" to table: "permissions"
CREATE UNIQUE INDEX "permissions_code_key" ON "permissions" ("code");
-- Modify "project_role_permissions" table
ALTER TABLE "project_role_permissions" DROP CONSTRAINT "fk_project_role_permissions_permission", DROP CONSTRAINT "fk_project_role_permissions_role", ADD CONSTRAINT "project_role_permissions_permission_id" FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "project_role_permissions_project_role_id" FOREIGN KEY ("project_role_id") REFERENCES "project_roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Drop index "user_profile_user_id" from table: "user_profiles"
DROP INDEX "user_profile_user_id";
-- Modify "user_profiles" table
ALTER TABLE "user_profiles" DROP CONSTRAINT "fk_user_profile_user", ADD CONSTRAINT "user_profiles_users_profile" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "user_profiles_user_id_key" to table: "user_profiles"
CREATE UNIQUE INDEX "user_profiles_user_id_key" ON "user_profiles" ("user_id");
