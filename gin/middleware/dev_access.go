package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DevAccess(c *gin.Context) {
	ip := net.ParseIP(c.ClientIP()).To4()
	if !ip.IsPrivate() && !ip.IsLoopback() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Next()
}
