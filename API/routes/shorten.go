package routes

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MehulSuthar-000/url-shortener/helpers"
	"github.com/MehulSuthar-000/url-shortener/services"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type request struct {
	URL         string        `json:"url" binding:"required"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(ctx *gin.Context) {
	body := new(request)

	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   "cannot parse json",
				"details": err.Error(),
			},
		)
		return
	}

	// Debug environment variables
	log.Println("Environment Variables:")
	log.Printf("API_QUOTA: %s\n", os.Getenv("API_QUOTA"))
	log.Printf("REDIS_ADDR: %s\n", os.Getenv("REDIS_ADDR"))
	log.Printf("DOMAIN: %s\n", os.Getenv("DOMAIN"))

	// implement rate limiting
	// everytime a user queries, check if the IP is already in services,
	// if yes, decrement the calls remaining by one, else add the IP to services
	// with expiry of `30mins`. So in this case the user will be able to send 10
	// requests every 30 minutes
	// r2 := services.CreateClient(1)
	// defer r2.Close()
	// val, err := r2.Get(services.Ctx, ctx.ClientIP()).Result()
	// if err == redis.Nil {
	// 	_ = r2.Set(services.Ctx, ctx.ClientIP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err() //change the rate_limit_reset here, change `30` to your number
	// } else {
	// 	val, _ = r2.Get(services.Ctx, ctx.ClientIP()).Result()
	// 	valInt, _ := strconv.Atoi(val)
	// 	if valInt <= 0 {
	// 		limit, _ := r2.TTL(services.Ctx, ctx.ClientIP()).Result()
	// 		ctx.JSON(http.StatusServiceUnavailable,
	// 			gin.H{
	// 				"error":            "Rate limit exceeded",
	// 				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
	// 			})
	// 		return
	// 	}
	// }

	// Check if the input is an actual URL
	if !govalidator.IsURL(body.URL) {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "Invalid URL",
			},
		)
		return
	}

	// Check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		ctx.JSON(
			http.StatusServiceUnavailable,
			gin.H{
				"error": "you can't hack the system (:",
			},
		)
		return
	}

	// Enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := services.CreateClient(0)
	defer r.Close()

	value, err := r.Get(services.Ctx, id).Result()
	if err != nil && err != redis.Nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Redis error occurred",
		})
		return
	}

	if value != "" {
		ctx.JSON(
			http.StatusForbidden,
			gin.H{
				"error": "URL custom short is already in use",
			},
		)
		return
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(services.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "Unable to connect to server",
			},
		)
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// r2.Decr(services.Ctx, ctx.ClientIP())
	// val, _ = r2.Get(services.Ctx, ctx.ClientIP()).Result()
	// resp.XRateRemaining, _ = strconv.Atoi(val)
	// ttl, _ := r2.TTL(services.Ctx, ctx.ClientIP()).Result()
	// resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	ctx.JSON(
		http.StatusOK,
		resp,
	)
}
