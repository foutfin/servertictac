package main

import (
	"fmt"
	"os"
	"tacklo/routers"

	"tacklo/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	// "github.com/joho/godotenv"
)

func main() {
	//Loading the Environment files
	// err := godotenv.Load()
	// if err != nil {
	// 	fmt.Println("Error in Loading Environment file")
	// 	return
	// }

	//Room initialised
	rooms := map[string]*types.Room{}

	//App initialing
	app := fiber.New()

	//middleware are here

	//First middleware CORS setup
	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	//WebSocket Upgradation middleware
	app.Use("/ws/:rid", func(c *fiber.Ctx) error {
		room, ok := rooms[c.Params("rid")]
		if ok && (len(room.Players) < 2) {
			if websocket.IsWebSocketUpgrade(c) {
				c.Locals("allowed", true)
				fmt.Println("upgrade called")
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		}
		return c.SendStatus(400)

	})

	//Creating Routers
	routers.CreateRouters(app, rooms)

	app.Listen(":" + os.Getenv("PORT"))

}
