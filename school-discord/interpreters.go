package schooldiscord

import "github.com/Petrify/simp-core/commands"

func interpreterGuild() (I *commands.Interpreter) {
	I = commands.NewInterpreter()
	I.AddCommand("terminal admin", adminTerminal)
	I.AddCommand("edit", classTerminal)
	I.AddCommand("ping", cmdTest)
	return
}

func adminCommands() (I *commands.Interpreter) {
	I = commands.NewInterpreter()
	// I.AddCommand("db add", dbAdd)
	// I.AddCommand("db del", dbDel)
	// I.AddCommand("db rename", dbRename)
	// I.AddCommand("server clear", serverClear)
	// I.AddCommand("del", chanDel)
	// I.AddCommand("search", search)
	return
}

func classEditCommands() (I *commands.Interpreter) {
	I = commands.NewInterpreter()
	I.AddCommand("search", cmdSearch)
	I.AddCommand("join", cmdJoin)
	I.AddCommand("leave", cmdLeave)
	I.AddCommand("list", cmdList)
	return
}
