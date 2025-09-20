package services

// TODO: GetPostsToScan should be remove later, also is_analyzed in post table

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/async"
	lg "github.com/qxbao/asfpc/pkg/logger"
	"github.com/qxbao/asfpc/pkg/utils"
	"go.uber.org/zap"
)

type ScanService struct {
	Server infras.Server
}

type GroupScanError struct {
	AccountID int32
	Error     []error
}

type PostScanResult struct {
	GID     int32
	Total   int32
	Success int32
}

type processGroupInput struct {
	Context    context.Context
	AccountID  int32
	GroupLimit int32
}

type processPostsInput struct {
	ScraperId int32
	Context   context.Context
	Group     *db.GetGroupsToScanRow
}

type processPostInput struct {
	ScraperId int32
	Context   context.Context
	Post      infras.Post
	GroupID   int32
}

type processCommentInput struct {
	Context   context.Context
	ScraperId int32
	Comment   infras.PostComment
	PostID    int32
}

var loggerName string = "ScanningService"
var logger *zap.SugaredLogger = lg.GetLogger(&loggerName)

type GroupScanSuccess struct {
	AccountID int32
	Result    []PostScanResult
}

func (s ScanService) ScanAllGroups() {
	logger.Info("Starting cron task [ScanAllGroups]")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	ids, err := s.Server.Queries.GetOKAccountIds(
		ctx,
	)

	if err != nil {
		logger.Errorf("Error fetching account IDs: %v", err)
		return
	}

	groupLimit, _ := strconv.ParseInt(s.Server.GetConfig("FACEBOOK_GROUP_LIMIT", "5"), 10, 32)
	mainConcurrency, _ := strconv.ParseInt(s.Server.GetConfig("SCAN_MAIN_CONCURRENCY", "2"), 10, 32)

	mainSemaphore := async.GetSemaphore[processGroupInput, GroupScanSuccess](int(mainConcurrency))

	for _, id := range ids {
		mainSemaphore.Assign(s.processGroups, processGroupInput{
			Context:    ctx,
			AccountID:  id,
			GroupLimit: int32(groupLimit),
		})
	}

	succs, errs := mainSemaphore.Run()

	for id, err := range errs {
		if err == nil {
			continue
		}
		s.Server.Queries.LogAction(ctx, db.LogActionParams{
			AccountID:   sql.NullInt32{Int32: ids[id], Valid: true},
			Action:      "scan_group",
			TargetID:    sql.NullInt32{Valid: false},
			Description: sql.NullString{String: err.Error(), Valid: true},
		})
		logger.Errorf("Account %d scan group failed: %s", ids[id], err.Error())
	}

	for _, success := range succs {
		for _, res := range success.Result {
			s.Server.Queries.LogAction(ctx, db.LogActionParams{
				AccountID:   sql.NullInt32{Int32: success.AccountID, Valid: true},
				Action:      "scan_group",
				TargetID:    sql.NullInt32{Int32: res.GID, Valid: true},
				Description: sql.NullString{String: fmt.Sprintf("Scanned group ID %d: %d/%d posts", res.GID, res.Success, res.Total), Valid: true},
			})
		}
		logger.Infof("Account %d scan group successfully: %d group(s)", success.AccountID, len(success.Result))
	}
	logger.Info("ScanAllGroups task completed.")
}

func (s ScanService) processGroups(input processGroupInput) GroupScanSuccess {
	groups, err := s.Server.Queries.GetGroupsToScan(input.Context, db.GetGroupsToScanParams{
		AccountID: sql.NullInt32{
			Int32: input.AccountID,
			Valid: true,
		},
		Limit: int32(input.GroupLimit),
	})
	logger.Infof("Account %d: Fetched %d groups to scan", input.AccountID, len(groups))
	if err != nil {
		panic(fmt.Errorf("failed to fetch groups to scan for account %d: %v", input.AccountID, err))
	}
	postsConcurrency, _ := strconv.ParseInt(s.Server.GetConfig("SCAN_POSTS_CONCURRENCY", "5"), 10, 32)
	semaphore := async.GetSemaphore[processPostsInput, PostScanResult](int(postsConcurrency))
	for _, group := range groups {
		semaphore.Assign(s.processPosts, processPostsInput{
			ScraperId: input.AccountID,
			Context:   input.Context,
			Group:     &group,
		})
	}
	success, errs := semaphore.Run()

	for _, err := range errs {
		if err != nil {
			logger.Errorf("Account %d: Error scanning posts: %v", input.AccountID, err)
			s.Server.Queries.LogAction(input.Context, db.LogActionParams{
				AccountID:   sql.NullInt32{Int32: input.AccountID, Valid: true},
				Action:      "scan_group",
				TargetID:    sql.NullInt32{Valid: false},
				Description: sql.NullString{String: fmt.Sprintf("Error scanning posts: %v", err), Valid: true},
			})
		}
	}
	return GroupScanSuccess{
		AccountID: input.AccountID,
		Result:    success,
	}
}

