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

func (cmd RegisterCommand) Execute(args []string) error {
	request := component.Request{Action: "register", Meta: map[string]interface{}{}}

	cmd.Connect()

	response, err := cmd.SendAndDisconnect(request)
	if err != nil {
		return err
	}

	fmt.Printf("%s", response.Data)

	return nil
}
