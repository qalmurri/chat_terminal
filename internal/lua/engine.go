package lua

import (
	"database/sql"
	"fmt"
	"log"
	"mager/internal/chat"
	"mager/internal/database"
	"sync"
	"github.com/fsnotify/fsnotify"
	"github.com/yuin/gopher-lua"
	"golang.org/x/crypto/ssh"
)

type LuaState struct {
	L  *lua.LState
	mu sync.Mutex
}

func NewLuaEngine(ch ssh.Channel, user database.User, hub *chat.Hub, db *sql.DB, client *chat.Client) *LuaState {
	L := lua.NewState()
	ls := &LuaState{L: L}

	// 1. Set Global Variables
	fullNick := fmt.Sprintf("[%d]%s", user.UID, user.Nick)
	L.SetGlobal("uid", lua.LNumber(user.UID))
	L.SetGlobal("nickname", lua.LString(user.Nick))
	L.SetGlobal("full_nick", lua.LString(fullNick))
	L.SetGlobal("current_room", lua.LNil)
	L.SetGlobal("user_lang", lua.LString(user.Language))

	// 2. Register all Go-to-Lua functions from functions.go
	ls.RegisterBindings(ch, user, hub, db, client)

	// 3. Initial Load
	ls.mu.Lock()
	if err := L.DoFile("scripts/commands.lua"); err != nil {
		log.Printf("Error loading initial commands: %v", err)
	}
	ls.mu.Unlock()

	return ls
}

func (ls *LuaState) HandleInput(input string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	fn := ls.L.GetGlobal("OnInput")
	if fn.Type() == lua.LTFunction {
		ls.L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LString(input))
	}
}

func (ls *LuaState) DisplayMessage(rawMsg string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	fn := ls.L.GetGlobal("OnMessageReceived")
	if fn.Type() != lua.LTFunction {
		ls.L.DoString(fmt.Sprintf("print([[%s]])", rawMsg))
		return
	}
	ls.L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LString(rawMsg))
}

func (ls *LuaState) StartWatcher() {
	watcher, _ := fsnotify.NewWatcher()
	go func() {
		for {
			select {
			case event, _ := <-watcher.Events:
				if event.Has(fsnotify.Write) {
					log.Printf("Hot Reloading: %s", event.Name)
					ls.mu.Lock()
					ls.L.DoFile(event.Name)
					ls.mu.Unlock()
				}
			case err, _ := <-watcher.Errors:
				log.Println("Watcher error:", err)
			}
		}
	}()
	watcher.Add("scripts")
	watcher.Add("scripts/commands")
}
