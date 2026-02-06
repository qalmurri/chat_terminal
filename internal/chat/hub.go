package chat

import (
	"database/sql"
	"mager/internal/database"
	"golang.org/x/crypto/ssh"
	"sync"
)

type Client struct {
	Nick    string
	Channel ssh.Channel
}

type Hub struct {
	Rooms      map[string]map[string]*Client
	Broadcast  chan Message
	Register   chan Registration
	Unregister chan Registration
	mu         sync.Mutex
	DB         *sql.DB
}

type Message struct {
	Room    string
	Content string
}

type Registration struct {
	Room   string
	Client *Client
}

func NewHub(db *sql.DB) *Hub {
	return &Hub{
		Rooms:      make(map[string]map[string]*Client),
		Broadcast:  make(chan Message),
		Register:   make(chan Registration),
		Unregister: make(chan Registration),
		DB:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.Register:
			h.mu.Lock()
			if h.Rooms[reg.Room] == nil {
				h.Rooms[reg.Room] = make(map[string]*Client)
			}
			h.Rooms[reg.Room][reg.Client.Nick] = reg.Client
			h.mu.Unlock()

		case unreg := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Rooms[unreg.Room]; ok {
				delete(clients, unreg.Client.Nick)
				
				if len(clients) == 0 {
					delete(h.Rooms, unreg.Room)
					// Jalankan penghapusan DB di background
					go func(name string) {
						database.DeleteRoom(h.DB, roomName)
					}(unreg.Room)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.Lock()
			if clients, ok := h.Rooms[msg.Room]; ok {
				for _, client := range clients {
					client.Channel.Write([]byte(msg.Content + "\r\n"))
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) GetRoomCount(roomName string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.Rooms[roomName])
}

func (h *Hub) UpdateClientNick(oldFullNick string, newFullNick string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Kita harus mencari client di semua room karena kita tidak tahu dia ada di room mana
	for roomName, clients := range h.Rooms {
		if client, ok := clients[oldFullNick]; ok {
			// Update nickname di objek client
			client.Nick = newFullNick
			// Hapus key lama di map room tersebut
			delete(clients, oldFullNick)
			// Simpan dengan key baru
			h.Rooms[roomName][newFullNick] = client
			return // Keluar setelah ketemu
		}
	}
}

func (h *Hub) GetUsersInRoom(roomName string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var nicks []string
	if clients, ok := h.Rooms[roomName]; ok {
		for nick := range clients {
			nicks = append(nicks, nick)
		}
	}
	return nicks
}
