package commander

import "github.com/sadeepa24/connected_bot/update"

// This is optinal not using yet
//Deprecated

type Command struct {
	Condition []func() bool
	Cmd       []func(*update.Updatectx) error
}

type Commander struct {
	allCommands map[string]*Command
}

func New(commands []string, handler []*Command) *Commander {
	cmdm := make(map[string]*Command)
	for i, comman := range commands {
		cmdm[comman] = handler[i]
	}
	return nil
}

func (c *Commander) Excute(cmdname string, upx *update.Updatectx) error {
	cmdes := c.allCommands[cmdname]
	for i, cmd := range cmdes.Cmd {
		if cmdes.Condition[i]() {
			return cmd(upx)
		}
	}
	return nil
}
