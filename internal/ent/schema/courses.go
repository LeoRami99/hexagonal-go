package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Courses holds the schema definition for the Courses entity.
type Courses struct {
	ent.Schema
}

// Fields of the Courses.
func (Courses) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_id").MaxLen(100).NotEmpty(),
		field.String("title").MaxLen(350).NotEmpty(),
		field.String("slug").MaxLen(150).NotEmpty(),
		field.Text("summary").NotEmpty(),
		field.Text("description").NotEmpty(),
		field.String("level").MaxLen(50).NotEmpty(),
		field.String("thumbnail_url").MaxLen(350).NotEmpty(),
		field.String("status").MaxLen(50).NotEmpty(),
		
	}
}

// Edges of the Courses.
func (Courses) Edges() []ent.Edge {
	return nil
}


// Mixin of the Courses.
func (Courses) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TenantMixin{},
		TimeMixin{},
	}
}
