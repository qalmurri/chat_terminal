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
)

const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorGray    = "\033[90m"
)

func colorize(color, text string) string {
	return color + text + ColorReset
}
const (
	MaxMessages   = 5
	WindowSeconds = 5
	MuteSeconds   = 10
)

type Client struct {
	Name        string
	Conn        net.Conn
	Room        *Room

	// anti flood
	MsgCount    int
	WindowStart time.Time
	MutedUntil  time.Time

	// info tracking
	ConnectedAt time.Time
	ChatCount   int
	NickHistory []string
}

type Room struct {
	Name    string
	Clients map[*Client]bool
}

type Server struct {
	Rooms map[string]*Room
	mu    sync.Mutex
}

func main() {
	rand.Seed(time.Now().UnixNano())

	server := &Server{
		Rooms: make(map[string]*Room),
	}

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Anon terminal chat running on :9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go server.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	client := &Client{
		Name:        fmt.Sprintf("anon-%04x", rand.Intn(65536)),
		Conn:        conn,
		WindowStart: time.Now(),
	}

	sendWelcome(client)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// mute check
		if time.Now().Before(client.MutedUntil) {
			remain := int(time.Until(client.MutedUntil).Seconds())
			client.send(fmt.Sprintf(
				"[system] muted for spam (%ds remaining)", remain,
			))
			continue
		}

		// flood check
		if client.isFlooding() {
			client.MutedUntil = time.Now().Add(MuteSeconds * time.Second)
			client.send(fmt.Sprintf(
				"[system] you are muted for %d seconds (spam detected)", MuteSeconds,
			))
			continue
		}

		if strings.HasPrefix(input, "/") {
			s.handleCommand(client, input)
			continue
		}

		if client.Room == nil {
			client.send("[system] you are not in a room")
			client.send("Type: /join <room>")
			continue
		}

		client.ChatCount++
		client.Room.broadcast(client, input)
	}

	// disconnect
	s.leaveRoom(client, true)
}

func (c *Client) isFlooding() bool {
	now := time.Now()

	if now.Sub(c.WindowStart) > WindowSeconds*time.Second {
		c.WindowStart = now
		c.MsgCount = 0
	}

	c.MsgCount++
	return c.MsgCount > MaxMessages
}

func sendWelcome(c *Client) {
	c.send("")
	c.send("========================================")
	c.send("      anon-chat :: terminal only")
	c.send("========================================")
	c.send("")
	c.send("Connected as: " + c.Name)
	c.send("No login • No history • Plain text")
	c.send("")
	c.send("Commands:")
	c.send("  /join <room>     join or create a room")
	c.send("  /rooms           list active rooms")
	c.send("  /nick <name>     change nickname")
	c.send("  /who             list users in room")
	c.send("  /me <action>     emote action")
	c.send("  /help            show this help")
	c.send("  /quit            disconnect")
	c.send("")
	c.send("Start by typing:")
	c.send("  /join lobby")
	c.send("")
	c.send("----------------------------------------")
}

func (s *Server) handleCommand(c *Client, input string) {
	parts := strings.SplitN(input, " ", 2)
	cmd := parts[0]

	switch cmd {

	case "/join":
		if len(parts) < 2 {
			c.send("usage: /join <room>")
			return
		}
		s.joinRoom(c, parts[1])

	case "/nick":
		if len(parts) < 2 {
			c.send("usage: /nick <name>")
			return
		}
		old := c.Name
		newNick := parts[1]
	
		c.NickHistory = append(c.NickHistory, old+" → "+newNick)
		c.Name = newNick
	
		if c.Room != nil {
			c.Room.broadcastSystem(old + " is now " + newNick)
		}

	case "/rooms":
		s.mu.Lock()
		if len(s.Rooms) == 0 {
			c.send("no active rooms")
		}
		for name := range s.Rooms {
			c.send("- " + name)
		}
		s.mu.Unlock()

	case "/who":
		if c.Room == nil {
			c.send("you are not in a room")
			return
		}
		c.send("users in room:")
		for u := range c.Room.Clients {
			c.send(" - " + u.Name)
		}

	case "/me":
		if len(parts) < 2 {
			c.send("usage: /me <action>")
			return
		}
		if c.Room != nil {
			c.Room.broadcastAction(c, parts[1])
		}

	case "/help":
		sendWelcome(c)

	case "/info":
		target := c
	
		if len(parts) == 2 && c.Room != nil {
			for u := range c.Room.Clients {
				if strings.EqualFold(u.Name, parts[1]) {
					target = u
					break
				}
			}
		}
	
		showUserInfo(c, target)
	
		case "/quit":
			c.send("bye")
			s.leaveRoom(c, true)
			c.Conn.Close()
	
		default:
			c.send("unknown command, type /help")
		}
	}

