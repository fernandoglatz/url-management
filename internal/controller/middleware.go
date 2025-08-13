package controller

import (
	"context"
	"errors"
	"fernandoglatz/url-management/internal/core/common/utils/constants"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/common/utils/log"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const FORMAT_TRACE_STR = "[%.3fms] HTTP %d %s %s %s"

func TraceMiddleware() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ctx := GetContext(ginCtx)
		requestId := uuid.New().String()

		traceMap := make(map[string]any)
		traceMap[constants.REQUEST_ID] = requestId

		ctx = context.WithValue(ctx, constants.TRACE_MAP, traceMap)
		ginCtx.Request = ginCtx.Request.WithContext(ctx)
		ginCtx.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		if log.IsLevelEnabled(log.TRACE) {
			ctx := GetContext(ginCtx)
			begin := time.Now()

			ginCtx.Next()

			elapsed := time.Since(begin)
			duration := float64(elapsed.Nanoseconds()) / 1e6
			reqUri := ginCtx.Request.RequestURI
			reqMethod := ginCtx.Request.Method
			statusCode := ginCtx.Writer.Status()
			clientIP := ginCtx.ClientIP()

			if !strings.Contains(reqUri, "/health") {
				formattedMessage := fmt.Sprintf(FORMAT_TRACE_STR, duration, statusCode, reqMethod, reqUri, clientIP)
				log.Trace(ctx).Msg(formattedMessage)
			}
		}

		ginCtx.Next()
	}
}

func RecoveryMiddleware(ctx context.Context) gin.HandlerFunc {
	errorLogWriter := log.NewLogWritter(*log.Error(ctx))
	return gin.CustomRecoveryWithWriter(errorLogWriter, errorHandleRecovery)
}

func errorHandleRecovery(ginCtx *gin.Context, obj any) {
	ctx := GetContext(ginCtx)

	err, ok := obj.(error)
	if !ok {
		err = errors.New(exceptions.GenericError.Code)
	}

	errw := &exceptions.WrappedError{
		Error: err,
	}

	HandleError(ctx, ginCtx, errw)
}
