package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProjectRole defines a functional role scoped to a specific project.
type ProjectRole struct {
	ent.Schema
}

func (ProjectRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_id").NotEmpty(),
		field.String("code").MaxLen(50).NotEmpty(),
		field.String("name").MaxLen(100).NotEmpty(),
		field.Text("description").Optional().Nillable(),
		field.Bool("is_system").Default(false),
		field.Bool("is_enabled").Default(true),
	}
}

func (ProjectRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

func (ProjectRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("roles").
			Field("project_id").
			Unique().
			Required(),
		edge.To("permissions", Permission.Type),
		edge.From("memberships", ProjectMembership.Type).
			Ref("roles"),
	}
}

func (ProjectRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "code").Unique(),
		index.Fields("is_system"),
		index.Fields("is_enabled"),
	}
}
