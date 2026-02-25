package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/brettcross/gator/internal/database"
	"github.com/google/uuid"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time_between_reqs>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error parsing arg to time: %w", err)
	}

	fmt.Printf("Collecting feeds every %s", &timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}


	// feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	// if err != nil {
	// 	return fmt.Errorf("Error fetching feed: %w", err)
	// }

	// fmt.Println(feed.Channel.Title)
	// fmt.Println(feed.Channel.Link)
	// fmt.Println(feed.Channel.Description)
	// for _, item := range feed.Channel.Item {
	// 	fmt.Println(item.Title)
	// 	fmt.Println(item.Link)
	// 	fmt.Println(item.Description)
	// 	fmt.Println(item.PubDate)
	// 	fmt.Println()
	// 	fmt.Println()
	// }

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(), 
		CreatedAt: time.Now(), 
		UpdatedAt: time.Now(), 
		Name: feedName,
		Url: feedUrl,
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed: %w", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}

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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}

	urlArg := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), urlArg)
	if err != nil {
		return fmt.Errorf("Error getting feed by URL: %w", err)
	}

	ff, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}

	fmt.Printf("%s is now following %s\n", ff.UserName, ff.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Error retrieving followed feeds for %s: %w", user.Name, err)
	}

	for _, feed := range feeds {
		fmt.Println(feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <feed_url>", cmd.Name)
	}

	feedUrl := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("Error retrieving feed by url: %w", err)
	}

	s.db.DeleteFeedFollows(context.Background(), database.DeleteFeedFollowsParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})

	return nil
}

func scrapeFeeds(s *state) error {

	feedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error fetching next feed: %w", err)
	}

	_, err = s.db.MarkFeedFetched(context.Background(), feedToFetch.ID)
	if err != nil {
		return fmt.Errorf("Error marking feed fetched: %w", err)
	}

	rssFeed, err := fetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}

	// save the items to the posts database
	for _, item := range rssFeed.Channel.Item {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			fmt.Printf("Couldn't parse date %s: %v", item.PubDate, err)
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: item.Title,
			Url: item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: sql.NullTime{
				Time:  pubDate,
				Valid: !pubDate.IsZero(),
			},
			FeedID: feedToFetch.ID,	
		})
		if err != nil {
			fmt.Printf("failed to create post: %v", err)
		}
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.Args) == 1 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("limit must be a number: %w", err)
		}
		if n <= 0 {
			return fmt.Errorf("limmi must be greater than 0")
		}
		limit = n
	}
	if len(cmd.Args) > 1 {
		return fmt.Errorf("usage: %s [limit]", cmd.Name)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit),
	})
	if err != nil {
		return fmt.Errorf("Error retrieving posts for current user %s: %w", user.Name, err)
	}

	const (
		reset  = "\033[0m"
		bold   = "\033[1m"
		blue   = "\033[34m"
		gray   = "\033[90m"
	)

	for _, post := range posts {
		fmt.Printf("%s%s%s\n", bold, blue, post.Title)
		
		if post.PublishedAt.Valid {
			fmt.Printf("%sFrom: %s | %v%s\n", gray, post.Name, post.PublishedAt.Time.Format("2006-01-02"), reset)
		} else {
			fmt.Printf("%sFrom: %s%s\n", gray, post.Name, reset)
		}

		if post.Description.Valid {
			fmt.Printf("%s\n", post.Description.String)
		}

		fmt.Printf("%s-------------------------------------------%s\n\n", gray, reset)
	}

	return nil
}