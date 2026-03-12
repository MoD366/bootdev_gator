package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/MoD366/bootdev_gator/internal/config"
	"github.com/MoD366/bootdev_gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	programState := &state{
		config: &conf,
	}

	db, err := sql.Open("postgres", programState.config.Dburl)

	programState.db = database.New(db)

	commandsList := &commands{
		cmd: make(map[string]func(*state, command) error),
	}

	commandsList.register("login", handlerLogin)
	commandsList.register("register", handlerRegister)
	commandsList.register("reset", handlerReset)
	commandsList.register("users", handlerUsers)
	commandsList.register("agg", handlerAgg)
	commandsList.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commandsList.register("feeds", handlerFeeds)
	commandsList.register("follow", middlewareLoggedIn(handlerFollow))
	commandsList.register("following", middlewareLoggedIn(handlerFollowing))
	commandsList.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commandsList.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args

	if len(args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
		return
	}

	cmdName := args[1]
	cmdArgs := args[2:]

	err = commandsList.run(programState, command{name: cmdName, args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
}
