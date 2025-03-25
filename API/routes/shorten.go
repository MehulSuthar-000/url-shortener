package routes

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MehulSuthar-000/url-shortener/helpers"
	"github.com/MehulSuthar-000/url-shortener/services"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type request struct {
	URL         string        `json:"url" binding:"required"`
	CustomShort string        `json:"short" binding:"required"`
	Expiry      time.Duration `json:"expiry" binding:"required"`
}

type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(ctx *gin.Context) {
	body := new(request)

	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "cannot parse json",
			},
		)
		return
	}

	// Implement rate limiting
	rateDB := services.CreateClient(1)
	defer rateDB.Close()

	clientIP := ctx.ClientIP()
	apiQuota, err := strconv.Atoi(os.Getenv("API_QUOTA"))
	if err != nil || apiQuota <= 0 {
		apiQuota = 10 // Default quota
	}
	// Decrease quota atomically
	val, err := rateDB.Decr(services.Ctx, clientIP).Result()
	if err == redis.Nil {
		// Initialize for a new user
		_ = rateDB.Set(services.Ctx, clientIP, apiQuota-1, 30*time.Minute).Err()
		return
	} else if err != nil {
		// Log Redis connection error
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error"})
		return
	}

	if val < 0 {
		// Get time left for the key to expire
		ttl, err := rateDB.TTL(services.Ctx, clientIP).Result()
		if err != nil || ttl < 0 {
			ttl = 0
		}
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"error":            "Rate limit exceeded",
			"rate_limit_reset": int(ttl.Seconds() / 60), // Convert TTL to minutes
		})
		return
	}

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
}
