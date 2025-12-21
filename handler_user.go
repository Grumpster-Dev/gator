package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

func handlerListUsers(s *state, cmd command) error {
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
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <duration>", cmd.Name)
	}

	durationStr := cmd.Args[0]

	timeBetweenRequests, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)
	fmt.Printf("Collecting feeds every %v \n", timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
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

	follow, err := s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
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

	follow, err := s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
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

	// Get all follows for this user
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("couldn't get feed follows: %w", err)
	}

	// Find the follow row for this feed
	var followToDelete database.GetFeedFollowsForUserRow
	found := false
	for _, f := range follows {
		if f.FeedID == feed.ID {
			followToDelete = f
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("user is not following that feed")
	}

	// Use the row's ID and CreatedAt, as required by DeleteFeedFollow
	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		ID:        followToDelete.ID,
		CreatedAt: followToDelete.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
	}

	fmt.Printf("User %s unfollowed feed %s.\n", user.Name, feed.Name)
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

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("failed to get next feed to fetch: %v\n", err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("failed to mark feed as fetched: %v\n", err)
		return
	}

	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("failed to fetch feed data: %v\n", err)
		return
	}
	for _, item := range feedData.Channel.Item {
		//fmt.Println(item.Title)
		//s.db.CreatePost(context.Background(),item.Title,item)
		var publishedAt sql.NullTime
		parsedTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err == nil {
			// Parsing succeeded!
			publishedAt = sql.NullTime{Time: parsedTime, Valid: true}
		} else {
			// Parsing failed, no date available
			publishedAt = sql.NullTime{Valid: false}
		}
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}
		_, err = s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"posts_url_key\"" {
				continue
			} else {
				fmt.Printf("failed to create post: %v\n", err)
			}
		} else {
			fmt.Printf("DEBUG: created post %s for feed %s\n", item.Title, feed.ID)
		}
	}
}

func handlerBrowsePosts(s *state, cmd command, user database.User) error {

	var postLimit int32 = 2
	if len(cmd.Args) == 1 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid post limit: %w", err)
		}
		postLimit = int32(n)
	}

	postParams := database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  postLimit,
	}

	posts, err := s.db.GetPostsByUser(context.Background(), postParams)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	fmt.Printf("DEBUG: got %d posts for user %s\n", len(posts), user.Name)

	for _, post := range posts {
		fmt.Printf("* %s (%s) from feed ID %s\n", post.Title, post.Url, post.FeedID.String())
	}
	return nil
}
