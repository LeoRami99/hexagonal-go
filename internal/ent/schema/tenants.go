package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Tenants holds the schema definition for the Tenants entity.
type Tenants struct {
	ent.Schema
}

// Fields of the Tenants.
func (Tenants) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(255),
		field.String("slug").MaxLen(100).NotEmpty(),
		field.String("contact_email").MaxLen(255).Optional().Nillable(),
		field.Enum("status").Values("ACTIVE", "INACTIVE", "SUSPENDED", "PENDING").Default("ACTIVE"),
		field.JSON("settings", map[string]any{}).Optional(),
	}
}

// Edges of the Tenants.
func (Tenants) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("projects", Project.Type),
		edge.To("memberships", TenantMembership.Type),
	}
}

// Mixin of the Tenants.
func (Tenants) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Indexes of the Tenants.
func (Tenants) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("status"),
	}
}
