package loaders

import (
	"context"
	"errors"
	"forum/db/DAO"
	sm "forum/system_models"
	"log"

	"github.com/vikstrous/dataloadgen"
)

type nearestСhildishCommentsReader struct {
	dao           DAO.Dao
	limit, offset int
	dbErrorsLog   *log.Logger
}

func (nccr *nearestСhildishCommentsReader) getComments(ctx context.Context, commentCtx []sm.CommentCtx) (res [][]*sm.Comment, errs []error) {
	res = make([][]*sm.Comment, len(commentCtx))
	errs = make([]error, len(commentCtx))

	commentsForEveryCtx, err := nccr.dao.ReadNearCommentsByCtx(commentCtx, nccr.limit, nccr.offset)

	if err != nil {
		nccr.dbErrorsLog.Println(err)
		for i := range errs {
			errs[i] = errors.New("ошибка работы загрузчика комментариев ближайшего вложенного уровня")
		}
	}

	for i, cCtx := range commentCtx {
		buf, ok := commentsForEveryCtx[cCtx]
		if !ok {
			buf = []*sm.Comment{}
		}
		res[i] = buf
	}

	return
}

func GetNearestChildishComments(ctx context.Context, commCtx sm.CommentCtx, limit, offset int) ([]*sm.Comment, error) {
	loaderCtx := sm.LoaderCtx{Limit: limit, Offset: offset}
	loaders := getCommentsLoadersFromCtx(ctx)

	exist := loaders.isNearestChildishCommentsLoaderExist(limit, offset)
	if !exist {
		loaders.initNearestChildishCommentsLoader(limit, offset)
	}
	loader, _ := loaders.nearestChildishCommentsLoader.Load(loaderCtx)

	return loader.(*dataloadgen.Loader[sm.CommentCtx, []*sm.Comment]).Load(ctx, commCtx)
}
