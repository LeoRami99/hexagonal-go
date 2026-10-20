package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProjectFeature represents the activation and configuration of a feature in a specific project.
type ProjectFeature struct {
	ent.Schema
}

func (ProjectFeature) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_id").NotEmpty(),
		field.String("feature_id").NotEmpty(),
		field.Bool("is_enabled").Default(true),
		field.JSON("settings", map[string]any{}).Optional(),
		field.Time("enabled_at").Default(time.Now).Immutable(),
	}
}

func (ProjectFeature) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (ProjectFeature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("features").
			Field("project_id").
			Unique().
			Required(),
		edge.From("feature", Feature.Type).
			Ref("project_features").
			Field("feature_id").
			Unique().
			Required(),
	}
}

func (ProjectFeature) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "feature_id").Unique(),
		index.Fields("is_enabled"),
	}
}
