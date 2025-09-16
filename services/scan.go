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

func toNullString(ptr *string) sql.NullString {
	if ptr != nil {
		return sql.NullString{String: *ptr, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func extractEntityName(entity *infras.EntityNameID) sql.NullString {
	if entity != nil && entity.Name != nil {
		return sql.NullString{String: *entity.Name, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func joinWork(work *[]infras.Work) sql.NullString {
	if work == nil || len(*work) == 0 {
		return sql.NullString{Valid: false}
	}

	var workStrings []string
	for _, w := range *work {
		if w.Employer != nil && w.Employer.Name != nil {
			workStr := *w.Employer.Name
			if w.Position != nil && w.Position.Name != nil {
				workStr += " - " + *w.Position.Name
			}
			workStrings = append(workStrings, workStr)
		}
	}

	if len(workStrings) > 0 {
		return sql.NullString{String: strings.Join(workStrings, "; "), Valid: true}
	}
	return sql.NullString{Valid: false}
}

func joinEducation(education *[]infras.Education) sql.NullString {
	if education == nil || len(*education) == 0 {
		return sql.NullString{Valid: false}
	}

	var eduStrings []string
	for _, edu := range *education {
		if edu.School != nil && edu.School.Name != nil {
			eduStrings = append(eduStrings, *edu.School.Name)
		}
	}

	if len(eduStrings) > 0 {
		return sql.NullString{String: strings.Join(eduStrings, "; "), Valid: true}
	}
	return sql.NullString{Valid: false}
}

func getStringOrDefault(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

type ScanService struct {
	Server infras.Server
}

func (s ScanService) ScanGroupFeed(c echo.Context) error {
	groupId := c.Param("id")
	queries := s.Server.Queries

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
		"limit": s.Server.GetConfig(
			"facebook_group_feed_limit", "20",
		),
	})

	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to fetch group feed: " + err.Error()})
	}

	queries.UpdateGroupScannedAt(c.Request().Context(), group.ID)

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

func (s ScanService) ScanPostComments(c echo.Context) error {
	postId := c.Param("id")
	queries := s.Server.Queries

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
		"limit": s.Server.GetConfig(
			"facebook_post_comments_limit",
			"50",
		),
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
				FacebookID:  *comment.From.ID,
				ScrapedByID: post.AccountID,
			})

			if err != nil {
				errChan <- err
				return
			}

			queries.CreateComment(c.Request().Context(), db.CreateCommentParams{
				PostID:    post.ID,
				CommentID: *comment.ID,
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

func (s ScanService) ScanUserProfile(c echo.Context) error {
	userId := c.Param("id")
	userIdInt, err := strconv.ParseInt(userId, 10, 32)

	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid User ID"})
	}

	queries := s.Server.Queries
	userProfile, err := queries.GetProfileByIdWithAccount(c.Request().Context(), int32(userIdInt))

	if err != nil {
		return c.JSON(404, map[string]string{"error": "User profile doesn't exist: " + err.Error()})
	}

	fg := FacebookGraph{
		AccessToken: userProfile.AccessToken.String,
	}

	fetchedProfile, err := fg.GetUserDetails(userProfile.FacebookID, &map[string]string{})

	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to fetch user profile: " + err.Error()})
	}

	params := db.UpdateProfileAfterScanParams{
		ID:                 userProfile.ID,
		Bio:                toNullString(fetchedProfile.About),
		Email:              toNullString(fetchedProfile.Email),
		Location:           extractEntityName(fetchedProfile.Location),
		Hometown:           extractEntityName(fetchedProfile.Hometown),
		Birthday:           toNullString(fetchedProfile.Birthday),
		Gender:             toNullString(fetchedProfile.Gender),
		RelationshipStatus: toNullString(fetchedProfile.RelationshipStatus),
		Work:               joinWork(fetchedProfile.Work),
		Education:          joinEducation(fetchedProfile.Education),
		ProfileUrl:         getStringOrDefault(fetchedProfile.Link, ""),
		Locale:             getStringOrDefault(fetchedProfile.Locale, "en_US"),
		Phone:              sql.NullString{Valid: false},
	}

	_, err = queries.UpdateProfileAfterScan(c.Request().Context(), params)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to update profile: " + err.Error()})
	}

	response := map[string]any{
		"user_profile":    userProfile,
		"fetched_profile": fetchedProfile,
		"message":         "Profile updated successfully",
	}

	return c.JSON(200, response)
}
