package utils

import (
	"fmt"
	"strconv"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ResponseWebSocketError(c *gin.Context, err *facade.FuturxError) {
	path := c.Request.URL.Path
	futurxLogger, ok := c.Get("logger")
	if ok {
		fields := []zap.Field{
			zap.String("path", path),
			zap.String("facade_message", err.FacadeMessage),
		}
		if err.InternalError != nil {
			fields = append(fields, zap.String("internal_error", fmt.Sprintf("%+v\n\n", err.InternalError)))
		}

		futurxLogger.(logger.ILogger).Errorf(c, "Gin Request Failed %+v", fields)
	}
}

func ResponseError(c *gin.Context, err *facade.FuturxError) {
	response := &facade.BaseResponse{}
	path := c.Request.URL.Path
	futurxLogger, ok := c.Get("logger")
	if ok {
		fields := []zap.Field{
			zap.String("path", path),
			zap.String("facade_message", err.FacadeMessage),
		}
		if err.InternalError != nil {
			fields = append(fields, zap.String("internal_error", fmt.Sprintf("%+v\n\n", err.InternalError)))
		}

		futurxLogger.(logger.ILogger).Errorf(c, "Gin Request Failed %+v", fields)
	}
	response.Status = "error"
	response.Error = err
	c.AbortWithStatusJSON(err.StatusCode(), response)
}

func GetPageNumAndSize(ctx *gin.Context) (pageNum, pageSize int) {

	pageNum, pageSize = 1, 10

	// 处理 pageNum
	if pageNumStr := ctx.Query("page_num"); pageNumStr != "" {
		if pageNumInt, err := strconv.Atoi(pageNumStr); err == nil && pageNumInt > 1 {
			pageNum = pageNumInt
		}

	}

	// 处理 pageSize
	if pageSizeStr := ctx.Query("page_size"); pageSizeStr != "" {
		if pageSizeInt, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeInt > 0 {
			pageSize = pageSizeInt
		}
	}

	return
}
