package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"
	"github.com/google/uuid"
	"github.com/beka652/rss_feed_aggregator/internal/config"
	"github.com/beka652/rss_feed_aggregator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	config *config.Config
	db *database.Queries
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
	registeredCmds.register("register", handlerRegister)

	db, err := sql.Open("postgres", st.config.DbUrl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	st.db = dbQueries
	
	
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
	user, err := s.db.GetUser(
		context.Background(),
		cmd.args[0],
	)
	if err != nil {
		return err
	}
	s.config.CurrentUserName = user.Name
	s.config.SetUser()
	fmt.Printf("User set to %v\n", user.Name)
	return nil 
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Error: Username required")
	}
	user, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name: cmd.args[0],
		},
	)
	if err != nil {
		return err
	}
	s.config.CurrentUserName = user.Name
	
	fmt.Println()
	fmt.Println(user)
	fmt.Println("User created successfully!")
	fmt.Println()

	return  nil 
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
