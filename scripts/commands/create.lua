return function(arg)
    if not arg or arg == "" then
        print("Gunakan: /create [nama_room]")
        return
    end

    if current_room then
        _G.pending_data.name = arg
        _G.pending_data.time = os.time()
        print("\n\033[33m[KONFIRMASI - 10 Detik]\033[0m")
        print(T("confirm_leave_to_create"))
        print("Ketik \033[32m/yes\033[0m atau \033[31m/no\033[0m")
    else
        execute_create_room(arg)
    end
end

--    elseif cmd == "create" then
--        if not arg or arg == "" then
--            print("Gunakan: /create [nama_room]")
--            return
--        end
--        if current_room then
--            -- Simpan ke Global
--            _G.pending_data.name = arg
--            _G.pending_data.time = now
--            print("\n\033[33m[KONFIRMASI - 30 Detik]\033[0m")
--            print(T("confirm_leave_to_create"))
--            print("Ketik \033[32m/yes\033[0m atau \033[31m/no\033[0m")
--        else
--            execute_create_room(arg)
--        end
