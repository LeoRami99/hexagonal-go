package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Permission defines the atomic actions allowed in the system.
type Permission struct {
	ent.Schema
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").MaxLen(100).NotEmpty().Unique(),
		field.String("module").MaxLen(50).NotEmpty().Default("core"),
		field.String("description").MaxLen(255).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", ProjectRole.Type).Ref("permissions"),
	}
}

func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("module"),
	}
}
