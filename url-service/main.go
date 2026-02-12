package main

import(
	"log"
	"url-service/handlers"
	"github.com/gofiber/fiber/v2"
)

func main(){
	app := fiber.New()

	app.Post("/shorten", handlers.ShortenURL)

	log.Println("URL Service running on port 8001")
	log.Fatal(app.Listen(":8001"))
}