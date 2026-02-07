dofile("scripts/locale.lua")

if user_lang == nil or user_lang == "" or user_lang == "nil" then
    print(T("welcome"))
    print("Please select your language / Silakan pilih bahasa:")
    print("--------------------------------------------------")
    print(" 1. Bahasa Indonesia (type: /id)")
    print(" 2. English          (type: /en)")
    print("--------------------------------------------------")
else
    dofile("scripts/dashboard.lua")
end
