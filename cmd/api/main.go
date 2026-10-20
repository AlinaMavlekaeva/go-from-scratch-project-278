package main

import (
	"net/http"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return router
}

func main() {
	err := godotenv.Load() 
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf(".env load error: %v\n", err)
			return
		}
	}
	err = sentry.Init(sentry.ClientOptions{})
	if err != nil {
	 fmt.Printf("sentry init error: %v\n", err)
	 return
	}
	router := setupRouter()
	err = router.Run(":8080")
	if err != nil {
		fmt.Printf("router run error: %v\n", err)
		return
	}
}