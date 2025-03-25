package routes

import (
	"net/http"

	"github.com/MehulSuthar-000/url-shortener/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func ResolveURL(ctx *gin.Context) {
	url := ctx.Param("url")

	r := services.CreateClient(0)
	defer r.Close()

	// Check whether the url exist in the cache or not
	value, err := r.Get(services.Ctx, url).Result()
	if err == redis.Nil {
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"error": "short not found in the database",
			},
		)
		return
	} else if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "cannot connect to DB",
			},
		)
		return
	}

	statsClient := services.CreateClient(1)
	defer statsClient.Close()

	if _, err = statsClient.Incr(services.Ctx, url+"_counter").Result(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update access counter",
		})
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, value)
	return
}
