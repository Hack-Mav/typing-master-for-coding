package database

import (
	"context"
	"errors"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// Helper functions for database operations
func NewQuery(kind string) *datastore.Query {
	return datastore.NewQuery(kind)
}

func NameKey(kind, name string) *datastore.Key {
	return datastore.NameKey(kind, name, nil)
}

type Query = datastore.Query
type Key = datastore.Key

var ErrNotFound = errors.New("entity not found")

// Tutorial operations
func GetAllTutorials(db *DatastoreClient) ([]models.Tutorial, error) {
	if db.IsMock {
		return db.Mock.GetAllTutorials()
	}

	ctx := context.Background()
	query := NewQuery("Tutorial").Filter("IsActive =", true).Order("CreatedAt")

	var tutorials []models.Tutorial
	keys, err := db.Client.GetAll(ctx, query, &tutorials)
	if err != nil {
		return nil, err
	}

	for i := range tutorials {
		tutorials[i].ID = keys[i].Name
	}

	return tutorials, nil
}

func GetTutorialsByCategory(db *DatastoreClient, category string) ([]models.Tutorial, error) {
	if db.IsMock {
		return db.Mock.GetTutorialsByCategory(category)
	}

	ctx := context.Background()
	query := NewQuery("Tutorial").Filter("IsActive =", true).Filter("Category =", category).Order("CreatedAt")

	var tutorials []models.Tutorial
	keys, err := db.Client.GetAll(ctx, query, &tutorials)
	if err != nil {
		return nil, err
	}

	for i := range tutorials {
		tutorials[i].ID = keys[i].Name
	}

	return tutorials, nil
}

func GetTutorialsByDifficulty(db *DatastoreClient, difficulty string) ([]models.Tutorial, error) {
	if db.IsMock {
		return db.Mock.GetTutorialsByDifficulty(difficulty)
	}

	ctx := context.Background()
	query := NewQuery("Tutorial").Filter("IsActive =", true).Filter("Difficulty =", difficulty).Order("CreatedAt")

	var tutorials []models.Tutorial
	keys, err := db.Client.GetAll(ctx, query, &tutorials)
	if err != nil {
		return nil, err
	}

	for i := range tutorials {
		tutorials[i].ID = keys[i].Name
	}

	return tutorials, nil
}

func GetTutorialByID(db *DatastoreClient, tutorialID string) (*models.Tutorial, error) {
	if db.IsMock {
		return db.Mock.GetTutorialByID(tutorialID)
	}

	ctx := context.Background()
	key := NameKey("Tutorial", tutorialID)

	var tutorial models.Tutorial
	err := db.Client.Get(ctx, key, &tutorial)
	if err != nil {
		return nil, err
	}

	tutorial.ID = tutorialID
	return &tutorial, nil
}

// User tutorial progress operations
func GetUserTutorialProgress(db *DatastoreClient, userID, tutorialID string) (*models.UserTutorialProgress, error) {
	ctx := context.Background()
	query := NewQuery("UserTutorialProgress").Filter("UserID =", userID).Filter("TutorialID =", tutorialID)

	var progress []models.UserTutorialProgress
	keys, err := db.Client.GetAll(ctx, query, &progress)
	if err != nil {
		return nil, err
	}

	if len(progress) == 0 {
		return nil, ErrNotFound
	}

	progress[0].ID = keys[0].Name
	return &progress[0], nil
}

func GetUserAllTutorialProgress(db *DatastoreClient, userID string) ([]models.UserTutorialProgress, error) {
	ctx := context.Background()
	query := NewQuery("UserTutorialProgress").Filter("UserID =", userID).Order("-LastAccessedAt")

	var progress []models.UserTutorialProgress
	keys, err := db.Client.GetAll(ctx, query, &progress)
	if err != nil {
		return nil, err
	}

	for i := range progress {
		progress[i].ID = keys[i].Name
	}

	return progress, nil
}

func CreateUserTutorialProgress(db *DatastoreClient, progress *models.UserTutorialProgress) error {
	ctx := context.Background()
	key := NameKey("UserTutorialProgress", progress.ID)

	_, err := db.Client.Put(ctx, key, progress)
	return err
}

func UpdateUserTutorialProgress(db *DatastoreClient, progress *models.UserTutorialProgress) error {
	ctx := context.Background()
	key := NameKey("UserTutorialProgress", progress.ID)

	_, err := db.Client.Put(ctx, key, progress)
	return err
}

// Tooltip operations
func GetTooltipsByContext(db *DatastoreClient, pageContext string) ([]models.Tooltip, error) {
	ctx := context.Background()
	query := NewQuery("Tooltip").Filter("IsActive =", true).Filter("PageContext =", pageContext).Order("CreatedAt")

	var tooltips []models.Tooltip
	keys, err := db.Client.GetAll(ctx, query, &tooltips)
	if err != nil {
		return nil, err
	}

	for i := range tooltips {
		tooltips[i].ID = keys[i].Name
	}

	return tooltips, nil
}

// Help article operations
func SearchHelpArticles(db *DatastoreClient, query, category string, limit int) ([]models.HelpArticle, error) {
	ctx := context.Background()

	var dsQuery *Query
	if category != "" {
		dsQuery = NewQuery("HelpArticle").Filter("IsActive =", true).Filter("Category =", category)
	} else {
		dsQuery = NewQuery("HelpArticle").Filter("IsActive =", true)
	}

	if query != "" {
		// Simple text search - in a real implementation, this would use a search index
		dsQuery = dsQuery.Filter("Title =", query)
	}

	dsQuery = dsQuery.Order("-SearchRank").Limit(limit)

	var articles []models.HelpArticle
	keys, err := db.Client.GetAll(ctx, dsQuery, &articles)
	if err != nil {
		return nil, err
	}

	for i := range articles {
		articles[i].ID = keys[i].Name
	}

	return articles, nil
}

func GetHelpArticleByID(db *DatastoreClient, articleID string) (*models.HelpArticle, error) {
	ctx := context.Background()
	key := NameKey("HelpArticle", articleID)

	var article models.HelpArticle
	err := db.Client.Get(ctx, key, &article)
	if err != nil {
		return nil, err
	}

	article.ID = articleID
	return &article, nil
}

func UpdateHelpArticle(db *DatastoreClient, article *models.HelpArticle) error {
	ctx := context.Background()
	key := NameKey("HelpArticle", article.ID)

	_, err := db.Client.Put(ctx, key, article)
	return err
}

// FAQ operations
func GetFAQsByCategory(db *DatastoreClient, category string) ([]models.FAQ, error) {
	ctx := context.Background()
	var query *Query

	if category != "" {
		query = NewQuery("FAQ").Filter("IsActive =", true).Filter("Category =", category).Order("Order")
	} else {
		query = NewQuery("FAQ").Filter("IsActive =", true).Order("Order")
	}

	var faqs []models.FAQ
	keys, err := db.Client.GetAll(ctx, query, &faqs)
	if err != nil {
		return nil, err
	}

	for i := range faqs {
		faqs[i].ID = keys[i].Name
	}

	return faqs, nil
}

// User onboarding state operations
func GetUserOnboardingState(db *DatastoreClient, userID string) (*models.UserOnboardingState, error) {
	ctx := context.Background()
	query := NewQuery("UserOnboardingState").Filter("UserID =", userID)

	var states []models.UserOnboardingState
	keys, err := db.Client.GetAll(ctx, query, &states)
	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, ErrNotFound
	}

	states[0].ID = keys[0].Name
	return &states[0], nil
}

