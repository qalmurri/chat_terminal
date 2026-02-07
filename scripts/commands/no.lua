--    elseif cmd == "no" then
--        _G.pending_data.name = nil
--        print(T("action_cancelled"))
return function()
    if _G.pending_data and _G.pending_data.name then
        _G.pending_data.name = nil
        print("\033[32m[OK]\033[0m " .. T("action_cancelled"))
    else
        print(T("nothing_to_confirm"))
    end
end
