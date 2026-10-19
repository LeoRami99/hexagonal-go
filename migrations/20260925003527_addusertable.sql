-- Create "users" table
CREATE TABLE IF NOT EXISTS "users" (
  "id" character varying NOT NULL,
  "auth_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "email" character varying NOT NULL,
  "avatar_url" character varying NULL,
  "is_super_admin" boolean NOT NULL DEFAULT false,
  "status" character varying NOT NULL DEFAULT 'ACTIVE',
  "active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "user_created_at" ON "users" ("created_at");
CREATE UNIQUE INDEX IF NOT EXISTS "user_email" ON "users" ("email");
CREATE UNIQUE INDEX IF NOT EXISTS "user_auth_id" ON "users" ("auth_id");
CREATE INDEX IF NOT EXISTS "user_status" ON "users" ("status");
CREATE INDEX IF NOT EXISTS "user_is_super_admin" ON "users" ("is_super_admin");

-- Create "user_profiles" table
CREATE TABLE IF NOT EXISTS "user_profiles" (
  "id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "first_name" character varying NULL,
  "last_name" character varying NULL,
  "phone" character varying NULL,
  "birth_date" timestamptz NULL,
  "gender" character varying NULL,
  "city" character varying NULL,
  "address" character varying NULL,
  "photo_url" character varying NULL,
  "bio" text NULL,
  "sociodemographic_data" jsonb NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_profile_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "user_profile_user_id" ON "user_profiles" ("user_id");

-- Create "features" table
CREATE TABLE IF NOT EXISTS "features" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" text NULL,
  "is_core" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "feature_code" ON "features" ("code");

-- Create "permissions" table
CREATE TABLE IF NOT EXISTS "permissions" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "module" character varying NOT NULL DEFAULT 'core',
  "description" text NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "permission_code" ON "permissions" ("code");
CREATE INDEX IF NOT EXISTS "permission_module" ON "permissions" ("module");
