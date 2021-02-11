package command

import (
	"fmt"

	"github.com/linklux/luxbox-client/component"
)

type RegisterCommand struct {
	command

	component.ServerConnector
}

func (cmd RegisterCommand) New() ICommand {
	return RegisterCommand{
		command{
			"register",
			"Register a new user at the Luxbox server",
			map[string]*commandFlag{},
		},
		component.ServerConnector{},
	}
}

func (cmd RegisterCommand) Execute(args []string) bool {
	request := component.Request{Action: "register", Meta: map[string]string{}}

	response, err := cmd.Send(request)
	if err != nil {
		cmd.printError(err.Error())
		return false
	}

	fmt.Printf("%s", response.Data)

	return true
}
