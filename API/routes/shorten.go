package routes

import (
	"net/http"
	"time"

	"github.com/MehulSuthar-000/url-shortener/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
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
