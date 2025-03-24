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
	app.POST("/api/v1", routes.ShortenURL)
}

func main() {
	err := godotenv.Load("/home/mehul/go_projects/url_shortner/config/.env")
	if err != nil {
		log.Println(err)
	}

	router := gin.New()

	// add logger currently gin logger suffice
	router.Use(gin.Logger())

	setupRoutes(router)

	log.Fatal(router.Run(os.Getenv("APP_PORT")))
}
