package loaders

import (
	"context"
	"forum/db/DAO"
	sm "forum/system_models"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/vikstrous/dataloadgen"
)

type ctxKey string

const (
	loaderKey            = ctxKey("comments_loader")
	defaultLoaderDelayMs = 200
)

type commentsLoaders struct {
	dao           DAO.Dao
	dbErrorsLog   *log.Logger
	loaderDelayMs int

	nearestChildishCommentsLoader sync.Map
}

func (cl *commentsLoaders) init() {
	buf := os.Getenv("LOADERS_DELAY_MS")
	if buf == "" {
		cl.loaderDelayMs = defaultLoaderDelayMs
	} else {
		d, err := strconv.Atoi(buf)
		if err != nil {
			cl.loaderDelayMs = d
		} else {
			cl.loaderDelayMs = defaultLoaderDelayMs
		}
	}
}

func (cl *commentsLoaders) isNearestChildishCommentsLoaderExist(limit, offset int) bool {
	loaderCtx := sm.LoaderCtx{Limit: limit, Offset: offset}
	_, ok := cl.nearestChildishCommentsLoader.Load(loaderCtx)
	return ok
}

func (cl *commentsLoaders) initNearestChildishCommentsLoader(limit, offset int) {
	loaderCtx := sm.LoaderCtx{Limit: limit, Offset: offset}
	pcr := &nearestСhildishCommentsReader{dao: cl.dao, limit: limit, offset: offset, dbErrorsLog: cl.dbErrorsLog}

	cl.nearestChildishCommentsLoader.Store(loaderCtx, dataloadgen.NewLoader(pcr.getComments, dataloadgen.WithWait(time.Millisecond*time.Duration(cl.loaderDelayMs))))
}

func getCommentsLoadersFromCtx(ctx context.Context) *commentsLoaders {
	return ctx.Value(loaderKey).(*commentsLoaders)
}
