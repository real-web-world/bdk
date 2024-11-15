package ginApp

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/real-web-world/bdk"
)

func LogFormatter(p gin.LogFormatterParams) string {
	if bdk.IsSkipLogReq(p.Request, p.StatusCode) {
		return ""
	}
	reqTime := p.TimeStamp.Format(time.DateTime)
	path := p.Request.URL.Path
	method := p.Request.Method
	code := p.StatusCode
	clientIp := p.ClientIP
	errMsg := p.ErrorMessage
	processTime := p.Latency
	return fmt.Sprintf("API: %s %d %s %s %s %v %s\n", reqTime, code, clientIp, path, method, processTime,
		errMsg)
}
