package conversion

import (
	"forum/graph/model"
	sm "forum/system_models"
)

func ConvPost(p sm.Post) (res model.Post) {
	res.AllowComments = p.AllowComments
	res.Author = p.Author
	res.Content = p.Content
	res.ID = p.ID
	res.Title = p.Title
	return
}

func ConvComment(c sm.Comment) (res model.Comment) {
	res.Author = c.Author
	res.Content = c.Content
	res.ID = c.ID
	res.ParentId = c.ParentId
	res.PostID = c.PostID
	return
}

func ConvArrPtrPost(arrP []*sm.Post) (arrRes []*model.Post) {
	arrRes = make([]*model.Post, len(arrP))
	for i, v := range arrP {
		res := ConvPost(*v)
		arrRes[i] = &res
	}

	return
}

func ConvArrPtrComment(arrP []*sm.Comment) (arrRes []*model.Comment) {
	arrRes = make([]*model.Comment, len(arrP))
	for i, v := range arrP {
		res := ConvComment(*v)
		arrRes[i] = &res
	}

	return
}
