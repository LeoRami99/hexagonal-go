package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// UserProfile defines the schema for extended user profile and sociodemographic information.
type UserProfile struct {
	ent.Schema
}

func (UserProfile) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().Unique(),
		field.String("first_name").MaxLen(100).Optional().Nillable(),
		field.String("last_name").MaxLen(100).Optional().Nillable(),
		field.String("phone").MaxLen(30).Optional().Nillable(),
		field.Time("birth_date").Optional().Nillable(),
		field.Enum("gender").Values("MALE", "FEMALE", "NON_BINARY", "PREFER_NOT_TO_SAY").Optional().Nillable(),
		field.String("city").MaxLen(100).Optional().Nillable(),
		field.String("address").MaxLen(255).Optional().Nillable(),
		field.String("photo_url").MaxLen(500).Optional().Nillable(),
		field.Text("bio").Optional().Nillable(),
		field.JSON("sociodemographic_data", map[string]any{}).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (UserProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (UserProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("profile").
			Field("user_id").
			Unique().
			Required(),
	}
}
