dofile("scripts/locale.lua")

_G.pending_data = _G.pending_data or { name = nil, time = 0 }

function OnMessageReceived(msg)
    -- Ambil UID dari Global
    local current_uid = tonumber(uid) or 0
    local my_uid_str = string.format("%.0f", current_uid)
    local pattern = "@" .. my_uid_str

    -- 1. Cek System Message (Gunakan fungsi dari locale.lua)
    if msg:find("__SYSTEM__", 1, true) then
        print(ParseSystemMessage(msg))
        return
    end

    -- 2. Logika Mention
    if msg:find(pattern, 1, true) then
        io.write("\a") -- Bunyi Beep
        print(string.rep("=", 35))
        local highlighted = msg:gsub(pattern, pattern )
        print("[MENTION]" .. highlighted)
        print(string.rep("=", 35))
    else
        print(msg)
    end
end

function execute_create_room(name)
    local desc = "Room baru dibuat via konfirmasi"
    if create_room_db(name, desc) then
        print("Room '" .. name .. "' berhasil dibuat!")
        join_room(name)
        print("=== Memasuki Room: " .. name .. " ===")
    else
        print("Gagal membuat room. Nama mungkin sudah dipakai.")
    end
end

function OnInput(input)
    if input:sub(1,1) == "/" then
        handleCommand(input)
    else
        if current_room then
            broadcast(full_nick .. ": " .. input)
        else
            print(T("no_room"))
        end
    end
end

function handleCommand(input)
    local cmd, arg = input:match("^/(%w+)%s*(.*)")
    if not cmd then return end

    local path = "scripts/commands/" .. cmd .. ".lua"
    
    local f, err = loadfile(path)
    
    if f then
        local command_func = f()
        command_func(arg)
    else
        print("Perintah /" .. cmd .. " tidak ditemukan.")
    end
end

