package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
	"regexp"

	lua "github.com/yuin/gopher-lua"
)

var roomPathRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(\/[a-zA-Z0-9]+)*$`)
var L *lua.LState

/* ================= DATA ================= */

type Client struct {
	Name        string
	Conn        net.Conn
	Room        string
	ConnectedAt time.Time
	ChatCount   int
	NickHistory []string
}

type Server struct {
	Rooms map[string]map[*Client]bool
	mu    sync.Mutex
}

/* ================= MAIN ================= */

func main() {
	rand.Seed(time.Now().UnixNano())

	L = lua.NewState()
	defer L.Close()

	if err := L.DoFile("commands.lua"); err != nil {
		log.Fatal(err)
	}

	srv := &Server{
		Rooms: make(map[string]map[*Client]bool),
	}

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("anon-chat :: phase 2 (revised) :: :9000")

	for {
		conn, _ := ln.Accept()
		go srv.handleClient(conn)
	}
}

/* ================= CLIENT ================= */

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	c := &Client{
		Name:        fmt.Sprintf("anon-%04x", rand.Intn(65536)),
		Conn:        conn,
		ConnectedAt: time.Now(),
	}

	send(c, "\033[36mWelcome to anon-chat (terminal)\033[0m")
	s.sendRooms(c)
	send(c, "Type /help")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "/") {
			s.handleCommand(c, line)
			continue
		}

		if c.Room == "" {
			send(c, "join room first: /join parent/child")
			continue
		}

		c.ChatCount++
		s.broadcastTree(c, line)
	}

	s.leave(c)
}

/* ================= COMMAND ================= */

func (s *Server) handleCommand(c *Client, input string) {
	fn := L.GetGlobal("handle_command")
	L.Push(fn)
	L.Push(lua.LString(input))

	if err := L.PCall(1, 2, nil); err != nil {
		send(c, "lua error")
		return
	}

	action := L.Get(-2).String()
	arg := L.Get(-1).String()
	L.Pop(2)

	switch action {

	case "join":
		if arg == "" {
			send(c, "[system] room name cannot be empty")
			return
		}
		if !isValidRoomPath(arg) {
			send(c, "[system] invalid room name (use a-z, 0-9, /)")
			return
		}
		
		s.joinRoom(c, arg)

	case "nick":
		if arg == "" {
			send(c, "usage: /nick <name>")
			return
		}
		old := c.Name
		c.Name = arg
		c.NickHistory = append(c.NickHistory, old+" -> "+arg)
		s.broadcastRoomSystem(c.Room, old+" is now "+arg)

	case "rooms":
		s.sendRooms(c)

	case "info":
		s.sendInfo(c, arg)

	case "help":
		send(c, "/join /nick /rooms /info [nick] /quit")

	case "who":
		s.showWho(c)
		
	case "quit":
		send(c, "bye")
		c.Conn.Close()
	}
}

/* ================= ROOM ================= */

func (s *Server) joinRoom(c *Client, room string) {
	parts := strings.Split(room, "/")
	if len(parts) > 4 {
		send(c, "max room depth is 4")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if c.Room != "" {
		old := c.Room
		delete(s.Rooms[old], c)
	
		if len(s.Rooms[old]) == 0 {
			delete(s.Rooms, old)
		}
	}

	if s.Rooms[room] == nil {
		s.Rooms[room] = make(map[*Client]bool)
	}

	s.Rooms[room][c] = true
	c.Room = room

	send(c, "\033[32mjoined "+room+"\033[0m")
	s.broadcastRoomSystem(room, c.Name+" joined the room")
}

/* ================= BROADCAST ================= */

func (s *Server) broadcastTree(sender *Client, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	senderRoom := sender.Room
	senderDepth := depth(senderRoom)
	root := root(senderRoom)

	for room, users := range s.Rooms {
		// hanya tree yang sama
		if !strings.HasPrefix(room, root) {
			continue
		}

		// hanya ke atas / sama
		if depth(room) > senderDepth {
			continue
		}

		for c := range users {
			if c == sender {
				continue
			}

			// room sama
			mentions, mentionAll := extractMentions(msg)
			
			if c.Room == senderRoom {
			
				// jangan kirim ke diri sendiri
				if c == sender {
					continue
				}
			
				mentioned := mentionAll
			
				if !mentioned {
					for _, m := range mentions {
						if m == c.Name {
							mentioned = true
							break
						}
					}
				}
			
				if mentioned {
					sendMention(c, sender, msg)
				} else {
					send(c, sender.Name+": "+msg)
				}
			
				continue
			}
		}
	}
}

func (s *Server) broadcastRoomSystem(room, msg string) {
	for c := range s.Rooms[room] {
		send(c, "\033[35m[system]\033[0m "+msg)
	}
}

/* ================= INFO ================= */

func (s *Server) sendRooms(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	send(c, "Active rooms:")
	if len(s.Rooms) == 0 {
		send(c, "none")
	}

	for r, users := range s.Rooms {
		send(c, fmt.Sprintf("%s [%d]", r, len(users)))
	}
}

func (s *Server) sendInfo(req *Client, targetName string) {
	var t *Client

	s.mu.Lock()
	defer s.mu.Unlock()

	if targetName == "" {
		t = req
	} else {
		for _, users := range s.Rooms {
			for c := range users {
				if c.Name == targetName {
					t = c
				}
			}
		}
	}

	if t == nil {
		send(req, "user not found")
		return
	}

	send(req, "Name   : "+t.Name)
	send(req, fmt.Sprintf("Msgs   : %d", t.ChatCount))
	send(req, "Room   : "+t.Room)
	send(req, "Joined : "+t.ConnectedAt.Format(time.RFC3339))

	if len(t.NickHistory) > 0 {
		send(req, "Nick log:")
		start := len(t.NickHistory) - 3
		if start < 0 {
			start = 0
		}
		for _, h := range t.NickHistory[start:] {
			send(req, h)
		}
	}
	send(req, "---------------------")
}

/* ================= LEAVE ================= */

func (s *Server) leave(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c.Room == "" {
		return
	}

	room := c.Room
	delete(s.Rooms[room], c)

	if len(s.Rooms[room]) == 0 {
		delete(s.Rooms, room)
	}

	c.Room = ""
}

/* ================= UTIL ================= */

func depth(room string) int {
	return len(strings.Split(room, "/"))
}

func root(room string) string {
	return strings.Split(room, "/")[0]
}

func send(c *Client, msg string) {
	c.Conn.Write([]byte(msg + "\r\n"))
}

func relativePath(receiver, sender string) string {
	r := strings.Split(receiver, "/")
	s := strings.Split(sender, "/")

	if len(s) <= len(r) {
		return ""
	}

	return strings.Join(s[len(r):], "/")
}

func isValidRoomPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	return roomPathRegex.MatchString(path)
}

func (s *Server) showWho(c *Client) {
	if c.Room == "" {
		send(c, "[system] you are not in a room")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	users, ok := s.Rooms[c.Room]
	if !ok {
		send(c, "[system] room not found")
		return
	}

	send(c, fmt.Sprintf(
		"[system] users in room [%s] (%d):",
		c.Room,
		len(users),
	))

	for u := range users {
		send(c, " - "+u.Name)
	}
}

func extractMentions(msg string) (mentions []string, mentionAll bool) {
	words := strings.Fields(msg)

	for _, w := range words {
		if strings.HasPrefix(w, "@") && len(w) > 1 {
			name := strings.TrimPrefix(w, "@")
			if name == "all" {
				mentionAll = true
			} else {
				mentions = append(mentions, name)
			}
		}
	}
	return
}

func sendMention(target *Client, sender *Client, msg string) {
	send(target, "-------------")
	send(target, "@"+sender.Name+" mention you : "+msg)
	send(target, "-------------")
}
