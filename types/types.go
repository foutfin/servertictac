package types

import (
	"fmt"
	"strconv"
	"tacklo/utils"

	"github.com/gofiber/websocket/v2"
)

type Room struct {
	Players    map[int]*Player
	Register   chan *Player
	Unregister chan *Player
	Broadcast  chan int
	State      [3][3]int8
	Start      bool
}

type Player struct {
	Conn   *websocket.Conn
	Send   chan interface{}
	Done   chan int8
	Chance bool
	Score  int8
	Id     int
	Icon   int
}

func (r *Room) Run() {

	for {
		select {

		case player := <-r.Unregister:
			delete(r.Players, player.Id)

		case code := <-r.Broadcast:
			if code == 0 {
				msg := map[string]interface{}{
					"mes": "padded",
				}
				for _, player := range r.Players {
					player.Send <- msg
				}
			} else if code == 1 { // player added
				msg := map[string]interface{}{
					"mes": "pdeleted",
				}
				for _, p := range r.Players {
					p.Send <- msg
				}
			} else if code == 2 { // Game Start
				var oscore int8
				for id, p := range r.Players {
					fmt.Println(id, p)
					if id == 1 {
						oscore = r.Players[2].Score
					} else {
						oscore = r.Players[1].Score
					}

					msg := map[string]interface{}{
						"mes":    "gstarted",
						"you":    p.Score,
						"other":  oscore,
						"chance": p.Chance,
						"id":     id,
						"icon":   p.Icon, // 1 -> circle 2-> cross
						"state":  r.State,
					}
					p.Send <- msg
				}
			} else if code == 3 {
				// fmt.Println("game updater called")
				comp := utils.IsGameCompleted(&r.State)
				if comp == true {
					r.Start = false
					var oscore int8
					for id, p := range r.Players {
						if id == 1 {
							oscore = r.Players[2].Score
						} else {
							oscore = r.Players[1].Score
						}
						if p.Chance == true {
							p.Score += 1
						}
						fmt.Println("game end ", p)
						msg := map[string]interface{}{
							"mes":     "gmend",
							"you":     p.Score,
							"other":   oscore,
							"payload": r.State,
						}
						p.Send <- msg
					}
				} else {
					for _, p := range r.Players {

						p.Chance = !p.Chance
						msg := map[string]interface{}{
							"mes":     "gmud",
							"payload": r.State,
							"chance":  p.Chance,
						}
						p.Send <- msg
					}
				}
			} else if code == 4 {
				msg := map[string]interface{}{
					"mes": "badreq",
				}
				for _, p := range r.Players {
					p.Send <- msg
				}
			}

		}
	}
}

func (player *Player) Writer() {
	// fmt.Println("Player Writer Fired")
	for {
		select {
		case msg := <-player.Send:
			// fmt.Println(msg)
			player.Conn.WriteJSON(msg)
		}
	}
}

func (player *Player) Reader(rooms map[string]*Room, roomId *string) {
	var msg map[string]string
	room := rooms[*roomId]
	for {
		err := player.Conn.ReadJSON(&msg)
		fmt.Println(msg, err)
		if err != nil {
			delete(rooms, *roomId)
			room.Unregister <- player
			room.Broadcast <- 1
			player.Done <- 0
			break
		}

		if msg["mes"] == "start" {
			// O i staken as starting
			if len(room.Players) == 2 {
				room.State = [3][3]int8{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
				chance, _ := strconv.Atoi(msg["payload"])
				if chance == 1 {
					for _, p := range room.Players {
						if p != player {
							p.Send <- map[string]interface{}{"mes": "select"}
							p.Chance = false
							p.Icon = 2
						} else {
							player.Chance = true
							player.Icon = 1
						}
					}
				} else if chance == 2 {
					for _, p := range room.Players {
						if p != player {
							p.Send <- map[string]interface{}{"mes": "select"}
							p.Chance = true
							p.Icon = 1
						} else {
							player.Chance = false
							player.Icon = 2
						}
					}
				}

				room.Start = true
				room.Broadcast <- 2
			} else {
				room.Broadcast <- 4
			}
		} else if msg["mes"] == "gmud" {
			if room.Start == true && player.Chance == true {
				row, _ := strconv.Atoi(msg["row"])
				column, _ := strconv.Atoi(msg["column"])
				if room.State[row][column] == 0 {
					room.State[row][column] = int8(player.Icon)
					room.Broadcast <- 3
					continue
				}
			}
			msg := map[string]interface{}{
				"mes": "badreq",
			}
			player.Send <- msg

		} else if msg["mes"] == "reset" {
			room.State = [3][3]int8{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
			chance, _ := strconv.Atoi(msg["payload"])
			if chance == 1 {
				for _, p := range room.Players {
					p.Score = 0
					if p != player {
						p.Chance = false
						p.Icon = 2
					} else {
						player.Chance = true
						player.Icon = 1
					}
				}
			} else if chance == 2 {
				for _, p := range room.Players {
					p.Score = 0
					if p != player {
						p.Chance = true
						p.Icon = 1
					} else {
						player.Chance = false
						player.Icon = 2
					}
				}
			}
			room.Start = true
			room.Broadcast <- 2
		}
	}
}
