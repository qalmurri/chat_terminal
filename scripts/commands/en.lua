--    elseif cmd == "en" then
--        if set_language("en") then
--            user_lang = "en"
--            dofile("scripts/dashboard.lua")
--        end
return function()
    if set_language("en") then
        user_lang = "en"
        dofile("scripts/dashboard.lua")
    end
end

