package routers

import (
	"tacklo/handlers"
	"tacklo/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func CreateRouters(app *fiber.App, rooms map[string]*types.Room) {

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"help": "new"})
	})

	//Router For Creating Room --> no payload required
	app.Get("/createroom", func(c *fiber.Ctx) error {
		return handlers.CreateRoom(c, rooms)

	})

	//Router for connecting to per room specific websocket
	app.Get("/ws/:rid", websocket.New(func(c *websocket.Conn) {
		rid := c.Params("rid")
		room := rooms[rid]
		player := types.Player{
			Conn:   c,
			Send:   make(chan interface{}),
			Done:   make(chan int8),
			Chance: false,
			Score:  0,
			Id:     (len(room.Players) + 1),
			Icon:   (len(room.Players) + 1),
		}

		room.Players[player.Id] = &player
		go player.Reader(rooms, &rid)
		go player.Writer()

		if len(room.Players) == 2 {
			room.State = [3][3]int8{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
			chance := 1
			for id, p := range room.Players {
				if id == chance {
					p.Chance = true
				} else {
					p.Chance = false
				}
			}
			room.Start = true
			room.Broadcast <- 2
		}
		<-player.Done
		return

	}))
}
