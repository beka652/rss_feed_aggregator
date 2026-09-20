package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/beka652/rss_feed_aggregator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string 
	args []string 
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func main() {
	cnf, err := config.Read()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 
	}
	st := state {config: cnf}
	registeredCmds := commands{ cmds: map[string]func(*state, command) error{}}
	registeredCmds.register("login", handlerLogin)

	args := os.Args
	if len(args)  < 2 {
		fmt.Fprintln(os.Stderr, "Error: Not enough arguments provided.")
		os.Exit(1)
	}
	cd := command{
		name: args[1],
		args: args[2:],
	}
	err = registeredCmds.run(&st, cd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
		
}

/*
	Handlers 
*/
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Error: Username required.")
	}
	s.config.CurrentUserName = cmd.args[0]
	s.config.SetUser()
	fmt.Printf("User set to %v\n", cmd.args[0])
	return nil 
}

/*
	Commands' struct methods  
*/

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f 
}

func (c *commands) run(s *state, cmd command) error {
	cd, ok := c.cmds[cmd.name] // gets the command handler using its name
	if !ok {
		return errors.New(cmd.name + " not found!")
	} 
	err := cd(s, cmd)
	if err != nil {
		return err
	}
	return nil 
}

// Helpers 
