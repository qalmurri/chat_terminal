--    elseif cmd == "accept" then
--	if current_room then
--		if set_room_owner(current_room, uid) then
--			print("Selamat! Anda sekarang adalah Owner dari room ini.")
--			broadcast("[SYSTEM] " .. full_nick .. " mengambil alih jabatan Owner!")
--		else
--			print("Gagal mengambil alih jabatan.")
--		end
--	else
--		print("Kamu harus berada di dalam room untuk menggunakan perintah ini.")
--	end
return function()
    if current_room then
        -- uid biasanya global dari engine.go
        if set_room_owner(current_room, uid) then
            print("Selamat! Anda sekarang adalah Owner dari room ini.")
            broadcast("[SYSTEM] " .. full_nick .. " mengambil alih jabatan Owner!")
        else
            print("Gagal mengambil alih jabatan.")
        end
    else
        print("Kamu harus berada di dalam room untuk menggunakan perintah ini.")
    end
end
