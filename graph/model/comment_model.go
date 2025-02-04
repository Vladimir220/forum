package model

type Comment struct {
	ID       string  `json:"id"`
	PostID   string  `json:"postId"`
	ParentId *string `json:"parentId"`
	Content  string  `json:"content"`
	Author   string  `json:"author"`
}

func (c *Comment) Comments(limit *int32, offset *int32) (comments []*Comment, err error) {

	return
}
