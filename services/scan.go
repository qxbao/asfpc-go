package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
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

type ScanPostResult struct {
	GID     int32
	Total   int32
	Success int32
}

var loggerName string = "ScanningService"
var logger *zap.SugaredLogger = lg.GetLogger(&loggerName)

type GroupScanSuccess struct {
	AccountID int32
	Result    []ScanPostResult
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
	wg := sync.WaitGroup{}
	errChannel := make(chan GroupScanError, len(ids))
	successChannel := make(chan GroupScanSuccess, len(ids))

	for _, id := range ids {
		wg.Add(1)
		go func(accountId int32) {
			defer wg.Done()
			groups, err := s.Server.Queries.GetGroupsToScan(ctx, db.GetGroupsToScanParams{
				AccountID: sql.NullInt32{
					Int32: accountId,
					Valid: true,
				},
				Limit: int32(groupLimit),
			})
			logger.Infof("Account %d: Fetched %d groups to scan", accountId, len(groups))
			if err != nil {
				errChannel <- GroupScanError{
					AccountID: accountId,
					Error:     []error{fmt.Errorf("failed to fetch group for account %d: %v", accountId, err)},
				}
				return
			}
			success := []ScanPostResult{}
			ers := []error{}
			for _, group := range groups {
				result, err := s.scanPosts(ctx, &group, ers)
				if err != nil {
					ers = append(ers, fmt.Errorf("failed to scan group %s: %v", group.GroupID, err))
				} else {
					success = append(success, *result)
				}
			}
			errChannel <- GroupScanError{
				AccountID: accountId,
				Error:     ers,
			}
			successChannel <- GroupScanSuccess{
				AccountID: accountId,
				Result:    success,
			}
		}(id)
	}
	wg.Wait()
	close(errChannel)
	close(successChannel)

	for err := range errChannel {
		for _, er := range err.Error {
			s.Server.Queries.LogAction(ctx, db.LogActionParams{
				AccountID:   sql.NullInt32{Int32: err.AccountID, Valid: true},
				Action:      "scan_group",
				TargetID:    sql.NullInt32{Valid: false},
				Description: sql.NullString{String: er.Error(), Valid: true},
			})
			logger.Errorf("Account %d scan group failed: %s", err.AccountID, er.Error())
		}
	}

	for success := range successChannel {
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

func (s ScanService) scanPosts(ctx context.Context, group *db.GetGroupsToScanRow, ers []error) (*ScanPostResult, error) {
	logger.Infof("Scanning posts for group %s (ID: %d)", group.GroupName, group.ID)
	feedLimit, _ := strconv.ParseInt(s.Server.GetConfig("FACEBOOK_GROUP_FEED_LIMIT", "10"), 10, 32)
	fg := FacebookGraph{
		AccessToken: group.AccessToken.String,
	}
	posts, err := fg.GetGroupFeed(&group.GroupID, &map[string]string{
		"limit": fmt.Sprintf("%d", feedLimit),
		"order": "chronological",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch group feed: %s", err.Error())
	}
	logger.Infof("Fetched %d posts from group %d", len(*posts.Data), group.ID)

	s.Server.Queries.UpdateGroupScannedAt(ctx, group.ID)
	wg := sync.WaitGroup{}
	successCountChannel := make(chan int32)
	successCount := int32(0)
	for _, post := range *posts.Data {
		wg.Add(1)
		go func(post infras.Post) {
			defer wg.Done()
			if post.ID == nil || post.UpdatedTime == nil {
				ers = append(ers, fmt.Errorf("post ID or UpdatedTime is nil for group %s", group.GroupID))
				return
			}
			if post.Comments.Count == nil || *post.Comments.Count == 0 {
				return
			}

			updatedTime, err := time.Parse("2006-01-02T15:04:05-0700", *post.UpdatedTime)
			if err != nil {
				ers = append(ers, fmt.Errorf("failed to parse UpdatedTime for post %s: %v", *post.ID, err))
				return
			}

			content := ""
			if post.Message != nil {
				content = *post.Message
			}

			postId := strings.Split(*post.ID, "_")[1]

			p, _ := s.Server.Queries.CreatePost(ctx, db.CreatePostParams{
				PostID:    postId,
				Content:   content,
				CreatedAt: updatedTime,
				GroupID:   group.ID,
			})
			successCountChannel <- 1

			if post.Comments != nil && post.Comments.Data != nil {
				commentsLimit, _ := strconv.ParseInt(s.Server.GetConfig("FACEBOOK_COMMENTS_LIMIT", "5"), 10, 32)
				for i, comment := range *post.Comments.Data {
					if i >= int(commentsLimit) {
						break
					}
					if comment.From == nil || comment.From.ID == nil {
						continue
					}
					profile, err := s.Server.Queries.CreateProfile(ctx, db.CreateProfileParams{
						FacebookID:  comment.From.ID.String(),
						Name:        utils.ToNullString(comment.From.Name),
						ScrapedByID: group.AccountID.Int32,
					})
					if err != nil {
						ers = append(ers, fmt.Errorf("failed to create profile for comment author %s: %v", comment.From.ID.String(), err))
						continue
					}
					parsedTime, err := time.Parse("2006-01-02T15:04:05-0700", *comment.CreatedTime)
					if err != nil {
						ers = append(ers, fmt.Errorf("failed to parse CreatedTime for comment %s: %v", *comment.ID, err))
						continue
					}
					_, err = s.Server.Queries.CreateComment(ctx, db.CreateCommentParams{
						CommentID: *comment.ID,
						PostID:    p.ID,
						Content:   utils.GetStringOrDefault(comment.Message, ""),
						AuthorID:  profile.ID,
						CreatedAt: parsedTime,
					})
					if err != nil {
						ers = append(ers, fmt.Errorf("failed to create comment %s: %v", *comment.ID, err))
						continue
					}
				}
			}
		}(post)
	}

	go func() {
		wg.Wait()
		close(successCountChannel)
	}()

	successCount = s.countSuccesses(successCountChannel, 10*time.Second)

	return &ScanPostResult{
		GID:     group.ID,
		Total:   int32(len(*posts.Data)),
		Success: successCount,
	}, nil
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

	var successCount int = 0
	wg := sync.WaitGroup{}
	for _, profile := range profiles {
		wg.Add(1)
		go func(profile db.GetProfilesToScanRow) {
			defer wg.Done()
			err := s.processProfile(ctx, profile)
			if err != nil {
				s.Server.Queries.LogAction(ctx, db.LogActionParams{
					AccountID:   sql.NullInt32{Int32: profile.AccountID, Valid: true},
					Action:      "scan_profile",
					TargetID:    sql.NullInt32{Int32: profile.ID, Valid: true},
					Description: sql.NullString{String: fmt.Sprintf("scan profile %d failed: %s", profile.ID, err.Error()), Valid: true},
				})
				logger.Errorf("Failed to process profile %d: %s", profile.ID, err.Error())
			} else {
				s.Server.Queries.LogAction(ctx, db.LogActionParams{
					AccountID:   sql.NullInt32{Int32: profile.AccountID, Valid: true},
					Action:      "scan_profile",
					TargetID:    sql.NullInt32{Int32: profile.ID, Valid: true},
					Description: sql.NullString{String: fmt.Sprintf("scanned profile %d successfully", profile.ID), Valid: true},
				})
				successCount++
			}
		}(profile)
	}
	wg.Wait()
	logger.Infof("ScanAllProfiles task completed: %d/%d", successCount, len(profiles))
}

func (s ScanService) processProfile(ctx context.Context, profile db.GetProfilesToScanRow) error {
	fg := FacebookGraph{
		AccessToken: profile.AccessToken.String,
	}
	fetchedProfile, err := fg.GetUserDetails(profile.FacebookID, &map[string]string{})

	if err != nil {
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

func (s ScanService) countSuccesses(ch <-chan int32, timeout time.Duration) int32 {
	var count int32
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return count
			}
			count++
		case <-timer.C:
			return count
		}
	}
}
