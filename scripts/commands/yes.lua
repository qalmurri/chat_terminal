--    elseif cmd == "yes" then
--        local data = _G.pending_data
--        if data.name then
--            local diff = now - data.time
--            if diff <= 10 then
--                local room_name = data.name
--                data.name = nil -- Reset
--                broadcast_system("leave_msg", full_nick)
--                leave_room()
--                execute_create_room(room_name)
--            else
--                data.name = nil
--                print("[EXPIRED] " .. T("confirmation_expired"))
--            end
--        else
--            print(T("nothing_to_confirm"))
--        end

return function()
    local data = _G.pending_data
    local now = os.time()
    local CONFIRM_TIMEOUT = 10 -- Sesuai dengan yang kamu tulis tadi

    if data and data.name then
        local diff = now - data.time
        
        if diff <= CONFIRM_TIMEOUT then
            local room_name = data.name
            data.name = nil -- Reset data setelah diambil
            
            -- Kirim pesan perpisahan ke room lama
            broadcast_system("leave_msg", full_nick)
            
            -- Keluar dan buat yang baru
            leave_room()
            execute_create_room(room_name)
        else
            data.name = nil -- Reset karena sudah expired
            print("\033[31m[EXPIRED]\033[0m " .. T("confirmation_expired"))
        end
    else
        print(T("nothing_to_confirm"))
    end
end
