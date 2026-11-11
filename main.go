package main

import (
	"net/http"
	"fmt"
	"os"
	"database/sql"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupRouter(conn *sql.DB) *gin.Engine {
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
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("database URl is not exist")
		return
	}
	conn, err := sql.Open("pgx", databaseURL)
	if err != nil {
		fmt.Printf("database initialization error: %v\n", err)
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		fmt.Printf("database unreachable: %v\n", err)
		return
	}
	fmt.Println("Database connection established")
	err = sentry.Init(sentry.ClientOptions{})
	if err != nil {
	 fmt.Printf("sentry init error: %v\n", err)
	 return
	}
	router := setupRouter(conn)
	err = router.Run(":8080")
	if err != nil {
		fmt.Printf("router run error: %v\n", err)
		return
	}
}