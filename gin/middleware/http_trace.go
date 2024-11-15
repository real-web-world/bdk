package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/real-web-world/bdk"
	"github.com/real-web-world/bdk/fastcurd"
	ginApp "github.com/real-web-world/bdk/gin"
	"go.uber.org/zap"
)

const (
	MaxTraceSize   = 1 << 10
	KeyNotSaveResp = "notSaveResp"
	KeyNotTrace    = "notTrace"
)

const (
	ContentTypeJson = "application/json"
)

func NewHttpTrace(logFn func(msg string, keysAndVals ...any)) func(c *gin.Context) {
	return func(c *gin.Context) {
		if bdk.IsSkipLogReq(c.Request, http.StatusOK) {
			c.Next()
			return
		}
		c.Next()
		if isNotTrace(c) {
			return
		}
		app := ginApp.GetApp(c)
		reqID := app.GetReqID()
		reqTime := app.GetProcBeginTime()
		postData := ""
		var resp *fastcurd.RetJSON
		statusCode := app.GetStatusCode()
		contentType := strings.Split(bdk.GetContentType(c.Request), ";")[0]
		switch contentType {
		case ContentTypeJson:
			bodyBts, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBts))
			postData = bdk.Bytes2Str(bodyBts)
		}
		procTime := app.GetProcTime()
		if !isNotSaveResp(c) {
			resp = app.GetCtxRespVal()
		}
		switch statusCode {
		case http.StatusForbidden, http.StatusNotFound:
			return
		default:
		}
		if len(postData) > MaxTraceSize {
			postData = postData[:MaxTraceSize]
		}
		keysAndValues := []any{
			zap.String("bdk.http.url", c.Request.RequestURI),
			zap.String("bdk.http.clientIP", c.ClientIP()),
			zap.Int("bdk.http.statusCode", statusCode),
			zap.Any("bdk.http.resp", resp),
			zap.String("bdk.gin.reqID", reqID),
			zap.Duration("bdk.gin.procTime", procTime),
			zap.Timep("bdk.gin.reqTime", reqTime),
			zap.String("bdk.gin.postData", postData),
		}
		logFn("httpTrace", keysAndValues...)
	}
}

func isNotSaveResp(c *gin.Context) bool {
	b, ok := c.Get(KeyNotSaveResp)
	if !ok {
		return false
	}
	if t, ok := b.(bool); ok {
		return t
	}
	return false
}
func setNotSaveResp(c *gin.Context) {
	c.Set(KeyNotSaveResp, true)
}
func isNotTrace(c *gin.Context) bool {
	b, ok := c.Get(KeyNotTrace)
	if !ok {
		return false
	}
	if t, ok := b.(bool); ok {
		return t
	}
	return false
}
func setNotTrace(c *gin.Context) {
	c.Set(KeyNotTrace, true)
}

func NotSaveResp(c *gin.Context) {
	setNotSaveResp(c)
	c.Next()
}
func NotTrace(c *gin.Context) {
	setNotTrace(c)
	c.Next()
}
