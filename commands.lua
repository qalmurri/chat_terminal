function handle_command(input)
    local cmd, arg = input:match("^(%S+)%s*(.*)$")

    if cmd == "/join" then
        return "join", arg
    elseif cmd == "/nick" then
        return "nick", arg
    elseif cmd == "/rooms" then
        return "rooms", ""
    elseif cmd == "/info" then
        return "info", arg
    elseif cmd == "/help" then
        return "help", ""
    elseif cmd == "/who" then
	return "who", ""
    elseif cmd == "/quit" then
        return "quit", ""
    end

    return "noop", ""
end
