package commands

func init() {
	RegisterCommand("pwd", HandlePwd)
	RegisterCommand("ls", HandleListFiles)
	RegisterCommand("cat", HandleCat)
	RegisterCommand("echo", HandleEcho)
	RegisterCommand("clear", HandleClear)
	RegisterCommand("exit", HandleExit)
	RegisterCommand("help", HandleHelp)
	RegisterCommand("whoami", HandleWhoami)
}