func UpdateUserOnboardingState(db *DatastoreClient, state *models.UserOnboardingState) error {
	ctx := context.Background()
	key := NameKey("UserOnboardingState", state.ID)

	_, err := db.Client.Put(ctx, key, state)
	return err
}

// Feedback operations
func CreateFeedback(db *DatastoreClient, feedback *models.Feedback) error {
	ctx := context.Background()
	key := NameKey("Feedback", feedback.ID)

	_, err := db.Client.Put(ctx, key, feedback)
	return err
}

func GetFeedback(db *DatastoreClient, status, category, type_ string) ([]models.Feedback, error) {
	ctx := context.Background()
	query := NewQuery("Feedback").Order("-CreatedAt")

	if status != "" {
		query = query.Filter("Status =", status)
	}
	if category != "" {
		query = query.Filter("Category =", category)
	}
	if type_ != "" {
		query = query.Filter("Type =", type_)
	}

	var feedbacks []models.Feedback
	keys, err := db.Client.GetAll(ctx, query, &feedbacks)
	if err != nil {
		return nil, err
	}

	for i := range feedbacks {
		feedbacks[i].ID = keys[i].Name
	}

	return feedbacks, nil
}

func GetFeedbackByID(db *DatastoreClient, feedbackID string) (*models.Feedback, error) {
	ctx := context.Background()
	key := NameKey("Feedback", feedbackID)

	var feedback models.Feedback
	err := db.Client.Get(ctx, key, &feedback)
	if err != nil {
		return nil, err
	}

	feedback.ID = feedbackID
	return &feedback, nil
}

