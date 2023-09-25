package routes

import (
	"url-shortner/database"

	"github.com/gofiber/fiber"
)

// ! just changing
func ResolveURL(c *fiber.Ctx) {

	url := c.Params("url")

	r := database.CreateClient(0)
	defer r.Close()

	r.Get(url)
}
