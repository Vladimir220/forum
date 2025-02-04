package system_models

type UserDb struct {
	ID             string
	Username       string
	HashedPassword string
}

type Comment struct {
	ID       string
	PostID   string
	ParentId *string
	Content  string
	Author   string
}

type User struct {
	ID       string
	Username string
}

type Post struct {
	ID            string
	Title         string
	Content       string
	AllowComments bool
	Author        string
}
