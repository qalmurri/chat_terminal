--    elseif cmd == "leave" then
--        if current_room then
--            broadcast_system("leave_msg", full_nick)
--            if leave_room() then
--                print("Anda telah kembali ke Dashboard.")
--                dofile("scripts/dashboard.lua")
--            end
--        else
--            print("Anda tidak sedang berada di dalam room.")
--        end
return function()
    if current_room then
        broadcast_system("leave_msg", full_nick)
        if leave_room() then
            print("Anda telah kembali ke Dashboard.")
            dofile("scripts/dashboard.lua")
        end
    else
        print("Anda tidak sedang berada di dalam room.")
    end
end
