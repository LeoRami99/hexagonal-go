package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProjectMembership represents a user's access and participation in a specific Project.
type ProjectMembership struct {
	ent.Schema
}

func (ProjectMembership) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_id").NotEmpty(),
		field.String("tenant_membership_id").NotEmpty(),
		field.Enum("status").Values("ACTIVE", "INACTIVE", "SUSPENDED").Default("ACTIVE"),
		field.Time("joined_at").Default(time.Now).Immutable(),
		field.Time("last_activity_at").Optional().Nillable(),
	}
}

func (ProjectMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (ProjectMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("memberships").
			Field("project_id").
			Unique().
			Required(),
		edge.From("tenant_membership", TenantMembership.Type).
			Ref("project_memberships").
			Field("tenant_membership_id").
			Unique().
			Required(),
		edge.To("roles", ProjectRole.Type),
	}
}

func (ProjectMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "tenant_membership_id").Unique(),
		index.Fields("status"),
	}
}
