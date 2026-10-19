package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Project represents a campus or educational program within an organization (Tenant).
type Project struct {
	ent.Schema
}

func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant_id").NotEmpty(),
		field.String("name").MaxLen(255).NotEmpty(),
		field.String("slug").MaxLen(100).NotEmpty(),
		field.Text("description").Optional().Nillable(),
		field.Enum("status").Values("ACTIVE", "INACTIVE", "ARCHIVED").Default("ACTIVE"),
		field.JSON("domains", []string{}).Optional(),
		field.String("custom_domain").MaxLen(255).Optional().Nillable(),
		field.String("logo_url").MaxLen(512).Optional().Nillable(),
		field.String("timezone").Default("America/Bogota"),
		field.String("locale").Default("es"),
		field.JSON("branding", map[string]any{}).Optional(),
		field.JSON("settings", map[string]any{}).Optional(),
	}
}

func (Project) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenants.Type).
			Ref("projects").
			Field("tenant_id").
			Unique().
			Required(),
		edge.To("memberships", ProjectMembership.Type),
		edge.To("roles", ProjectRole.Type),
		edge.To("features", ProjectFeature.Type),
		edge.To("courses", Courses.Type),
	}
}

func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "slug").Unique(),
		index.Fields("status"),
		index.Fields("custom_domain"),
	}
}