func UpdateFeedback(db *DatastoreClient, feedback *models.Feedback) error {
	ctx := context.Background()
	key := NameKey("Feedback", feedback.ID)

	_, err := db.Client.Put(ctx, key, feedback)
	return err
}

func GetUserFeedback(db *DatastoreClient, userID string) ([]models.Feedback, error) {
	ctx := context.Background()
	query := NewQuery("Feedback").Filter("UserID =", userID).Order("-CreatedAt")

	var feedbacks []models.Feedback
	keys, err := db.Client.GetAll(ctx, query, &feedbacks)
	if err != nil {
		return nil, err
	}

	for i := range feedbacks {
		feedbacks[i].ID = keys[i].Name
	}

	return feedbacks, nil
}

func GetFeedbackStats(db *DatastoreClient) (map[string]interface{}, error) {
	ctx := context.Background()

	// Get total feedback count
	totalQuery := NewQuery("Feedback")
	totalCount, err := db.Client.Count(ctx, totalQuery)
	if err != nil {
		return nil, err
	}

	// Simplified stats - in real implementation, you'd process results after fetching
	return map[string]interface{}{
		"total_feedback": totalCount,
		"by_status": []map[string]interface{}{
			{"status": "new", "count": 0},
			{"status": "in_progress", "count": 0},
			{"status": "resolved", "count": 0},
		},
		"by_category": []map[string]interface{}{
			{"category": "bug_report", "count": 0},
			{"category": "feature_request", "count": 0},
			{"category": "general", "count": 0},
		},
	}, nil
}

// Support ticket operations
func CreateSupportTicket(db *DatastoreClient, ticket *models.SupportTicket) error {
	ctx := context.Background()
	key := NameKey("SupportTicket", ticket.ID)

	_, err := db.Client.Put(ctx, key, ticket)
	return err
}

func GetUserSupportTickets(db *DatastoreClient, userID, status string) ([]models.SupportTicket, error) {
	ctx := context.Background()
	query := NewQuery("SupportTicket").Filter("UserID =", userID).Order("-CreatedAt")

	if status != "" {
		query = query.Filter("Status =", status)
	}

	var tickets []models.SupportTicket
	keys, err := db.Client.GetAll(ctx, query, &tickets)
	if err != nil {
		return nil, err
	}

	for i := range tickets {
		tickets[i].ID = keys[i].Name
	}

	return tickets, nil
}

func GetSupportTicketByID(db *DatastoreClient, ticketID string) (*models.SupportTicket, error) {
	ctx := context.Background()
	key := NameKey("SupportTicket", ticketID)

	var ticket models.SupportTicket
	err := db.Client.Get(ctx, key, &ticket)
	if err != nil {
		return nil, err
	}

	ticket.ID = ticketID
	return &ticket, nil
}

func UpdateSupportTicket(db *DatastoreClient, ticket *models.SupportTicket) error {
	ctx := context.Background()
	key := NameKey("SupportTicket", ticket.ID)

	_, err := db.Client.Put(ctx, key, ticket)
	return err
}

func AddSupportMessage(db *DatastoreClient, message *models.SupportMessage) error {
	ctx := context.Background()
	key := NameKey("SupportMessage", message.ID)

	_, err := db.Client.Put(ctx, key, message)
	return err
}

func GetActiveSupportChannels(db *DatastoreClient) ([]models.SupportChannel, error) {
	ctx := context.Background()
	query := NewQuery("SupportChannel").Filter("IsActive =", true).Order("Name")

	var channels []models.SupportChannel
	keys, err := db.Client.GetAll(ctx, query, &channels)
	if err != nil {
		return nil, err
	}

	for i := range channels {
		channels[i].ID = keys[i].Name
	}

	return channels, nil
}

// Analytics operations
func TrackAnalyticsEvent(db *DatastoreClient, event *models.AnalyticsEvent) error {
	ctx := context.Background()
	key := NameKey("AnalyticsEvent", event.ID)

	_, err := db.Client.Put(ctx, key, event)
	return err
}

