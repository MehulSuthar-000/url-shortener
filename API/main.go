package main

import (
	"log"
	"os"

	"github.com/MehulSuthar-000/url-shortener/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRoutes(app *gin.Engine) {
	app.GET("/:url", routes.ResolveURL)
	app.POST("/api/v1", routes.ShortenUrl)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = ":8080" // Default fallback
	}

	router := gin.New()
	// add logger currently gin logger suffice
	router.Use(gin.Logger())

	setupRoutes(router)

	log.Fatal(router.Run(appPort))
}
