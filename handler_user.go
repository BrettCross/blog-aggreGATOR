package main

import (
	"context"
	"fmt"
	"time"

	"github.com/brettcross/blog-aggreGATOR/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	name := cmd.Args[0]

	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user doesn't exist: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	// create new user in db
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(), 
		CreatedAt: time.Now(), 
		UpdatedAt: time.Now(), 
		Name: cmd.Args[0],
	})

	if err != nil {
		return fmt.Errorf("user %w already exists", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("user %v created successfully!\n\n", user.Name)
	fmt.Printf("ID: %v\nCreatedAt: %v\nUpdatedAt: %v\nName: %v\n", user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
	return nil
}