func BatchTrackAnalyticsEvents(db *DatastoreClient, events []*models.AnalyticsEvent) error {
	ctx := context.Background()

	var keys []*Key
	for _, event := range events {
		keys = append(keys, NameKey("AnalyticsEvent", event.ID))
	}

	entities := make([]interface{}, len(events))
	for i, event := range events {
		entities[i] = event
	}

	_, err := db.Client.PutMulti(ctx, keys, entities)
	return err
}

func GetUserAnalyticsConsent(db *DatastoreClient, userID string) (*models.AnalyticsConsent, error) {
	ctx := context.Background()
	query := NewQuery("AnalyticsConsent").Filter("UserID =", userID)

	var consents []models.AnalyticsConsent
	keys, err := db.Client.GetAll(ctx, query, &consents)
	if err != nil {
		return nil, err
	}

	if len(consents) == 0 {
		return nil, ErrNotFound
	}

	consents[0].ID = keys[0].Name
	return &consents[0], nil
}

func UpdateAnalyticsConsent(db *DatastoreClient, consent *models.AnalyticsConsent) error {
	ctx := context.Background()
	key := NameKey("AnalyticsConsent", consent.ID)

	_, err := db.Client.Put(ctx, key, consent)
	return err
}

func GetAnalyticsEvents(db *DatastoreClient, eventType, eventName, startDate, endDate, limit string) ([]models.AnalyticsEvent, error) {
	ctx := context.Background()
	query := NewQuery("AnalyticsEvent").Order("-Timestamp")

	if eventType != "" {
		query = query.Filter("EventType =", eventType)
	}
	if eventName != "" {
		query = query.Filter("EventName =", eventName)
	}
	if startDate != "" {
		// Parse date and add filter
		start, _ := time.Parse(time.RFC3339, startDate)
		query = query.Filter("Timestamp >=", start)
	}
	if endDate != "" {
		// Parse date and add filter
		end, _ := time.Parse(time.RFC3339, endDate)
		query = query.Filter("Timestamp <=", end)
	}
	if limit != "" {
		if lim, err := strconv.Atoi(limit); err == nil {
			query = query.Limit(lim)
		}
	}

	var events []models.AnalyticsEvent
	keys, err := db.Client.GetAll(ctx, query, &events)
	if err != nil {
		return nil, err
	}

	for i := range events {
		events[i].ID = keys[i].Name
	}

	return events, nil
}

func GetAnalyticsStats(db *DatastoreClient) (map[string]interface{}, error) {
	ctx := context.Background()

	// Get total events count
	totalQuery := NewQuery("AnalyticsEvent")
	totalCount, err := db.Client.Count(ctx, totalQuery)
	if err != nil {
		return nil, err
	}

	// Simplified stats - GroupBy is not available in Cloud Datastore
	return map[string]interface{}{
		"total_events": totalCount,
		"by_type": []map[string]interface{}{
			{"event_type": "user", "count": 0},
			{"event_type": "session", "count": 0},
			{"event_type": "interaction", "count": 0},
		},
	}, nil
}

// A/B Test operations
func GetABTests(db *DatastoreClient, status string) ([]models.ABTest, error) {
	ctx := context.Background()
	query := NewQuery("ABTest").Order("-CreatedAt")

	if status != "" {
		query = query.Filter("Status =", status)
	}

	var tests []models.ABTest
	keys, err := db.Client.GetAll(ctx, query, &tests)
	if err != nil {
		return nil, err
	}

	for i := range tests {
		tests[i].ID = keys[i].Name
	}

	return tests, nil
}

func GetABTestByID(db *DatastoreClient, testID string) (*models.ABTest, error) {
	ctx := context.Background()
	key := NameKey("ABTest", testID)

	var test models.ABTest
	err := db.Client.Get(ctx, key, &test)
	if err != nil {
		return nil, err
	}

	test.ID = testID
	return &test, nil
}

func CreateABTest(db *DatastoreClient, test *models.ABTest) error {
	ctx := context.Background()
	key := NameKey("ABTest", test.ID)

	_, err := db.Client.Put(ctx, key, test)
	return err
}

func UpdateABTest(db *DatastoreClient, test *models.ABTest) error {
	ctx := context.Background()
	key := NameKey("ABTest", test.ID)

	_, err := db.Client.Put(ctx, key, test)
	return err
}

