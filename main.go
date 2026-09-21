package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/beka652/rss_feed_aggregator/internal/config"
	"github.com/beka652/rss_feed_aggregator/internal/database"
	"github.com/google/uuid"
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
	registeredCmds.register("reset", handlerReset)
	registeredCmds.register("users", handlerUsers)
	registeredCmds.register("agg", agg)

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
	s.config.SetUser()
	
	fmt.Println()
	fmt.Println(user)
	fmt.Println("User created successfully!")
	fmt.Println()

	return  nil 
}

func handlerReset(s *state, _ command) error {
	err := s.db.ResetDB(context.Background())
	if err != nil {
		return err 
	}
	s.config.CurrentUserName = ""
	s.config.SetUser()
	fmt.Println("Database reset successfully")
	return nil 
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	currUser := s.config.CurrentUserName
	for _, user := range users {
		if currUser == user.Name {
			fmt.Printf(" * %v (current)\n",user.Name)
		} else {
			fmt.Printf(" * %v \n",user.Name)
		}
	}
	return nil 
}

func agg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	rssFeed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}
	fmt.Println(*rssFeed)
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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	var rssFeed RSSFeed
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet, 
		feedUrl,
		nil,
	)
	if err != nil {
		return &rssFeed, err 
	}
	req.Header.Set("user-agent", "gator")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &rssFeed, err 
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &rssFeed, err 
	}
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		return &rssFeed, err
	}
	sanitizeRSSFeed(&rssFeed)
	return &rssFeed, nil 
}

func sanitizeRSSFeed(feed *RSSFeed) {
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := 0; i < len(feed.Channel.Item); i ++ {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
}

// Helpers 
