package lua

import (
	"database/sql"
	"fmt"
	"mager/internal/chat"
	"mager/internal/database"
	"github.com/yuin/gopher-lua"
	"golang.org/x/crypto/ssh"
)

type LuaState struct {
	L *lua.LState
}

func NewLuaEngine(ch ssh.Channel, user database.User, hub *chat.Hub, db *sql.DB) *LuaState {
	L := lua.NewState()

	// Simpan client sebagai variabel lokal agar bisa diakses closure
	fullNick := fmt.Sprintf("[%d]%s", user.UID, user.Nick)
	client := &chat.Client{Nick: fullNick, Channel: ch}

	// 1. Variabel Global Awal
	L.SetGlobal("uid", lua.LNumber(user.UID))
	L.SetGlobal("nickname", lua.LString(user.Nick))
	L.SetGlobal("full_nick", lua.LString(fullNick))
	L.SetGlobal("current_room", lua.LNil)

	// 2. Bridge Dasar (Print & Broadcast)
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		ch.Write([]byte(msg + "\r\n"))
		return 0
	}))

	L.SetGlobal("broadcast", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		roomObj := L.GetGlobal("current_room")
		if roomObj.Type() == lua.LTString {
			hub.Broadcast <- chat.Message{Room: roomObj.String(), Content: msg}
		}
		return 0
	}))

	// 3. Bridge Room Management
	L.SetGlobal("get_rooms", L.NewFunction(func(L *lua.LState) int {
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
	}))

	L.SetGlobal("join_room", L.NewFunction(func(L *lua.LState) int {
		roomName := L.ToString(1)
		oldRoom := L.GetGlobal("current_room")
		
		if oldRoom.Type() == lua.LTString {
			hub.Unregister <- chat.Registration{Room: oldRoom.String(), Client: client}
		}

		hub.Register <- chat.Registration{Room: roomName, Client: client}
		L.SetGlobal("current_room", lua.LString(roomName))
		return 0
	}))

	L.SetGlobal("create_room_db", L.NewFunction(func(L *lua.LState) int {
		name := L.ToString(1)
		desc := L.ToString(2)
		err := database.CreateRoom(db, name, user.UID, desc)
		if err != nil { return 0 }
		return 1
	}))

	L.SetGlobal("set_room_owner", L.NewFunction(func(L *lua.LState) int {
		roomName := L.ToString(1)
		newUID := L.ToInt(2)
		err := database.UpdateRoomOwner(db, roomName, newUID)
		if err != nil { return 0 }
		return 1
	}))

	// 4. Nickname Management
	L.SetGlobal("set_nick", L.NewFunction(func(L *lua.LState) int {
		newName := L.ToString(1)
		err := database.UpdateNickname(db, user.UID, newName)
		if err != nil { return 0 }

		oldFull := L.GetGlobal("full_nick").String()
		newFull := fmt.Sprintf("[%d]%s", user.UID, newName)
		
	// 1. Update Hub (Memori Global)
	        hub.UpdateClientNick(oldFull, newFull)
	
	        // 2. Update Status Sesi (Memori Lokal Lua)
	        L.SetGlobal("nickname", lua.LString(newName))
	        L.SetGlobal("full_nick", lua.LString(newFull))
	        
	        // 3. Update objek client agar Unregister nanti tidak error
	        client.Nick = newFull 
	        
	        return 1
	}))

	L.SetGlobal("get_online", L.NewFunction(func(L *lua.LState) int {
	    roomObj := L.GetGlobal("current_room")
	    if roomObj.Type() != lua.LTString {
	        L.Push(L.NewTable()) // Kembalikan tabel kosong jika tidak di room
	        return 1
	    }
		
	    
        // Panggil fungsi exported yang baru kita buat
            nicks := hub.GetUsersInRoom(roomObj.String())
            
            table := L.NewTable()
            for _, nick := range nicks {
                table.Append(lua.LString(nick))
            }


	    
	    L.Push(table)
	    return 1
	}))

        L.SetGlobal("leave_room", L.NewFunction(func(L *lua.LState) int {
            roomObj := L.GetGlobal("current_room")
            if roomObj.Type() != lua.LTString { return 0 }
            
            roomName := roomObj.String()
            
            // 1. Cek apakah user ini owner di DB
            ownerUID, _ := database.GetRoomOwner(db, roomName)
            isOwner := (ownerUID == int(user.UID))
        
            // 2. Cek sisa orang di Hub
            userCount := hub.GetRoomCount(roomName)
        
            // 3. Jika owner keluar dan masih ada orang lain
            if isOwner && userCount > 1 {
                hub.Broadcast <- chat.Message{
                    Room:    roomName,
                    Content: "\r\n\033[35m[SYSTEM] Owner telah pergi. Ketik /accept untuk menjadi owner baru!\033[0m",
                }
            }

    // 4. Proses Unregister
    hub.Unregister <- chat.Registration{Room: roomName, Client: client}
    L.SetGlobal("current_room", lua.LNil)
    
    return 1
}))

	L.SetGlobal("exit_session", L.NewFunction(func(L *lua.LState) int {
	    ch.Close()
	    return 0
	}))

	return &LuaState{L: L}
}

func (ls *LuaState) HandleInput(input string) {
	err := ls.L.CallByParam(lua.P{
		Fn:      ls.L.GetGlobal("OnInput"),
		NRet:    0,
		Protect: true,
	}, lua.LString(input))
	if err != nil {
		ls.L.DoString("print('Error: input tidak bisa diproses')")
	}
}
