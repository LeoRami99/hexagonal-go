package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Courses holds the schema definition for the Courses entity.
type Courses struct {
	ent.Schema
}

// Fields of the Courses.
func (Courses) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_id").MinLen(26).MaxLen(26).NotEmpty().Immutable(),
		field.String("title").MaxLen(350).NotEmpty(),
		field.String("slug").MaxLen(150).NotEmpty(),
		field.Text("summary").NotEmpty(),
		field.Text("description").NotEmpty(),
		field.Enum("level").Values("beginner", "intermediate", "advanced"),
		field.String("thumbnail_url").MaxLen(350).NotEmpty(),
		field.Enum("status").Values("draft", "published", "archived").Default("draft"),
	}
}

// Edges of the Courses.
func (Courses) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("courses").
			Field("project_id").
			Unique().
			Immutable().
			Required(),
	}
}

func (Courses) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "slug").Unique(),
		index.Fields("tenant_id"),
		index.Fields("status"),
		index.Fields("level"),
	}
}

// Mixin of the Courses.
func (Courses) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TenantMixin{},
		TimeMixin{},
	}
}
