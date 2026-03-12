package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/MoD366/bootdev_gator/internal/config"
	"github.com/MoD366/bootdev_gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	config *config.Config
	db     *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmd map[string]func(*state, command) error
}

func SetConf(c *config.Config) state {
	return state{config: c}
}

func CreateCommandsMap() commands {
	return commands{cmd: make(map[string]func(*state, command) error)}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("No arguments provided for command.")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("User %v has been logged in.\n", cmd.args[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("No arguments provided for command.")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		return errors.New("User already exists!")
	}

	user := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]}

	s.db.CreateUser(context.Background(), user)

	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("User %v was created successfully!\n", cmd.args[0])

	log.Printf("User crated: %+v", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return errors.New("Could not delete all users from database.")
	}

	fmt.Println("All users have been deleted from database.")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	names, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Failed to read users from database.")
		return err
	}

	for _, name := range names {
		if name == s.config.CurrentUser {
			fmt.Printf("%v (current)\n", name)
			continue
		}
		fmt.Println(name)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("Missing interval.")
	}
	fmt.Printf("Collecting feeds every %v\n", cmd.args[0])

	interval, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("addfeed command needs two arguments.")
	}

	feed_id := uuid.New()

	s.db.AddFeed(context.Background(), database.AddFeedParams{ID: feed_id, CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID})

	_, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed_id})
	if err != nil {
		return err
	}

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedWithUsername(context.Background())
	if err != nil {
		return err
	}

	for i, feed := range feeds {
		fmt.Printf("Feed %d:\n - Name: %v\n - URL: %v\n - User: %v\n", i, feed.Url, feed.Name, feed.Name_2)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("No URL to follow provided.")
	}

	feed, err := s.db.GetFeedFromUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	res, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return err
	}

	fmt.Printf("%v is now following feed '%v'", res.UserName, res.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	res, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("%v follows these feeds:\n", s.config.CurrentUser)

	for _, val := range res {
		fmt.Printf(" - %v\n", val.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("No URL to unfollow provided.")
	}

	feed, err := s.db.GetFeedFromUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{FeedID: feed.ID, UserID: user.ID})

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) == 1 {
		if argLimit, err := strconv.Atoi(cmd.args[0]); err == nil {
			limit = argLimit
		} else {
			return fmt.Errorf("Invalid limit given %w", err)
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit)})
	if err != nil {
		return err
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Fatal("Failed reading feed to fetch: " + string(err.Error()))
		return
	}
	s.db.MarkFeedFetched(context.Background(), feed.ID)

	rss, err := fetchFeed(context.Background(), feed.Url)

	fmt.Printf("Fetching feed from %v\n", feed.Url)

	for _, item := range rss.Channel.Item {
		notnull := true
		if item.Description == "" {
			notnull = false
		}

		desc := sql.NullString{String: item.Description, Valid: notnull}

		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Println(err)
		}

		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Title: item.Title, Url: item.Link, Description: desc, PublishedAt: pubDate, FeedID: feed.ID})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Println(err)
		}
	}
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.cmd[cmd.name]
	if !ok {
		return errors.New("Command not found.")
	}

	return f(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmd[name] = f
}
