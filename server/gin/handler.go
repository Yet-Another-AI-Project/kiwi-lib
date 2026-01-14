package futurxgin

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/utils"
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/limiter"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func NormalHandler[T any](f func(*gin.Context) (T, *facade.FuturxError)) func(*gin.Context) {

	return func(c *gin.Context) {

		defer func() {
			if v := recover(); v != nil {
				err := facade.ErrServerInternal.Wrap(fmt.Errorf("%v\n%s", v, debug.Stack()))
				utils.ResponseError(c, err)
			}
		}()

		data, err := f(c)
		if err != nil {
			utils.ResponseError(c, err)
			return
		}

		c.JSON(http.StatusOK, &facade.BaseResponse{
			Status: facade.StatusSuccess,
			Data:   data,
		})
	}
}

func WebsocketHandler(
	f func(*gin.Context, *websocket.Conn) *facade.FuturxError,
	upgrader websocket.Upgrader,
	limiter limiter.Limiter) func(*gin.Context) {

	return func(c *gin.Context) {

		upgraded := false

		defer func() {
			if v := recover(); v != nil {
				err := facade.ErrServerInternal.Wrap(fmt.Errorf("%v\n%s", v, debug.Stack()))
				if upgraded {
					utils.ResponseWebSocketError(c, err)
				} else {
					utils.ResponseError(c, err)
				}
			}
		}()

		// 1. 检查限流
		allowed, err := limiter.Acquire()
		if err != nil {
			utils.ResponseError(c, facade.ErrServerInternal.Wrap(err))
			return
		}
		if !allowed {
			utils.ResponseError(c, facade.ErrTooManyRequests)
			return
		}

		defer func() {
			_ = limiter.Release()
		}()

		// 2. 升级为 websocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, c.Writer.Header())
		if err != nil {
			utils.ResponseWebSocketError(c, facade.ErrServerInternal.Wrap(fmt.Errorf("upgrade websocket error %w", err)))
			return
		}

		upgraded = true

		defer conn.Close()

		if ferr := f(c, conn); ferr != nil {
			utils.ResponseWebSocketError(c, ferr)
			return
		}
	}
}

func EventStreamHandler(f func(*gin.Context) *facade.FuturxError) func(*gin.Context) {

	return func(c *gin.Context) {

		defer func() {
			if v := recover(); v != nil {
				err := facade.ErrServerInternal.Wrap(fmt.Errorf("%v\n%s", v, debug.Stack()))
				utils.ResponseError(c, err)
			}
		}()

		err := f(c)
		if err != nil {
			utils.ResponseError(c, err)
		}
	}
}
