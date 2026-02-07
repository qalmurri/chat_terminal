return function(arg)
    if arg and #arg > 2 then
        local oldFullNick = full_nick 
        if set_nick(arg) then
            broadcast_system("nick_change", oldFullNick .. "|" .. full_nick)
            print(T("success_change_name"))
        else
            print("Gagal mengubah nama.")
        end
    else
        print("Nama minimal 3 karakter!")
    end
end

--    elseif cmd == "nick" then
--        local newName = arg
--        if newName and #newName > 2 then
--            local oldFullNick = full_nick 
--            if set_nick(newName) then
--                broadcast_system("nick_change", oldFullNick .. "|" .. full_nick)
--                print(T("success_change_name"))
--            else
--                print("Gagal mengubah nama.")
--            end
--        else
--            print("Nama minimal 3 karakter!")
--        end
