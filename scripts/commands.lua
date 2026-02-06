function handleCommand(input)
    local cmd, arg = input:match("^/(%w+)%s*(.*)")
    
    -- 1. COMMAND: WHO
    if cmd == "who" then
        local users = get_online()
        if users and #users > 0 then
            local list = ""
            for i, name in ipairs(users) do
                list = list .. name .. " "
            end
            print("\033[33mUser di room ini:\033[0m " .. list)
        else
            print("Hanya ada kamu di sini.")
        end

    -- 2. COMMAND: NICK
    elseif cmd == "nick" then
        local newName = arg
        if newName and #newName > 2 then
            local oldFullNick = full_nick 
            if set_nick(newName) then
                local announcement = "\033[33m[SYSTEM] " .. oldFullNick .. " mengganti nama menjadi " .. full_nick .. "\033[0m"
                print("Nama berhasil diubah!")
                if current_room then
                    broadcast(announcement)
                end
            else
                print("Gagal mengubah nama.")
            end
        else
            print("Nama minimal 3 karakter!")
        end

    -- 3. COMMAND: CREATE (INI YANG TADI HILANG)
    elseif cmd == "create" then
        local name, desc = arg:match("^(%S+)%s*(.*)")
        if name then
            if create_room_db(name, desc or "Tidak ada deskripsi") then
                print("\033[32mRoom '" .. name .. "' berhasil dibuat!\033[0m")
                join_room(name)
                print("\033[36m=== Memasuki Room: " .. name .. " ===\033[0m")
            else
                print("\033[31mGagal membuat room. Nama mungkin sudah dipakai.\033[0m")
            end
        else
            print("Gunakan: /create [nama_room] [deskripsi]")
        end

    -- 4. COMMAND: JOIN
    elseif cmd == "join" then
        if arg ~= "" then
            join_room(arg)
            print("\033[2J\033[H") -- Clear screen
            print("\033[36m=== Memasuki Room: " .. arg .. " ===\033[0m")
            broadcast(full_nick .. " bergabung ke ruangan.")
        else
            print("Gunakan: /join [nama_room]")
        end

    -- 5. COMMAND: LEAVE
    elseif cmd == "leave" then
        if current_room then
            broadcast("\033[33m[SYSTEM] " .. full_nick .. " telah meninggalkan ruangan.\033[0m")
            if leave_room() then
                print("\033[2J\033[H")
                print("Anda telah kembali ke Dashboard.")
                dofile("scripts/dashboard.lua")
            end
        else
            print("Anda tidak sedang berada di dalam room.")
        end
    
    -- 6. COMMAND: EXIT
    elseif cmd == "exit" then
        print("Sampai jumpa lagi!")
        exit_session()

    -- 7. COMMAND: ACCEPT
    elseif cmd == "accept" then
        if current_room then
            if set_room_owner(current_room, uid) then
                print("\033[32mSelamat! Anda sekarang adalah Owner dari room ini.\033[0m")
                broadcast("\033[33m[SYSTEM] " .. full_nick .. " mengambil alih jabatan Owner!\033[0m")
            else
                print("Gagal mengambil alih jabatan.")
            end
        else
            print("Kamu harus berada di dalam room untuk menggunakan perintah ini.")
        end

    -- 8. COMMAND: ROOMS (TAMBAHAN BIAR BISA CEK LIST)
    elseif cmd == "rooms" then
        dofile("scripts/dashboard.lua")

    else
        print("Perintah /" .. (cmd or "unknown") .. " tidak ditemukan.")
    end
end

function OnInput(input)
    if input:sub(1,1) == "/" then
        local status, err = pcall(function() handleCommand(input) end)
        if not status then
            print("\033[31mLua Error:\033[0m " .. err)
        end
    else
        if current_room then
            broadcast(full_nick .. ": " .. input)
        else
            print("\033[31mAnda belum masuk room. Ketik /join [nama_room]\033[0m")
        end
    end
end
