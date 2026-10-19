package domain

// field.String("project_id").MaxLen(100).NotEmpty(),
// 		field.String("title").MaxLen(350).NotEmpty(),
// 		field.String("slug").MaxLen(150).NotEmpty(),
// 		field.Text("summary").NotEmpty(),
// 		field.Text("description").NotEmpty(),
// 		field.String("level").MaxLen(50).NotEmpty(),
// 		field.String("thumbnail_url").MaxLen(350).NotEmpty(),
// 		field.String("status").MaxLen(50).NotEmpty(),

type Course struct {
	ID 	string
	ProjectID	string
	Title 	string
	Slug 	string
	Summary 	string
	Description string
	Level 	string
	ThumbnailURL string
	Status 	string
	CreatedAt 	string
	UpdatedAt 	string
}