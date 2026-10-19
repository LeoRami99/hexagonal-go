package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TenantMembership represents a user's association with an organization (Tenant).
type TenantMembership struct {
	ent.Schema
}

func (TenantMembership) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant_id").NotEmpty(),
		field.String("user_id").NotEmpty(),
		field.Enum("status").Values("ACTIVE", "INACTIVE", "SUSPENDED").Default("ACTIVE"),
		field.Time("joined_at").Default(time.Now).Immutable(),
	}
}

func (TenantMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (TenantMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenants.Type).
			Ref("memberships").
			Field("tenant_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("tenant_memberships").
			Field("user_id").
			Unique().
			Required(),
		edge.To("project_memberships", ProjectMembership.Type),
	}
}

func (TenantMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "user_id").Unique(),
		index.Fields("status"),
	}
}
