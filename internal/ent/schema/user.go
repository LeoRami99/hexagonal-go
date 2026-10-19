package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User defines the database schema for application users.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("auth_id").NotEmpty().Unique(),
		field.String("name").MaxLen(100).NotEmpty(),
		field.String("email").MaxLen(255).NotEmpty(),
		field.String("avatar_url").MaxLen(512).Optional().Nillable(),
		field.Bool("is_super_admin").Default(false),
		field.Enum("status").Values("ACTIVE", "INACTIVE", "SUSPENDED", "BLOCKED").Default("ACTIVE"),
		field.Bool("active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("profile", UserProfile.Type).Unique(),
		edge.To("tenant_memberships", TenantMembership.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
		index.Fields("auth_id").Unique(),
		index.Fields("created_at"),
		index.Fields("status"),
		index.Fields("is_super_admin"),
	}
}
