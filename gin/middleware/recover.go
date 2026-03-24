package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	ginApp "github.com/real-web-world/bdk/gin"
)

func RecoveryWithLogFn(logFn func(msg string, keysAndVals ...zap.Field)) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			app := ginApp.GetApp(c)
			reqID := app.GetReqID()
			var actErr error
			logFn("bdk.gin.recover", zap.Any("recoverErr", err), zap.String("reqID", reqID),
				zap.StackSkip("errStack", 2))
			switch err := err.(type) {
			case error:
				actErr = err
			case string:
				errMsg := err
				actErr = errors.New(errMsg)
			default:
				actErr = errors.New("server panic")
			}
			app.ServerError(actErr)
		}()
		c.Next()
	}
}
