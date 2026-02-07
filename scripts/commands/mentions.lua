return function()
    local users = get_online()
    if users and #users > 0 then
        print("\n\033[36m--- User yang bisa kamu mention ---\033[0m")
        for i, full_name in ipairs(users) do
            -- Kita ambil bagian namanya saja setelah [ID]
            local name_only = full_name:match("%]%s*(.*)") or full_name
            print("\033[32m@" .. name_only .. "\033[0m")
        end
        print("-----------------------------------\n")
    else
        print(T("else_who_title"))
    end
end
