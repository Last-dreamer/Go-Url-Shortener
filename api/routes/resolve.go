package routes

import (
	"url-shortner/database"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber"
)

// ! just changing
func ResolveURL(c *fiber.Ctx) error {

	url := c.Params("url")

	r := database.CreateClient(0)
	defer r.Close()

	value, err := r.Get(url).Result()

	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "short url is not found"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot connect to db",
		})
	}

	incr := database.CreateClient(1)
	defer incr.Close()

	incr.Incr("counter")
	//! i have to see this if working .....
	c.Redirect(value, 301)
	return nil
}
