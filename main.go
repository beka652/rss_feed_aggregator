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
	"strconv"
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
	registeredCmds.register("agg", handlerAgg)
	registeredCmds.register("addfeed",middlewareLoggedIn(handlerAddFeed) )
	registeredCmds.register("feeds", handlerFeeds)
	registeredCmds.register("follow", middlewareLoggedIn(handlerFollow))
	registeredCmds.register("following", middlewareLoggedIn(handlerFollowing))
	registeredCmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	registeredCmds.register("browse", middlewareLoggedIn(handlerBrowse))

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

func handlerAgg(s *state, cmd command) error {
	dur  := time.Second * 10
	fmt.Println("collecting feeds every", dur)

	ticker := time.NewTicker(dur)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return err 
		}
		fmt.Println("\n\n\n")
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return errors.New("Invalid argument number.")
	}
	feed, err := s.db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID: uuid.New(),
			UserID: user.ID,
			Name: cmd.args[0],
			Url: cmd.args[1],
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)
	_, err = s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID: uuid.New(),
			UserID: feed.UserID,
			FeedID: feed.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		return err 
	}
	fmt.Println(feed)
	fmt.Println("Feed added and subscribed  successfully!")

	return nil 
}

func handlerFeeds(c *state, cmd command) error {
	feeds, err := c.db.GetFeeds(context.Background())
	if err != nil {
		return err 
	}
	for _, feed := range feeds {
		fmt.Printf(" * feed: %v, url: %v, username: %v\n", feed.FeedName, feed.Url, feed.UserName )
	}
	return nil 
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("Invalid arguement number")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err 
	}
	_ , err = s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID: uuid.New(),
			UserID: user.ID,
			FeedID: feed.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		return err 
	}
	fmt.Printf("%v successfully started following %v\n", user.Name, feed.Name)
	

	return nil 
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("Invalid number of arguments")
	}
	err := s.db.DeleteFeedFollowByUrl(
		context.Background(),
		database.DeleteFeedFollowByUrlParams{
			UserID: user.ID,
			Url: cmd.args[0],
		},
	)
	if err != nil {
		return err 
	}
	fmt.Println(user.Name, "unfollowed", cmd.args[0], "successfully!")
	return nil 
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return errors.New("Unknown number of arguements")
	}
	
	userFeedFollows, err := s.db.GetFeedFollowsForUser(
		context.Background(),
		user.ID,
	)
	if err != nil {
		return err 
	}
	fmt.Printf("* Feeds followed by %v:\n", s.config.CurrentUserName)
	for i , feed := range userFeedFollows {
		fmt.Printf("%v: %v\n", i +1, feed.FeedName)
	}
	return nil 
}
func handlerBrowse(s *state, cmd command, user database.User ) error {
	
	if len(cmd.args) == 0 { // to be removed later
		cmd.args = append(cmd.args, "invalid")
	}
	i64, err  := strconv.ParseInt(cmd.args[0], 10, 32)
	if err != nil {
		i64 = 2
	}
	limit := int32(i64)
	posts, err := s.db.GetPostsForUser(
		context.Background(),
		database.GetPostsForUserParams{
			UserID: user.ID,
			Limit: limit,
		},
	)
	if err != nil {
		return err 
	}	
	err = printPosts(posts)
	if err != nil {
		return  err 
	}
	return nil 
}
func printPosts(posts []database.Post) error {
	for i, post := range posts {
		fmt.Println(" ---------------------------------------------------------------------------")
		fmt.Println(" >> Feed number:", i + 1)
		fmt.Println()
		fmt.Printf(" * ID : %v\n", post.ID)
		fmt.Printf(" * FeedID : %v\n", post.FeedID)
		fmt.Printf(" * Url : %v\n", post.Url)
		fmt.Printf(" * Title: %v\n", post.Title)
		fmt.Printf(" * Description: %v\n", post.Description)
		fmt.Printf(" * Published at: %v\n", post.PublishedAt)
		fmt.Printf(" * Created at at: %v\n", post.CreatedAt)
		fmt.Printf(" * Updated  at: %v\n", post.UpdatedAt)	
		fmt.Println(" ---------------------------------------------------------------------------")
		fmt.Println()
	}
	return nil 
}
func scrapeFeeds(s *state) error {
	nextFeed, err  := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err 
	}
	fmt.Printf("fetched %v\n", nextFeed.Url)
	err = s.db.MarkFeedFetched(
		context.Background(), 
		database.MarkFeedFetchedParams{
			ID: nextFeed.ID,
			LastFetchedAt: sql.NullTime{ Time: time.Now(), Valid: true},
			UpdatedAt: time.Now(),
		})
	if err != nil {
		return err 
	}
	feed, err := fetchFeed(
		context.Background(),
		nextFeed.Url,
	)
	if err != nil {
		return err 
	}
	savePosts(s, feed, nextFeed.ID)
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

func savePosts(s *state, rssFeed *RSSFeed, FeedId uuid.UUID) error {
	for _,  item := range rssFeed.Channel.Item {
		err := s.db.CreatePost(
			context.Background(),
			database.CreatePostParams{
				ID: uuid.New(),
				FeedID: FeedId,
				Title: item.Title,
				Url: item.Link,
				Description: item.Description,
				PublishedAt: ParseDateToNullTime(item.PubDate),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	return nil 
}

func ParseDateToNullTime(dateStr string) sql.NullTime {
	if dateStr == "" || dateStr== "null"{
		return sql.NullTime{Valid: false}
	}
	layouts := []string{
			time.RFC1123Z, // "Tue, 22 Sep 2026 06:08:39 +0000"
			time.RFC1123,  // "Tue, 22 Sep 2026 06:08:39 UTC"
			time.RFC3339,  // "2026-09-22T06:08:39Z"
			"2006-01-02",  // Custom "YYYY-MM-DD" date-only format
	}
	for _, layout := range layouts {
		if parsedTime, err := time.Parse(layout, dateStr); err != nil  {
			return sql.NullTime{
				Time: parsedTime,
				Valid: true,
			}
		}
	}
	return sql.NullTime{Valid: false}
}

func sanitizeRSSFeed(feed *RSSFeed) {
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := 0; i < len(feed.Channel.Item); i ++ {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
}
// Middle ware 
func middlewareLoggedIn( handler func(s *state, cmd command, user database.User) error,
	) func (*state, command) error {
	return func(s *state, cmd command) error {
		user, err  := s.db.GetUser(context.Background(),s.config.CurrentUserName)
		if err != nil {
			return err 
		}
		err = handler(s, cmd, user)
		if err != nil {
			return err 
		}
		return err 
	}
}
