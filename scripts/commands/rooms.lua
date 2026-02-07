--    elseif cmd == "rooms" then
--        dofile("scripts/dashboard.lua")
return function()
    -- Memanggil ulang file dashboard untuk menampilkan daftar room terbaru
    local status, err = pcall(function()
        dofile("scripts/dashboard.lua")
    end)
    
    if not status then
        print("Gagal memuat daftar ruangan: " .. tostring(err))
    end
end

