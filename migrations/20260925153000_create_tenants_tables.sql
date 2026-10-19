-- Create "tenants" table
CREATE TABLE IF NOT EXISTS "tenants" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "slug" character varying NOT NULL,
  "contact_email" character varying NULL,
  "status" character varying NOT NULL DEFAULT 'ACTIVE',
  "settings" jsonb NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);

-- Create index "tenant_status" to table: "tenants"
CREATE UNIQUE INDEX IF NOT EXISTS "tenants_slug" ON "tenants" ("slug");
CREATE INDEX IF NOT EXISTS "tenants_status" ON "tenants" ("status");

-- Create "tenant_memberships" table
CREATE TABLE IF NOT EXISTS "tenant_memberships" (
  "id" character varying NOT NULL,
  "tenant_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'ACTIVE',
  "joined_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tenant_membership_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_tenant_membership_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "tenant_membership_tenant_user" ON "tenant_memberships" ("tenant_id", "user_id");
CREATE INDEX IF NOT EXISTS "tenant_membership_status" ON "tenant_memberships" ("status");

-- Create "projects" table
CREATE TABLE IF NOT EXISTS "projects" (
  "id" character varying NOT NULL,
  "tenant_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "slug" character varying NOT NULL,
  "description" text NULL,
  "status" character varying NOT NULL DEFAULT 'ACTIVE',
  "domains" jsonb NULL,
  "custom_domain" character varying NULL,
  "logo_url" character varying NULL,
  "timezone" character varying NOT NULL DEFAULT 'America/Bogota',
  "locale" character varying NOT NULL DEFAULT 'es',
  "branding" jsonb NULL,
  "settings" jsonb NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_project_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "project_tenant_slug" ON "projects" ("tenant_id", "slug");
CREATE INDEX IF NOT EXISTS "project_status" ON "projects" ("status");
CREATE INDEX IF NOT EXISTS "project_custom_domain" ON "projects" ("custom_domain");

-- Create "project_memberships" table
CREATE TABLE IF NOT EXISTS "project_memberships" (
  "id" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  "tenant_membership_id" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'ACTIVE',
  "joined_at" timestamptz NOT NULL,
  "last_activity_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_project_membership_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_project_membership_tenant_member" FOREIGN KEY ("tenant_membership_id") REFERENCES "tenant_memberships" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "project_membership_project_tenant_member" ON "project_memberships" ("project_id", "tenant_membership_id");
CREATE INDEX IF NOT EXISTS "project_membership_status" ON "project_memberships" ("status");

-- Create "project_roles" table
CREATE TABLE IF NOT EXISTS "project_roles" (
  "id" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" text NULL,
  "is_system" boolean NOT NULL DEFAULT false,
  "is_enabled" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_project_role_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "project_role_project_code" ON "project_roles" ("project_id", "code");
CREATE INDEX IF NOT EXISTS "project_role_is_system" ON "project_roles" ("is_system");
CREATE INDEX IF NOT EXISTS "project_role_is_enabled" ON "project_roles" ("is_enabled");

-- Create "project_role_permissions" join table
CREATE TABLE IF NOT EXISTS "project_role_permissions" (
  "project_role_id" character varying NOT NULL,
  "permission_id" character varying NOT NULL,
  PRIMARY KEY ("project_role_id", "permission_id"),
  CONSTRAINT "fk_project_role_permissions_role" FOREIGN KEY ("project_role_id") REFERENCES "project_roles" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_project_role_permissions_permission" FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON DELETE CASCADE
);

-- Create "project_membership_roles" join table
CREATE TABLE IF NOT EXISTS "project_membership_roles" (
  "project_membership_id" character varying NOT NULL,
  "project_role_id" character varying NOT NULL,
  PRIMARY KEY ("project_membership_id", "project_role_id"),
  CONSTRAINT "fk_project_membership_roles_member" FOREIGN KEY ("project_membership_id") REFERENCES "project_memberships" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_project_membership_roles_role" FOREIGN KEY ("project_role_id") REFERENCES "project_roles" ("id") ON DELETE CASCADE
);

-- Create "project_features" table
CREATE TABLE IF NOT EXISTS "project_features" (
  "id" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  "feature_id" character varying NOT NULL,
  "is_enabled" boolean NOT NULL DEFAULT true,
  "settings" jsonb NULL,
  "enabled_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_project_feature_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_project_feature_feature" FOREIGN KEY ("feature_id") REFERENCES "features" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "project_feature_project_feature" ON "project_features" ("project_id", "feature_id");
CREATE INDEX IF NOT EXISTS "project_feature_is_enabled" ON "project_features" ("is_enabled");