func (s ScanService) processPosts(input processPostsInput) PostScanResult {
	logger.Infof("Scanning posts for group %s (ID: %d)", input.Group.GroupName, input.Group.ID)
	feedLimit, _ := strconv.ParseInt(s.Server.GetConfig("FACEBOOK_GROUP_FEED_LIMIT", "10"), 10, 32)
	fg := FacebookGraph{
		AccessToken: input.Group.AccessToken.String,
	}

	posts, err := fg.GetGroupFeed(&input.Group.GroupID, &map[string]string{
		"limit": fmt.Sprintf("%d", feedLimit),
		"order": "chronological",
	})

	if err != nil {
		panic(fmt.Errorf("failed to fetch group feed: %s", err.Error()))
	}

	if posts.Data == nil {
		logger.Infof("No posts data returned for group %d", input.Group.ID)
		s.Server.Queries.UpdateGroupScannedAt(input.Context, input.Group.ID)
		return PostScanResult{
			GID:     input.Group.ID,
			Total:   0,
			Success: 0,
		}
	}

	s.Server.Queries.UpdateGroupScannedAt(input.Context, input.Group.ID)
	postConcurrency, _ := strconv.ParseInt(s.Server.GetConfig("SCAN_POST_CONCURRENCY", "5"), 10, 32)
	semaphore := async.GetSemaphore[processPostInput, bool](int(postConcurrency))

	for _, post := range *posts.Data {
		semaphore.Assign(s.processPost, processPostInput{
			ScraperId: input.ScraperId,
			Context:   input.Context,
			Post:      post,
			GroupID:   input.Group.ID,
		})
	}

	_, errs := semaphore.Run()
	successCount := int32(0)
	for _, err := range errs {
		if err == nil {
			successCount++
		}
	}
	return PostScanResult{
		GID:     input.Group.ID,
		Total:   int32(len(*posts.Data)),
		Success: successCount,
	}
}

func (s ScanService) processPost(input processPostInput) bool {
	logger.Infof("Processing post (ID: %d)", input.Post.ID)

	if input.Post.ID == nil || input.Post.UpdatedTime == nil {
		panic(fmt.Errorf("post ID or UpdatedTime is nil for post %s", *input.Post.ID))
	}

	if input.Post.Comments.Count == nil || *input.Post.Comments.Count == 0 {
		return false
	}

	updatedTime, err := time.Parse("2006-01-02T15:04:05-0700", *input.Post.UpdatedTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse UpdatedTime for post %s: %v", *input.Post.ID, err))
	}

	content := ""
	if input.Post.Message != nil {
		content = *input.Post.Message
	}

	postId := strings.Split(*input.Post.ID, "_")[1]

	p, _ := s.Server.Queries.CreatePost(input.Context, db.CreatePostParams{
		PostID:    postId,
		Content:   content,
		CreatedAt: updatedTime,
		GroupID:   input.GroupID,
	})

	if input.Post.Comments != nil && input.Post.Comments.Data != nil {
		commentConcurrency, _ := strconv.ParseInt(s.Server.GetConfig("SCAN_COMMENT_CONCURRENCY", "5"), 10, 32)
		semaphore := async.GetSemaphore[processCommentInput, bool](int(commentConcurrency))
		commentsLimit, _ := strconv.ParseInt(s.Server.GetConfig("FACEBOOK_COMMENTS_LIMIT", "15"), 10, 32)
		for i, comment := range *input.Post.Comments.Data {
			if i >= int(commentsLimit) {
				break
			}
			if comment.From == nil || comment.From.ID == nil {
				continue
			}
			semaphore.Assign(s.processComment, processCommentInput{
				ScraperId: input.ScraperId,
				Context:   input.Context,
				Comment:   comment,
				PostID:    p.ID,
			})
		}
		_, errs := semaphore.Run()
		for _, err := range errs {
			if err != nil {
				logger.Errorf("Failed to process comment for post %s: %v", *input.Post.ID, err)
				s.Server.Queries.LogAction(input.Context, db.LogActionParams{
					AccountID:   sql.NullInt32{Int32: input.ScraperId, Valid: true},
					Action:      "scan_comment",
					TargetID:    sql.NullInt32{Int32: p.ID, Valid: true},
					Description: sql.NullString{String: fmt.Sprintf("Failed to process comment for post %s: %v", *input.Post.ID, err), Valid: true},
				})
			}
		}
	}
	return true
}

