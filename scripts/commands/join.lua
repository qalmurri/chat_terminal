--     elseif cmd == "join" then
--         if arg ~= "" then
--             local success = join_room(arg)
--             if success then
--                 clear_screen()
--                 print("=== " .. T("welcome") .. " " .. arg .. " ===")
--                 broadcast_system("join_msg", full_nick)
--             else
--                 print("[ERROR] Room '" .. arg .. "' tidak ditemukan!")
--             end
--         end
return function(arg)
    if arg and arg ~= "" then
        local success = join_room(arg)
        if success then
            clear_screen()
            print("=== " .. T("welcome") .. " " .. arg .. " ===")
            broadcast_system("join_msg", full_nick)
        else
            print("[ERROR] Room '" .. arg .. "' tidak ditemukan!")
        end
    else
        print("Gunakan: /join [nama_room]")
    end
end
