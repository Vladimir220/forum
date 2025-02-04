package model

import (
	"context"
	"errors"
	"forum/loaders"
	sm "forum/system_models"
)

type Post struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	AllowComments bool   `json:"allowComments"`
	Author        string `json:"author"`
}

func (p *Post) Comments(ctx context.Context, limit *int32, offset *int32) (comments []*Comment, err error) {
	commentsCtx := sm.CommentCtx{PostId: p.ID}

	res, dbErr := loaders.GetNearestChildishComments(ctx, commentsCtx, int(*limit), int(*offset))
	if dbErr != nil {
		err = errors.New("база данных не смогла обработать запрос")
	}

	comments = make([]*Comment, len(res))
	for i, v := range res {
		res := Comment{}
		res.Author = v.Author
		res.Content = v.Content
		res.ID = v.ID
		res.ParentId = v.ParentId
		res.PostID = v.PostID
		comments[i] = &res
	}

	return

}
