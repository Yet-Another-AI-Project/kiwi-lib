package utils

import (
	"strconv"

	"github.com/Yet-Another-AI-Project/kiwi-lib/logger"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

func ResponseWebSocketError(c *gin.Context, err *facade.Error) {
	path := c.Request.URL.Path
	futurxLogger, ok := c.Get("logger")
	if ok {
		futurxLogger.(logger.ILogger).Errorf(c, "Gin Request Failed, path: %s, facade_message: %s, internal_error: %v", path, err.FacadeMessage, err.InternalError)
	}
}

func ResponseError(c *gin.Context, err *facade.Error) {
	response := &facade.BaseResponse{}
	path := c.Request.URL.Path
	futurxLogger, ok := c.Get("logger")
	if ok {
		futurxLogger.(logger.ILogger).Errorf(c, "Gin Request Failed, path: %s, facade_message: %s, internal_error: %v", path, err.FacadeMessage, err.InternalError)
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
