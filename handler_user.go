package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Grumpster-Dev/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]
	// Attempt to get the user from the database
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	err = s.cfg.SetUser(name) // Note: no `:=` here because 'err' is already declared
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("User %s logged in successfully.\n", name)
	return nil

}

func handlerRegister(s *state, cmd command) error {

	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	name := cmd.Args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		return fmt.Errorf("user already exists")
	}
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("couldn't check existing users: %w", err)
	}

	user, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name:      name,
		},
	)
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("User %s registered successfully with ID %s.\n", user.Name, user.ID)
	return nil
}
func handlerReset(s *state, cmd command) error {

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't reset database: %w", err)
	}

	fmt.Println("Database has been reset to its initial state.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Println("* ", user.Name, "(current)")
		} else {
			fmt.Println("* ", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	// Placeholder for future implementation
	ctx := context.Background()
	feed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}
	fmt.Printf("Feed Title: %s\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <feed_name> <feed_url>", cmd.Name)
	}
	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]

	feed, err := s.db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name:      feedName,
			Url:       feedURL,
			UserID:    user.ID, // Replace with actual user ID
		},
	)
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	fmt.Printf("Feed %s created successfully with ID %s.\n", feed.Name, feed.ID)

	follow, err := s.db.CreateFeedFollows(
		context.Background(),
		database.CreateFeedFollowsParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		},
	)

	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w", err)
	}
	fmt.Printf("User %s is now following feed %s.\n", follow.UserName, follow.FeedName)
	return nil

}

func handlerFeeds(s *state, cmd command) error {
	// Placeholder for future implementation

	listfeeds, err := s.db.GetFeedsByUserName(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}
	for _, feed := range listfeeds {
		fmt.Printf("* %s (%s) by %s\n", feed.Name, feed.Url, feed.UserName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	url := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't find feed by URL: %w", err)
	}

	follow, err := s.db.CreateFeedFollows(
		context.Background(),
		database.CreateFeedFollowsParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		},
	)

	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w", err)
	}

	fmt.Printf("User %s is now following feed %s.\n", follow.UserName, follow.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feed follows: %w", err)
	}
	for _, follow := range follows {
		fmt.Printf("* %s \n", follow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	feedURL := cmd.Args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("couldn't find feed by URL: %w", err)
	}
	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("couldn't find current user: %w", err)
		}
		return handler(s, cmd, user)
	}
}
