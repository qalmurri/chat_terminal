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
        // Kita kirim 'db' ke handleConnection agar bisa digunakan di sana
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
    err = db.QueryRow("SELECT uid, nickname FROM users WHERE pubkey_fingerprint = ?", userFP).Scan(&user.UID, &user.Nick)
    if err != nil {
        user = database.User{UID: 0, Nick: userNick}
    }

    go gossh.DiscardRequests(reqs)

    for newChannel := range chans {
        // CEK APAKAH USER MEMINTA SESSION (TERMINAL)
        if newChannel.ChannelType() != "session" {
            newChannel.Reject(gossh.UnknownChannelType, "hanya menerima session")
            continue
        }

        // DEFINISIKAN VARIABLE channel DI SINI
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

        // REGISTRASI KE HUB
// 	client := &chat.Client{
// 	    Nick: fmt.Sprintf("[%d]%s", user.UID, user.Nick), 
// 	    Channel: channel,
// 	}
// 	hub.Register <- client

	client := &chat.Client{Nick: fmt.Sprintf("[%d]%s", user.UID, user.Nick), Channel: channel}

        // SETUP TERMINAL & LUA
        termObj := term.NewTerminal(channel, "> ")
        luaEngine := lua.NewLuaEngine(channel, user, hub, db)

        luaEngine.L.DoFile("scripts/dashboard.lua")
        luaEngine.L.DoFile("scripts/commands.lua")

        // LOOP INPUT
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

        // UNREGISTER SAAT DISCONNECT
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
