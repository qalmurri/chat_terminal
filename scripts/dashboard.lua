dofile("scripts/locale.lua")

local rooms = get_rooms()

clear_screen()
print(T("dash_title"))
print(T("dash_line"))
print(T("dash_header"))
print(T("dash_line"))

if #rooms == 0 then
    print("  " .. T("dash_no_room"))
else
    for _, r in ipairs(rooms) do
        local row = string.format("%-15s | %-20s | %-10d", r.name, r.desc, r.count)
        print(row)
    end
end

print(T("dash_line"))
print(T("dash_instruction"))
print("")
