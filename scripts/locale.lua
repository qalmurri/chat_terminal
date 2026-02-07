local texts = {
    id = {
        dash_title = "=== DAFTAR RUANGAN AKTIF ===",
        dash_no_room = "Belum ada ruangan. Ketik /create [nama] untuk membuat.",
        dash_instruction = "Ketik /join [nama_room] untuk masuk.",
        dash_header = string.format("%-15s | %-20s | %-10s", "Nama Room", "Deskripsi", "User"),
        dash_line = "------------------------------------------------------------",
        welcome = "Selamat Datang di MAGER Chat Server, ",
        who_title = "User di room ini: ",
	else_who_title = "Hanya ada kamu disini.",
	success_change_name = "Nama berhasil diubah!!",
        no_room = "Silakan join room dulu!",
        cmd_not_found = "Perintah tidak ditemukan.",
        owner_left = "[SYSTEM] Owner telah pergi. Ketik /accept untuk menjadi owner baru!",
	nick_change = "[SISTEM] %s mengganti nama menjadi %s",
        join_msg = "[SISTEM] %s bergabung ke ruangan.",
        leave_msg = "[SISTEM] %s keluar dari ruangan.",
        owner_succession = "[SISTEM] Owner telah pergi. Ketik /accept untuk menjadi owner baru!",
	exit = "Sampai jumpa lagi!",
	help_title = "=== DAFTAR PERINTAH MAGER ===",
        help_rooms  = "Melihat daftar ruangan yang aktif",
        help_join   = "Masuk ke dalam ruangan chat",
        help_create = "Membuat ruangan chat baru",
        help_nick   = "Mengganti nama panggilan Anda",
        help_who    = "Melihat siapa saja yang ada di ruangan",
        help_lang   = "Mengubah bahasa (Indonesia/English)",
        help_leave  = "Keluar dari ruangan ke Dashboard",
        help_exit   = "Memutuskan koneksi dari server",
	confirm_leave_to_create = "Anda sedang berada di dalam ruangan. Ingin keluar dan membuat ruangan baru?",
        nothing_to_confirm = "Tidak ada aksi yang perlu dikonfirmasi.",
        action_cancelled = "Aksi dibatalkan.",
	confirmation_expired = "Waktu konfirmasi telah habis. Silakan ulangi perintah /create.",
    },
    en = {
	dash_title = "=== ACTIVE ROOMS LIST ===",
        dash_no_room = "No rooms available. Type /create [name] to create one.",
        dash_instruction = "Type /join [room_name] to enter.",
        dash_header = string.format("%-15s | %-20s | %-10s", "Room Name", "Description", "Users"),
        dash_line = "------------------------------------------------------------",
        welcome = "Welcome to MAGER Chat Server, ",
        who_title = "Users in this room: ",
	else_who_title = "Only you.",
	success_change_name = "name success change",
        no_room = "Please join a room first!",
        cmd_not_found = "Command not found.",
        owner_left = "[SYSTEM] Owner has left. Type /accept to become the new owner!",
	nick_change = "[SYSTEM] %s changed name to %s",
        join_msg = "[SYSTEM] %s joined the room.",
        leave_msg = "[SYSTEM] %s has left the room.",
        owner_succession = "[SYSTEM] Owner has left. Type /accept to become the new owner!",
	exit = "See you again!",
	help_title = "=== MAGER COMMAND LIST ===",
        help_rooms  = "Show active rooms list",
        help_join   = "Join a chat room",
        help_create = "Create a new chat room",
        help_nick   = "Change your nickname",
        help_who    = "List users in current room",
        help_lang   = "Change language (Indonesian/English)",
        help_leave  = "Leave room to Dashboard",
        help_exit   = "Disconnect from server",
	confirm_leave_to_create = "You are currently in a room. Leave and create a new one?",
        nothing_to_confirm = "Nothing to confirm.",
        action_cancelled = "Action cancelled.",
	confirmation_expired = "Confirmation timed out. Please repeat the /create command.",
    },
}

function ParseSystemMessage(payload)
    payload = payload:gsub("%z", ""):gsub("[\r\n]", "")
    local data_tag = "__DATA__"
    local system_tag = "__SYSTEM__"
    local system_start = payload:find(system_tag, 1, true)
    local data_start = payload:find(data_tag, 1, true)
    if not system_start or not data_start then
        return payload
    end
    local key = payload:sub(system_start + #system_tag, data_start - 1)
    local data = payload:sub(data_start + #data_tag)
    local lang = user_lang or "en"
    local template = (texts[lang] and texts[lang][key]) or key
    local params = {}
    if data and data ~= "" then
        local search_from = 1
        while true do
            local find_start, find_end = data:find("|", search_from, true) -- true = plain text search
            if not find_start then
                local p = data:sub(search_from)
                if p ~= "" then table.insert(params, p) end
                break
            end
            local p = data:sub(search_from, find_start - 1)
            table.insert(params, p)
            search_from = find_end + 1
        end
    end
    local status, result = pcall(function()
        if #params == 2 then
            -- Ini untuk nick_change (membutuhkan 2 %s)
            return string.format(template, params[1], params[2])
        elseif #params == 1 then
            -- Ini untuk join_msg atau leave_msg (membutuhkan 1 %s)
            return string.format(template, params[1])
        else
            return template
        end
    end)
    if status then
        return result
    else
        return template .. " (Data: " .. data .. ")"
    end
end

function T(key)
    local lang = user_lang or "en"
    if texts[lang] and texts[lang][key] then
        return texts[lang][key]
    end
    return key
end
