package lua

import (
	"database/sql"
	"fmt"
	"mager/internal/chat"
	"mager/internal/database"
	"github.com/yuin/gopher-lua"
	"golang.org/x/crypto/ssh"
)

func (ls *LuaState) goPrint(ch ssh.Channel) lua.LGFunction {
	return func(L *lua.LState) int {
		msg := L.ToString(1)
		ch.Write([]byte(msg + "\r\n"))
		return 0
	}
}

func (ls *LuaState) goBroadcast(hub *chat.Hub, client *chat.Client) lua.LGFunction {
	return func(L *lua.LState) int {
		msg := L.ToString(1)
		roomObj := L.GetGlobal("current_room")
		if roomObj.Type() == lua.LTString {
			hub.Broadcast <- chat.Message{
				Room:    roomObj.String(),
				Content: msg,
				Sender:  client,
			}
		}
		return 0
	}
}

func (ls *LuaState) goJoinRoom(db *sql.DB, hub *chat.Hub, client *chat.Client) lua.LGFunction {
	return func(L *lua.LState) int {
		roomName := L.ToString(1)
		var exists bool
		db.QueryRow("SELECT EXISTS(SELECT 1 FROM rooms WHERE name = ?)", roomName).Scan(&exists)
		if !exists {
			L.Push(lua.LBool(false))
			return 1
		}
		oldRoom := L.GetGlobal("current_room")
		if oldRoom.Type() == lua.LTString {
			hub.Unregister <- chat.Registration{Room: oldRoom.String(), Client: client}
		}
		hub.Register <- chat.Registration{Room: roomName, Client: client}
		L.SetGlobal("current_room", lua.LString(roomName))
		L.Push(lua.LBool(true))
		return 1
	}
}

func (ls *LuaState) goSetNick(db *sql.DB, user database.User, hub *chat.Hub, client *chat.Client) lua.LGFunction {
	return func(L *lua.LState) int {
		newName := L.ToString(1)
		if err := database.UpdateNickname(db, user.UID, newName); err != nil { return 0 }
		oldFull := L.GetGlobal("full_nick").String()
		newFull := fmt.Sprintf("[%d]%s", user.UID, newName)
		hub.UpdateClientNick(oldFull, newFull)
		L.SetGlobal("nickname", lua.LString(newName))
		L.SetGlobal("full_nick", lua.LString(newFull))
		client.Nick = newFull
		return 1
	}
}

func (ls *LuaState) goClearScreen(ch ssh.Channel) lua.LGFunction {
	return func(L *lua.LState) int {
		ch.Write([]byte("\033[2J\033[H"))
		return 0
	}
}

func (ls *LuaState) goGetRooms(db *sql.DB, hub *chat.Hub) lua.LGFunction {
	return func(L *lua.LState) int {
		rooms, _ := database.GetActiveRooms(db)
		table := L.NewTable()
		for _, r := range rooms {
			rTable := L.NewTable()
			L.SetField(rTable, "name", lua.LString(r.Name))
			L.SetField(rTable, "desc", lua.LString(r.Description))
			L.SetField(rTable, "count", lua.LNumber(hub.GetRoomCount(r.Name)))
			table.Append(rTable)
		}
		L.Push(table)
		return 1
	}
}

func (ls *LuaState) goGetOnline(hub *chat.Hub) lua.LGFunction {
	return func(L *lua.LState) int {
		roomObj := L.GetGlobal("current_room")
		table := L.NewTable()
		if roomObj.Type() == lua.LTString {
			nicks := hub.GetUsersInRoom(roomObj.String())
			for _, nick := range nicks { table.Append(lua.LString(nick)) }
		}
		L.Push(table)
		return 1
	}
}

func (ls *LuaState) goLeaveRoom(db *sql.DB, user database.User, hub *chat.Hub, client *chat.Client) lua.LGFunction {
	return func(L *lua.LState) int {
		roomObj := L.GetGlobal("current_room")
		if roomObj.Type() != lua.LTString { return 0 }
		roomName := roomObj.String()
		hub.Unregister <- chat.Registration{Room: roomName, Client: client}
		L.SetGlobal("current_room", lua.LNil)
		return 1
	}
}

func (ls *LuaState) goBroadcastSystem(hub *chat.Hub, client *chat.Client) lua.LGFunction {
	return func(L *lua.LState) int {
		key, params := L.ToString(1), L.ToString(2)
		roomObj := L.GetGlobal("current_room")
		if roomObj.Type() == lua.LTString {
			hub.Broadcast <- chat.Message{Room: roomObj.String(), Type: "system", Key: key, Params: params, Sender: client}
		}
		return 0
	}
}

func (ls *LuaState) goCreateRoom(db *sql.DB, user database.User) lua.LGFunction {
	return func(L *lua.LState) int {
		name, desc := L.ToString(1), L.ToString(2)
		if err := database.CreateRoom(db, name, user.UID, desc); err != nil { return 0 }
		return 1
	}
}

func (ls *LuaState) goSetRoomOwner(db *sql.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		room, newUID := L.ToString(1), L.ToInt(2)
		if err := database.UpdateRoomOwner(db, room, newUID); err != nil { return 0 }
		return 1
	}
}

func (ls *LuaState) goSetLanguage(db *sql.DB, user database.User) lua.LGFunction {
	return func(L *lua.LState) int {
		lang := L.ToString(1)
		if err := database.UpdateUserLanguage(db, user.UID, lang); err != nil { return 0 }
		L.SetGlobal("user_lang", lua.LString(lang))
		return 1
	}
}

func (ls *LuaState) goExitSession(ch ssh.Channel) lua.LGFunction {
	return func(L *lua.LState) int {
		ch.Close()
		return 0
	}
}