func GetUserABTestAssignment(db *DatastoreClient, userID, testID string) (*models.UserABTestAssignment, error) {
	ctx := context.Background()
	query := NewQuery("UserABTestAssignment").Filter("UserID =", userID).Filter("ABTestID =", testID)

	var assignments []models.UserABTestAssignment
	keys, err := db.Client.GetAll(ctx, query, &assignments)
	if err != nil {
		return nil, err
	}

	if len(assignments) == 0 {
		return nil, ErrNotFound
	}

	assignments[0].ID = keys[0].Name
	return &assignments[0], nil
}

func CreateUserABTestAssignment(db *DatastoreClient, assignment *models.UserABTestAssignment) error {
	ctx := context.Background()
	key := NameKey("UserABTestAssignment", assignment.ID)

	_, err := db.Client.Put(ctx, key, assignment)
	return err
}

func GetABTestResults(db *DatastoreClient, testID string) ([]models.ABTestResult, error) {
	ctx := context.Background()
	query := NewQuery("ABTestResult").Filter("ABTestID =", testID).Order("-PeriodStart")

	var results []models.ABTestResult
	keys, err := db.Client.GetAll(ctx, query, &results)
	if err != nil {
		return nil, err
	}

	for i := range results {
		results[i].ID = keys[i].Name
	}

	return results, nil
}

// Community forum operations
func GetForums(db *DatastoreClient, category, forumType string) ([]models.CommunityForum, error) {
	ctx := context.Background()
	query := NewQuery("CommunityForum").Filter("IsActive =", true).Order("Name")

	if category != "" {
		query = query.Filter("Category =", category)
	}
	if forumType != "" {
		query = query.Filter("Type =", forumType)
	}

	var forums []models.CommunityForum
	keys, err := db.Client.GetAll(ctx, query, &forums)
	if err != nil {
		return nil, err
	}

	for i := range forums {
		forums[i].ID = keys[i].Name
	}

	return forums, nil
}

func GetForumByID(db *DatastoreClient, forumID string) (*models.CommunityForum, error) {
	ctx := context.Background()
	key := NameKey("CommunityForum", forumID)

	var forum models.CommunityForum
	err := db.Client.Get(ctx, key, &forum)
	if err != nil {
		return nil, err
	}

	forum.ID = forumID
	return &forum, nil
}

func UpdateForum(db *DatastoreClient, forum *models.CommunityForum) error {
	ctx := context.Background()
	key := NameKey("CommunityForum", forum.ID)

	_, err := db.Client.Put(ctx, key, forum)
	return err
}

func CreateForumPost(db *DatastoreClient, post *models.ForumPost) error {
	ctx := context.Background()
	key := NameKey("ForumPost", post.ID)

	_, err := db.Client.Put(ctx, key, post)
	return err
}

func GetForumPosts(db *DatastoreClient, forumID, status, tag, limit string) ([]models.ForumPost, error) {
	ctx := context.Background()
	query := NewQuery("ForumPost").Filter("ForumID =", forumID).Order("-CreatedAt")

	if status != "" {
		query = query.Filter("Status =", status)
	}
	if tag != "" {
		query = query.Filter("Tags =", tag)
	}
	if limit != "" {
		if lim, err := strconv.Atoi(limit); err == nil {
			query = query.Limit(lim)
		}
	}

	var posts []models.ForumPost
	keys, err := db.Client.GetAll(ctx, query, &posts)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].ID = keys[i].Name
	}

	return posts, nil
}

func GetForumPostByID(db *DatastoreClient, postID string) (*models.ForumPost, error) {
	ctx := context.Background()
	key := NameKey("ForumPost", postID)

	var post models.ForumPost
	err := db.Client.Get(ctx, key, &post)
	if err != nil {
		return nil, err
	}

	post.ID = postID
	return &post, nil
}

func UpdateForumPost(db *DatastoreClient, post *models.ForumPost) error {
	ctx := context.Background()
	key := NameKey("ForumPost", post.ID)

	_, err := db.Client.Put(ctx, key, post)
	return err
}

func DeleteForumPost(db *DatastoreClient, postID string) error {
	ctx := context.Background()
	key := NameKey("ForumPost", postID)

	return db.Client.Delete(ctx, key)
}

func CreateForumReply(db *DatastoreClient, reply *models.ForumReply) error {
	ctx := context.Background()
	key := NameKey("ForumReply", reply.ID)

	_, err := db.Client.Put(ctx, key, reply)
	return err
}

