package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Feature defines the global catalog of platform features (courses, quizzes, certificates, etc.).
type Feature struct {
	ent.Schema
}

func (Feature) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").MaxLen(50).NotEmpty().Unique(),
		field.String("name").MaxLen(100).NotEmpty(),
		field.Text("description").Optional().Nillable(),
		field.Bool("is_core").Default(false),
	}
}

func (Feature) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

func (Feature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("project_features", ProjectFeature.Type),
	}
}

func (Feature) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}
