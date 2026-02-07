package lua

import (
	"database/sql"
	"mager/internal/chat"
	"mager/internal/database"
	"github.com/yuin/gopher-lua"
	"golang.org/x/crypto/ssh"
)

// RegisterBindings adalah "Daftar Menu" untuk Lua
func (ls *LuaState) RegisterBindings(ch ssh.Channel, user database.User, hub *chat.Hub, db *sql.DB, client *chat.Client) {
	
	// Map ini menghubungkan Nama di Lua -> Fungsi di functions.go
	exports := map[string]lua.LGFunction{
		"print":            ls.goPrint(ch),
		"broadcast":        ls.goBroadcast(hub, client),
		"get_rooms":        ls.goGetRooms(db, hub),
		"join_room":        ls.goJoinRoom(db, hub, client),
		"broadcast_system": ls.goBroadcastSystem(hub, client),
		"create_room_db":   ls.goCreateRoom(db, user),
		"set_room_owner":   ls.goSetRoomOwner(db),
		"clear_screen":     ls.goClearScreen(ch),
		"set_nick":         ls.goSetNick(db, user, hub, client),
		"get_online":       ls.goGetOnline(hub),
		"set_language":     ls.goSetLanguage(db, user),
		"leave_room":       ls.goLeaveRoom(db, user, hub, client),
		"exit_session":     ls.goExitSession(ch),
	}

	for name, fn := range exports {
		ls.L.SetGlobal(name, ls.L.NewFunction(fn))
	}
}
