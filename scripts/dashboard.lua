print("\033[2J\033[H") -- Clear screen
print("==========================================")
print("      SELAMAT DATANG DI MAGER CHAT")
print("==========================================")
print("ID Anda   : [" .. uid .. "]")
print("Nickname : " .. nickname)
print("------------------------------------------")
print("ROOMS YANG AKTIF:")

local rooms = get_rooms()
if #rooms == 0 then
    print(" (Belum ada room aktif)")
else
    for i, r in ipairs(rooms) do
        print(i .. ". " .. r.name .. " (" .. r.count .. " user) - " .. r.desc)
    end
end

print("------------------------------------------")
print("COMMANDS:")
print(" /join [nama]         - Masuk ke ruangan")
print(" /create [nama] [des] - Buat ruangan baru")
print(" /nick [nama_baru]    - Ganti nama")
print(" /exit                - Keluar")
print("==========================================")
