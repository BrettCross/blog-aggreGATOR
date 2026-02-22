package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/brettcross/blog-aggreGATOR/internal/database"
	"github.com/google/uuid"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}

	fmt.Println(feed.Channel.Title)
	fmt.Println(feed.Channel.Link)
	fmt.Println(feed.Channel.Description)
	for _, item := range feed.Channel.Item {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.Description)
		fmt.Println(item.PubDate)
		fmt.Println()
		fmt.Println()
	}

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	curUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Error retrieving user: %w", err)
	}

	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]

	s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(), 
		CreatedAt: time.Now(), 
		UpdatedAt: time.Now(), 
		Name: feedName,
		Url: feedUrl,
		UserID: curUser.ID,
	})

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Couldn't retrieve feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("%s %s %s\n", feed.FeedName, feed.Url, feed.UserName)
	}

	return nil 
}