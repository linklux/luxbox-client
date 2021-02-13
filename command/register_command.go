package command

import (
	"errors"
	"fmt"

	"github.com/linklux/luxbox-client/component"
	"github.com/linklux/luxbox-client/data"
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

// TODO Do not overwrite current when a user configuration exists
func (cmd RegisterCommand) Execute(args []string) error {
	cmd.Connect()

	request := component.Request{Action: "register", Meta: map[string]interface{}{}}
	response, err := cmd.SendAndDisconnect(request)
	if err != nil {
		return err
	}

	if response.Code != 0 {
		cmd.printError(errors.New(fmt.Sprintf("failed to register new user, response payload: %v\n", response.Data)))
	}

	conf := data.GetConfig()

	conf.User.User = response.Data["user"].(string)
	conf.User.Token = response.Data["token"].(string)

	data.WriteConfig(conf)

	fmt.Printf("successfully registered\n")

	return nil
}
