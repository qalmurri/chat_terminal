package main
import (
	"fmt"
	"log"
	"mager/internal/chat"
	"mager/internal/database"
	"mager/internal/lua"
	"mager/internal/ssh"
	"net"
	"os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)
func main() {
	db, err := database.InitDB("./chat.db")
	if err != nil {
		log.Fatal("Gagal inisialisasi DB:", err)
	}
	hub := chat.NewHub(db)
	go hub.Run()
	config := ssh.CreateServerConfig(db)
	privateBytes, err := os.ReadFile("host_key")
	if err != nil {
		log.Fatal("Gagal membaca host_key. Jalankan: ssh-keygen -t rsa -f host_key -N ''")
	}
	private, err := gossh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Gagal parse host_key")
	}
	config.AddHostKey(private)
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatal("Gagal listen di port 2222:", err)
	}
	fmt.Println("Chat Server berjalan di port 2222...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, config, hub, db)
	}
}
func handleConnection(nConn net.Conn, config *gossh.ServerConfig, hub *chat.Hub, db *sql.DB) {
    sshConn, chans, reqs, err := gossh.NewServerConn(nConn, config)
    if err != nil {
        return
    }
    defer sshConn.Close()
    userNick := sshConn.Permissions.Extensions["nick"]
    userFP := sshConn.Permissions.Extensions["fp"]
    var user database.User
    err = db.QueryRow("SELECT uid, nickname, language FROM users WHERE pubkey_fingerprint = ?", userFP).Scan(&user.UID, &user.Nick, &user.Language)
    if err != nil {
	    user = database.User{UID: 0, Nick: userNick, Language: ""}
    }
    go gossh.DiscardRequests(reqs)
    for newChannel := range chans {
        if newChannel.ChannelType() != "session" {
            newChannel.Reject(gossh.UnknownChannelType, "hanya menerima session")
            continue
        }
        channel, requests, err := newChannel.Accept()
        if err != nil {
            continue
        }
        go func(in <-chan *gossh.Request) {
            for req := range in {
                if req.Type == "shell" {
                    req.Reply(true, nil)
                }
            }
        }(requests)
	client := &chat.Client{
	        Nick:    fmt.Sprintf("[%d]%s", user.UID, user.Nick), 
	        Channel: channel,
	        MsgChan: make(chan chat.Message, 100), // Beri buffer!
	}
        termObj := term.NewTerminal(channel, "> ")
        luaEngine := lua.NewLuaEngine(channel, user, hub, db, client)

	luaEngine.StartWatcher()


	go func() {
	    for msg := range client.MsgChan {
	        var rawPayload string
	        if msg.Type == "system" {
	            rawPayload = fmt.Sprintf("__SYSTEM__%s__DATA__%s", msg.Key, msg.Params)
	        } else {
	            rawPayload = msg.Content
	        }
		// fmt.Printf("GO DEBUG: Kirim ke Lua [%s]\n", rawPayload)
		log.Printf("GO DEBUG: Kirim ke Lua [%s]\n", rawPayload)
	        luaEngine.DisplayMessage(rawPayload)
	    }
	}()

	// if err := luaEngine.L.DoFile("scripts/commands.lua"); err != nil {
	//     fmt.Fprintf(channel, "Error loading commands: %v\r\n", err)
	// }

	if err := luaEngine.L.DoFile("scripts/start.lua"); err != nil {
	    fmt.Fprintf(channel, "Error starting session: %v\r\n", err)
	}
        for {
            line, err := termObj.ReadLine()
            if err != nil {
                break
            }
            if line == "" {
                continue
            }
            luaEngine.HandleInput(line)
        }
	roomObj := luaEngine.L.GetGlobal("current_room")
        if roomObj.Type().String() == "string" {
            hub.Unregister <- chat.Registration{
                Room:   roomObj.String(),
                Client: client,
            }
        }
	channel.Close()
    }
}