func (s ScanService) processComment(input processCommentInput) bool {
	if input.Comment.From == nil || input.Comment.From.ID == nil {
		return false
	}
	logger.Debugf("Processing comment %s from post %d", *input.Comment.ID, input.PostID)

	if input.Comment.ID == nil || len(*input.Comment.ID) > 15 {
		panic(fmt.Errorf("anonymous comment or invalid comment ID: %v", input.Comment.ID))
	}
	
	profile, err := s.Server.Queries.CreateProfile(input.Context, db.CreateProfileParams{
		FacebookID:  input.Comment.From.ID.String(),
		Name:        utils.ToNullString(input.Comment.From.Name),
		ScrapedByID: input.ScraperId,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create profile for comment author %s: %v", input.Comment.From.ID.String(), err))
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05-0700", *input.Comment.CreatedTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse CreatedTime for comment %s: %v", *input.Comment.ID, err))
	}

	commentID := strings.Split(*input.Comment.ID, "_")
	if len(commentID) < 2 {
		panic(fmt.Errorf("invalid comment ID format: %s", *input.Comment.ID))
	}

	_, err = s.Server.Queries.CreateComment(input.Context, db.CreateCommentParams{
		CommentID: commentID[len(commentID)-1],
		PostID:    input.PostID,
		Content:   utils.GetStringOrDefault(input.Comment.Message, ""),
		AuthorID:  profile.ID,
		CreatedAt: parsedTime,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create comment %s: %v", *input.Comment.ID, err))
	}
	return true
}

type processProfileInput struct {
	Context context.Context
	Profile db.GetProfilesToScanRow
}

func (s ScanService) ScanAllProfiles() {
	logger.Info("Starting cron task [ScanAllProfiles]")
	ctx, ctx_cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer ctx_cancel()

	profileLimit, _ := strconv.ParseInt(s.Server.GetConfig("FACEBOOK_PROFILE_SCAN_LIMIT", "80"), 10, 32)
	profiles, err := s.Server.Queries.GetProfilesToScan(ctx, int32(profileLimit))

	if err != nil {
		logger.Errorf("Failed to fetch profiles to scan: %s", err.Error())
		return
	}
	logger.Infof("Fetched %d profiles to scan", len(profiles))

	profileConcurrency, _ := strconv.ParseInt(s.Server.GetConfig("SCAN_PROFILE_CONCURRENCY", "5"), 10, 32)
	semaphore := async.GetSemaphore[processProfileInput, bool](int(profileConcurrency))

	for _, profile := range profiles {
		semaphore.Assign(s.processProfileWithSemaphore, processProfileInput{
			Context: ctx,
			Profile: profile,
		})
	}

	results, errs := semaphore.Run()

	var successCount int = 0
	for i, err := range errs {
		profile := profiles[i]
		if err != nil {
			s.Server.Queries.LogAction(ctx, db.LogActionParams{
				AccountID:   sql.NullInt32{Int32: profile.AccountID, Valid: true},
				Action:      "scan_profile",
				TargetID:    sql.NullInt32{Int32: profile.ID, Valid: true},
				Description: sql.NullString{String: fmt.Sprintf("scan profile %d failed: %s", profile.ID, err.Error()), Valid: true},
			})
			logger.Errorf("Failed to process profile %d: %s", profile.ID, err.Error())
		} else if results[i] {
			s.Server.Queries.LogAction(ctx, db.LogActionParams{
				AccountID:   sql.NullInt32{Int32: profile.AccountID, Valid: true},
				Action:      "scan_profile",
				TargetID:    sql.NullInt32{Int32: profile.ID, Valid: true},
				Description: sql.NullString{String: fmt.Sprintf("scanned profile %d successfully", profile.ID), Valid: true},
			})
			successCount++
		}
	}

	logger.Infof("ScanAllProfiles task completed: %d/%d", successCount, len(profiles))
}

func (s ScanService) processProfileWithSemaphore(input processProfileInput) bool {
	err := s.processProfile(input.Context, input.Profile)
	if err != nil {
		panic(err)
	}
	return true
}

func (s ScanService) processProfile(ctx context.Context, profile db.GetProfilesToScanRow) error {
	fg := FacebookGraph{
		AccessToken: profile.AccessToken.String,
	}
	fetchedProfile, err := fg.GetUserDetails(profile.FacebookID, &map[string]string{})

	if err != nil {
		s.Server.Queries.UpdateProfileScanStatus(ctx, profile.ID)
		return fmt.Errorf("failed to fetch user profile: %s", err.Error())
	}

	params := db.UpdateProfileAfterScanParams{
		ID:                 profile.ID,
		Bio:                utils.ToNullString(fetchedProfile.About),
		Email:              utils.ToNullString(fetchedProfile.Email),
		Location:           utils.ExtractEntityName(fetchedProfile.Location),
		Hometown:           utils.ExtractEntityName(fetchedProfile.Hometown),
		Birthday:           utils.ToNullString(fetchedProfile.Birthday),
		Gender:             utils.ToNullString(fetchedProfile.Gender),
		RelationshipStatus: utils.ToNullString(fetchedProfile.RelationshipStatus),
		Work:               utils.JoinWork(fetchedProfile.Work),
		Education:          utils.JoinEducation(fetchedProfile.Education),
		ProfileUrl:         utils.GetStringOrDefault(fetchedProfile.Link, ""),
		Locale:             utils.GetStringOrDefault(fetchedProfile.Locale, "en_US"),
		Phone:              sql.NullString{Valid: false},
	}

	_, err = s.Server.Queries.UpdateProfileAfterScan(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update profile: %s", err.Error())
	}
	return nil
}
