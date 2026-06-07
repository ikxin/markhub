package main

import (
	"log"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"

	"markhub/internal/server"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(nil)
	defer vips.Shutdown()

	router := server.NewRouter(server.Options{})

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if err := router.Run(host + ":" + port); err != nil {
		log.Fatal(err)
	}
}
