return function()
    local users = get_online()
    if users and #users > 0 then
        local list = table.concat(users, " ")
        print("User di room ini: " .. list)
    else
        print(T("else_who_title"))
    end
end
