package command

import (
	"errors"
	"flag"
	"fmt"
	"sort"

	"github.com/mgutz/ansi"
)

type ICommand interface {
	New() ICommand

	ParseFlags(args []string)
	Execute(args []string) error

	PrintUsage()
	GetDescription() string
}

type commandFlag struct {
	Name        string
	Shortname   string
	Datatype    string
	Description string
	Value       interface{}
}

type command struct {
	name        string
	description string

	flags map[string]*commandFlag
}

func (this command) ParseFlags(args []string) {
	fs := flag.NewFlagSet(this.name, flag.ExitOnError)
	fs.Usage = func() {
		this.printError(errors.New("Invalid flag given, the following flags are allowed:"))
		this.PrintUsage()
	}

	// TODO There must be a better way to do this...
	bools := map[string]*bool{}
	ints := map[string]*int{}
	strings := map[string]*string{}

	for key, element := range this.flags {
		switch element.Datatype {
		case "bool":
			if defaultValue, ok := element.Value.(bool); ok {
				bools[key] = &defaultValue

				if element.Shortname != "" {
					fs.BoolVar(bools[key], element.Shortname, defaultValue, element.Description)
				}

				fs.BoolVar(bools[key], element.Name, defaultValue, element.Description)
			}

		case "int":
			if defaultValue, ok := element.Value.(int); ok {
				ints[key] = &defaultValue

				if element.Shortname != "" {
					fs.IntVar(ints[key], element.Shortname, defaultValue, element.Description)
				}

				fs.IntVar(ints[key], element.Name, defaultValue, element.Description)
			}

		case "string":
			if defaultValue, ok := element.Value.(string); ok {
				strings[key] = &defaultValue

				if element.Shortname != "" {
					fs.StringVar(strings[key], element.Shortname, defaultValue, element.Description)
				}

				fs.StringVar(strings[key], element.Name, defaultValue, element.Description)
			}
		}
	}

	fs.Parse(args)

	for key, element := range bools {
		this.flags[key].Value = *element
	}

	for key, element := range ints {
		this.flags[key].Value = *element
	}

	for key, element := range strings {
		this.flags[key].Value = *element
	}
}

func (this command) PrintUsage() {
	fmt.Println(this.description)
	fmt.Println("\nAllowed flags:")

	// The order is different from time to time when printing, sort it first
	keys := make([]string, 0)
	for k, _ := range this.flags {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("    -%s|--%s <%s> (default: %s) %s\n",
			this.flags[k].Shortname,
			this.flags[k].Name,
			this.flags[k].Datatype,
			fmt.Sprint(this.flags[k].Value),
			this.flags[k].Description,
		)
	}
}

func (cmd command) GetDescription() string {
	return cmd.description
}

func (cmd command) printError(err error) {
	fmt.Println(ansi.Color(err.Error(), "red"))
}
