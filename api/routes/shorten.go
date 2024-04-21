package routes

import (
	"os"
	"strconv"
	"time"
	"url-shortner/database"
	"url-shortner/helpers"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber"
)

type Request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type Response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(Request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	// rate limiter

	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(c.IP()).Result()
	if err == redis.Nil {
		r2.Set(c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err() // for 30 min
	} else {
		value, _ := strconv.Atoi(val)
		if value <= 0 {
			limit, _ := r2.TTL(c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	if !goValidator.IsUrl(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// enforce https etc
	body.URL = helpers.EnforceHTTP(body.URL)

	r2.Decr(c.IP())

	return nil
}
