package routes

import (
	"log"
	"os"
	"strconv"
	"time"
	"url-shortner/database"
	"url-shortner/helpers"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber"
	"github.com/google/uuid"
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

func ShortenURL(c *fiber.Ctx) {

	log.Println("testing if working")
	body := new(Request)

	if err := c.BodyParser(&body); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
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
			c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	if !govalidator.IsURL(body.URL) {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// enforce https etc
	body.URL = helpers.EnforceHTTP(body.URL)

	//
	var id string

	if body.CustomShort == "" {
		id = uuid.NewString()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val2, _ := r.Get(id).Result()
	if val2 != "" {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "short url is already in use"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to server"})
	}

	resp := Response{
		URL:            body.URL,
		CustomShort:    "",
		Expiry:         body.Expiry,
		XRateRemaining: 10,
		XRateLimitRest: 30,
	}

	r2.Decr(c.IP())

	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(c.IP()).Result()
	resp.XRateLimitRest = ttl / time.Nanosecond / time.Minute
	resp.CustomShort = os.Getenv("BASE_URL") + "/" + id

	c.Status(fiber.StatusOK).JSON(resp)

}
