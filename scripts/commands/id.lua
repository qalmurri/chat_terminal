--    elseif cmd == "id" then
--        if set_language("id") then
--            user_lang = "id"
--            dofile("scripts/dashboard.lua")
--        end
return function()
    if set_language("id") then
        user_lang = "id"
        dofile("scripts/dashboard.lua")
    end
end

