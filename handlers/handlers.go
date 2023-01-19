package handlers

import (
	"tacklo/types"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateRoom(c *fiber.Ctx, rooms map[string]*types.Room) error {
	id := uuid.New().String()
	r := types.Room{Register: make(chan *types.Player),
		Players:    make(map[int]*types.Player),
		Unregister: make(chan *types.Player),
		Broadcast:  make(chan int),
		Start:      false,
	}
	rooms[id] = &r
	go r.Run()
	return c.JSON(fiber.Map{"status": 200, "roomId": id})

}