func (s *Server) joinRoom(c *Client, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c.Room != nil {
		c.Room.broadcastSystem(c.Name + " left the room")
		delete(c.Room.Clients, c)
	}

	room, exists := s.Rooms[name]
	if !exists {
		room = &Room{
			Name:    name,
			Clients: make(map[*Client]bool),
		}
		s.Rooms[name] = room
	}

	c.Room = room
	room.Clients[c] = true

	c.send("joined room " + name)
	room.broadcastSystem(c.Name + " joined the room")
}

func (s *Server) leaveRoom(c *Client, notify bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c.Room != nil {
		if notify {
			c.Room.broadcastSystem(c.Name + " left the room")
		}
		delete(c.Room.Clients, c)
		c.Room = nil
	}
}

func (r *Room) broadcast(sender *Client, message string) {
	mentions := extractMentions(message)

	for c := range r.Clients {
		if c == sender {
			continue
		}

		mentioned := false
		for _, m := range mentions {
			if strings.EqualFold(m, c.Name) {
				mentioned = true
				break
			}
		}

		if mentioned {
			box := asciiBox(
				colorize(ColorMagenta, "MENTION FROM "+sender.Name),
				"",
				colorize(ColorMagenta, message),
			)
			for _, line := range box {
				c.send(line)
			}
		} else {
			c.send(fmt.Sprintf(
				"%s%s%s %s",
				ColorCyan,
				"["+sender.Name+"]",
				ColorReset,
				message,
			))
		}
	}
}

func (r *Room) broadcastSystem(message string) {
	for c := range r.Clients {
		c.send(colorize(ColorYellow, "[system] "+message))
	}
}

func (r *Room) broadcastAction(sender *Client, action string) {
	for c := range r.Clients {
		if c != sender {
			c.send(colorize(
				ColorGreen,
				"* "+sender.Name+" "+action,
			))
		}
	}
}

func (c *Client) send(msg string) {
	fmt.Fprintln(c.Conn, msg+ColorReset)
}

func extractMentions(msg string) []string {
	words := strings.Fields(msg)
	var mentions []string

	for _, w := range words {
		if strings.HasPrefix(w, "@") && len(w) > 1 {
			name := strings.TrimPrefix(w, "@")
			mentions = append(mentions, name)
		}
	}
	return mentions
}

func asciiBox(lines ...string) []string {
	width := 0
	for _, l := range lines {
		if len(l) > width {
			width = len(l)
		}
	}

	border := "+" + strings.Repeat("-", width+2) + "+"
	var box []string

	box = append(box, border)
	for _, l := range lines {
		padding := strings.Repeat(" ", width-len(l))
		box = append(box, "| "+l+padding+" |")
	}
	box = append(box, border)

	return box
}

func humanDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func showUserInfo(requester, target *Client) {
	now := time.Now()
	uptime := now.Sub(target.ConnectedAt)

	requester.send(ColorBlue + "--------------------------------" + ColorReset)
	requester.send(ColorCyan + "User info: " + target.Name + ColorReset)
	requester.send(ColorBlue + "--------------------------------" + ColorReset)

	requester.send("Online since : " + target.ConnectedAt.Format("15:04:05"))
	requester.send("Online time  : " + humanDuration(uptime))
	requester.send(fmt.Sprintf("Messages    : %d", target.ChatCount))

	if target.Room != nil {
		requester.send("Room        : " + target.Room.Name)
	} else {
		requester.send("Room        : none")
	}

	if len(target.NickHistory) > 0 {
		requester.send("Nick history:")
		for _, h := range target.NickHistory {
			requester.send("  - " + h)
		}
	} else {
		requester.send("Nick history: none")
	}

	if time.Now().Before(target.MutedUntil) {
		remain := int(time.Until(target.MutedUntil).Seconds())
		requester.send(ColorRed +
			fmt.Sprintf("Muted       : yes (%ds left)", remain) +
			ColorReset)
	} else {
		requester.send("Muted       : no")
	}

	requester.send(ColorBlue + "--------------------------------" + ColorReset)
}
