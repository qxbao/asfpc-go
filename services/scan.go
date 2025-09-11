package services

import (
	"database/sql"
	"strconv"
	"strings"
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

			postId := strings.Split(*post.ID, "_")[1]

			queries.CreatePost(c.Request().Context(), db.CreatePostParams{
				PostID:    postId,
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

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	response := map[string]any{
		"posts_fetched":   len(*posts.Data),
		"posts_processed": processed,
		"data":            *posts.Data,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["error_count"] = len(errors)
	}

	return c.JSON(200, response)
}

func ScanPostComments(s infras.Server, c echo.Context) error {
	postId := c.Param("id")
	queries := s.Queries

	postIDInt, err := strconv.ParseInt(postId, 10, 32)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid Post ID"})
	}

	post, err := queries.GetPostByIdWithAccount(c.Request().Context(), int32(postIDInt))

	if err != nil {
		return c.JSON(404, map[string]string{"error": "Cannot find post: " + err.Error()})
	}

	fg := FacebookGraph{
		AccessToken: post.AccessToken.String,
	}

	comments, err := fg.GetPostComments(&post.PostID, &map[string]string{
		"limit": "50",
	})

	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to fetch post comments: " + err.Error()})
	}

	var wg sync.WaitGroup

	errChan := make(chan error, len(*comments.Data))
	successCount := make(chan int, len(*comments.Data))

	for _, comment := range *comments.Data {
		wg.Add(1)
		go func(comment infras.Comment) {
			defer wg.Done()

			if comment.ID == nil || comment.CreatedTime == nil {
				return
			}

			createdTime, err := time.Parse("2006-01-02T15:04:05-0700", *comment.CreatedTime)

			if err != nil {
				errChan <- err
				return
			}

			content := ""
			if comment.Message != nil {
				content = *comment.Message
			}
			author, err := queries.CreateProfile(c.Request().Context(), db.CreateProfileParams{
				Name: sql.NullString{
					String: *comment.From.Name,
					Valid:  true,
				},
				FacebookID: *comment.From.ID,
			})

			if err != nil {
				errChan <- err
				return
			}

			commentID := strings.Split(*comment.ID, "_")[1]

			queries.CreateComment(c.Request().Context(), db.CreateCommentParams{
				PostID:    post.ID,
				CommentID: commentID,
				Content:   content,
				CreatedAt: createdTime,
				AuthorID:  author.ID,
			})
			successCount <- 1
		}(comment)
	}

	wg.Wait()
	close(errChan)
	close(successCount)

	processed := 0
	for range successCount {
		processed++
	}

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	response := map[string]any{
		"comments_fetched":   len(*comments.Data),
		"comments_processed": processed,
		"data":               *comments.Data,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["error_count"] = len(errors)
	}

	return c.JSON(200, response)
}
