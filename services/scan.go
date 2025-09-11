package services

import (
	"strconv"
	"sync"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

func ScanGroupFeed(s infras.Server, c echo.Context) error {
	groupId := c.Param("id")
	queries := s.Queries

	if groupId == "" {
		return c.JSON(400, map[string]string{"error": "Group ID is required"})
	}

	groupIDInt, err := strconv.ParseInt(groupId, 10, 32)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid Group ID"})
	}

	group, err := queries.GetGroupByIdWithAccount(c.Request().Context(), int32(groupIDInt))

	if err != nil {
		return c.JSON(404, map[string]string{"error": "Cannot find group: " + err.Error()})
	}

	fg := FacebookGraph{
		AccessToken: group.AccessToken.String,
	}

	posts, err := fg.GetGroupFeed(&group.GroupID, &map[string]string{
		// TODO: Should be replaced with the config later
		"limit": "20",
	})

	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to fetch group feed: " + err.Error()})
	}

	var wg sync.WaitGroup

	errChan := make(chan error, len(*posts.Data))
	successCount := make(chan int, len(*posts.Data))

	for _, post := range *posts.Data {
		wg.Add(1)
		go func(post infras.Post) {
			defer wg.Done()

			if post.ID == nil || post.UpdatedTime == nil {
				return
			}

			updatedTime, err := time.Parse("2006-01-02T15:04:05-0700", *post.UpdatedTime)
			if err != nil {
				errChan <- err
				return
			}

			content := ""
			if post.Message != nil {
				content = *post.Message
			}

			queries.CreatePost(c.Request().Context(), db.CreatePostParams{
				PostID:    *post.ID,
				Content:   content,
				CreatedAt: updatedTime,
				GroupID:   group.ID,
			})
			successCount <- 1
		}(post)
	}

	wg.Wait()
	close(errChan)
	close(successCount)

	processed := 0
	for range successCount {
		processed++
	}

	return c.JSON(200, map[string]any{
		// Posts that have been added to the database
		"posts_fetched":   len(*posts.Data),
		// Posts that have been fetched from Facebook
		"posts_processed": processed,
		"data":            *posts.Data,
	})
}