func GetForumReplies(db *DatastoreClient, postID string) ([]models.ForumReply, error) {
	ctx := context.Background()
	query := NewQuery("ForumReply").Filter("PostID =", postID).Order("CreatedAt")

	var replies []models.ForumReply
	keys, err := db.Client.GetAll(ctx, query, &replies)
	if err != nil {
		return nil, err
	}

	for i := range replies {
		replies[i].ID = keys[i].Name
	}

	return replies, nil
}

func GetForumReplyByID(db *DatastoreClient, replyID string) (*models.ForumReply, error) {
	ctx := context.Background()
	key := NameKey("ForumReply", replyID)

	var reply models.ForumReply
	err := db.Client.Get(ctx, key, &reply)
	if err != nil {
		return nil, err
	}

	reply.ID = replyID
	return &reply, nil
}

func UpdateForumReply(db *DatastoreClient, reply *models.ForumReply) error {
	ctx := context.Background()
	key := NameKey("ForumReply", reply.ID)

	_, err := db.Client.Put(ctx, key, reply)
	return err
}

func GetUserForumPosts(db *DatastoreClient, userID string) ([]models.ForumPost, error) {
	ctx := context.Background()
	query := NewQuery("ForumPost").Filter("UserID =", userID).Order("-CreatedAt")

	var posts []models.ForumPost
	keys, err := db.Client.GetAll(ctx, query, &posts)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].ID = keys[i].Name
	}

	return posts, nil
}

func SearchForumPosts(db *DatastoreClient, query, forumID, tag, limit string) ([]models.ForumPost, error) {
	ctx := context.Background()
	dsQuery := NewQuery("ForumPost").Order("-CreatedAt")

	if query != "" {
		dsQuery = dsQuery.Filter("Title =", query)
	}
	if forumID != "" {
		dsQuery = dsQuery.Filter("ForumID =", forumID)
	}
	if tag != "" {
		dsQuery = dsQuery.Filter("Tags =", tag)
	}
	if limit != "" {
		if lim, err := strconv.Atoi(limit); err == nil {
			dsQuery = dsQuery.Limit(lim)
		}
	}

	var posts []models.ForumPost
	keys, err := db.Client.GetAll(ctx, dsQuery, &posts)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].ID = keys[i].Name
	}

	return posts, nil
}

func GetPopularForumPosts(db *DatastoreClient, period, limit string) ([]models.ForumPost, error) {
	ctx := context.Background()
	query := NewQuery("ForumPost").Filter("Status =", "open").Order("-LikeCount")

	if limit != "" {
		if lim, err := strconv.Atoi(limit); err == nil {
			query = query.Limit(lim)
		}
	}

	var posts []models.ForumPost
	keys, err := db.Client.GetAll(ctx, query, &posts)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].ID = keys[i].Name
	}

	return posts, nil
}

func GetForumStats(db *DatastoreClient) (map[string]interface{}, error) {
	ctx := context.Background()

	// Get total posts count
	postsQuery := NewQuery("ForumPost")
	postsCount, err := db.Client.Count(ctx, postsQuery)
	if err != nil {
		return nil, err
	}

	// Get total replies count
	repliesQuery := NewQuery("ForumReply")
	repliesCount, err := db.Client.Count(ctx, repliesQuery)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_posts":   postsCount,
		"total_replies": repliesCount,
	}, nil
}

// Like operations (simplified - would need separate Like entities in real implementation)
func HasUserLikedPost(db *DatastoreClient, userID, postID string) (bool, error) {
	// Simplified implementation
	return false, nil
}

func LikeForumPost(db *DatastoreClient, userID, postID string) error {
	// Simplified implementation
	return nil
}

func UnlikeForumPost(db *DatastoreClient, userID, postID string) error {
	// Simplified implementation
	return nil
}

func HasUserLikedReply(db *DatastoreClient, userID, replyID string) (bool, error) {
	// Simplified implementation
	return false, nil
}

func LikeForumReply(db *DatastoreClient, userID, replyID string) error {
	// Simplified implementation
	return nil
}

func UnlikeForumReply(db *DatastoreClient, userID, replyID string) error {
	// Simplified implementation
	return nil
}

func UnmarkAllAnswers(db *DatastoreClient, postID string) error {
	// Simplified implementation
	return nil
}
