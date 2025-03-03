package daemon

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/autoai-org/aid/internal/utilities"
	"github.com/gin-gonic/gin"
)

// beforeResponse set global header to enable cors and set response header
func beforeResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		c.Writer.Header().Set("aid-version", "1.0.1 @ dev")
		if c.Writer.Header().Get("Access-Control-Allow-Origin") == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
		if c.Request.Method == "OPTIONS" {
			c.Writer.WriteHeader(http.StatusOK)
		}
	}
}

// RunServer starts the http(s) service
func RunServer(port string) {
	if port == "" {
		port = "10590"
		utilities.Formatter.Warn("Port not specified, using the default " + port)
	}
	utilities.Formatter.Info("Starting the server...")
	r := getRouter()
	err := r.Run("127.0.0.1:" + port)
	utilities.ReportError(err, "Cannot start server")
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utilities.Formatter.Info("Server is shutting down gracefully...")
}
