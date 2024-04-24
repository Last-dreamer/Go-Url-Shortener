package main

import (
	"fmt"
	"log"
	"os"
	"url-shortner/routes"

	"github.com/gofiber/fiber"
	"github.com/subosito/gotenv"
)

func setupRoutes(app *fiber.App) {

	// app.Get("/:url", routes.ResolveURL)
	// app.Post("/api/v1", routes.ShortenURL)
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)

}

func main() {
	err := gotenv.Load()

	if err != nil {
		fmt.Println(err)

	}
	app := fiber.New()
	// app.Use(logger.New())

	setupRoutes(app)
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
