package system_models

type CommentCtx struct {
	PostId   string
	ParentId string
}

type LoaderCtx struct {
	Limit, Offset int
}
