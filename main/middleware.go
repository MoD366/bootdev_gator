package main

import (
	"context"

	"github.com/MoD366/bootdev_gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUser)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}
