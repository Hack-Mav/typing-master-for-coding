package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/typing-master-for-coding-backend/internal/models"
)

type ForumSupportRepository struct {
	db *sqlx.DB
}

func NewForumSupportRepository(db *sqlx.DB) *ForumSupportRepository {
	return &ForumSupportRepository{db: db}
}

// Forum Category Operations

func (r *ForumSupportRepository) GetCategories(ctx context.Context) ([]models.ForumCategory, error) {
	query := `
		SELECT 
			c.*,
			COUNT(DISTINCT t.id) as topic_count,
			COUNT(DISTINCT p.id) as post_count
		FROM forum_categories c
		LEFT JOIN forum_topics t ON t.category_id = c.id
		LEFT JOIN forum_topic_posts p ON p.topic_id = t.id
		WHERE c.is_public = true
		GROUP BY c.id
		ORDER BY c.display_order, c.name
	`

	var categories []models.ForumCategory
	err := r.db.SelectContext(ctx, &categories, query)
	return categories, err
}

func (r *ForumSupportRepository) GetCategoryBySlug(ctx context.Context, slug string) (*models.ForumCategory, error) {
	query := `
		SELECT * FROM forum_categories 
		WHERE slug = $1 AND is_public = true
	`

	var category models.ForumCategory
	err := r.db.GetContext(ctx, &category, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
	}

	return &category, nil
}

// Forum Topic Operations

func (r *ForumSupportRepository) GetTopics(ctx context.Context, categoryID string, page, pageSize int) ([]models.ForumTopic, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM forum_topics WHERE category_id = $1`
	err := r.db.GetContext(ctx, &total, countQuery, categoryID)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	query := `
		SELECT 
			t.*,
			u.id as "user.id",
			u.username as "user.username",
			u.avatar_url as "user.avatar_url",
			p.content as "last_post.content",
			p.created_at as "last_post.created_at",
			u2.username as "last_post.user.username"
		FROM forum_topics t
		JOIN users u ON u.id = t.user_id
		LEFT JOIN LATERAL (
			SELECT * FROM forum_topic_posts 
			WHERE topic_id = t.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) p ON true
		LEFT JOIN users u2 ON u2.id = p.user_id
		WHERE t.category_id = $1
		ORDER BY t.is_pinned DESC, t.last_post_at DESC NULLS LAST
		LIMIT $2 OFFSET $3
	`

	var topics []models.ForumTopic
	err = r.db.SelectContext(ctx, &topics, query, categoryID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get post counts for each topic
	for i := range topics {
		err = r.db.GetContext(ctx, &topics[i].PostCount,
			"SELECT COUNT(*) FROM forum_topic_posts WHERE topic_id = $1", topics[i].ID)
		if err != nil {
			return nil, 0, err
		}
	}

	return topics, total, nil
}

func (r *ForumSupportRepository) GetTopicBySlug(ctx context.Context, categoryID, slug string) (*models.ForumTopic, error) {
	query := `
		SELECT 
			t.*,
			u.id as "user.id",
			u.username as "user.username",
			u.avatar_url as "user.avatar_url"
		FROM forum_topics t
		JOIN users u ON u.id = t.user_id
		WHERE t.slug = $1 AND t.category_id = $2
	`

	var topic models.ForumTopic
	err := r.db.GetContext(ctx, &topic, query, slug, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("topic not found")
		}
		return nil, err
	}

	// Get post count
	err = r.db.GetContext(ctx, &topic.PostCount,
		"SELECT COUNT(*) FROM forum_topic_posts WHERE topic_id = $1", topic.ID)
	if err != nil {
		return nil, err
	}

	return &topic, nil
}

func (r *ForumSupportRepository) CreateTopic(ctx context.Context, topic *models.ForumTopic) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert topic
	query := `
		INSERT INTO forum_topics 
		(id, category_id, user_id, title, slug, content, is_pinned, is_locked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, query,
		topic.ID,
		topic.CategoryID,
		topic.UserID,
		topic.Title,
		topic.Slug,
		topic.Content,
		topic.IsPinned,
		topic.IsLocked,
		topic.CreatedAt,
		topic.UpdatedAt,
	).Scan(&topic.ID, &topic.CreatedAt, &topic.UpdatedAt)
	if err != nil {
		return err
	}

	// Create first post
	post := &models.ForumTopicPost{
		ID:          generateID(),
		TopicID:     topic.ID,
		UserID:      topic.UserID,
		Content:     topic.Content,
		IsFirstPost: true,
		CreatedAt:   topic.CreatedAt,
		UpdatedAt:   topic.UpdatedAt,
	}

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO forum_topic_posts 
		(id, topic_id, user_id, content, is_first_post, created_at, updated_at)
		VALUES (:id, :topic_id, :user_id, :content, :is_first_post, :created_at, :updated_at)
	`, post)

	if err != nil {
		return err
	}

	// Update topic's last post
	_, err = tx.ExecContext(ctx, `
		UPDATE forum_topics 
		SET last_post_at = $1, updated_at = $2
		WHERE id = $3
	`, topic.CreatedAt, topic.UpdatedAt, topic.ID)

	if err != nil {
		return err
	}

	// Update category's post count
	_, err = tx.ExecContext(ctx, `
		UPDATE forum_categories 
		SET post_count = post_count + 1, 
		    topic_count = topic_count + 1,
		    updated_at = $1
		WHERE id = $2
	`, time.Now(), topic.CategoryID)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// Support Ticket Operations

func (r *ForumSupportRepository) CreateSupportTicket(ctx context.Context, ticket *models.UserSupportTicket) error {
	query := `
		INSERT INTO support_tickets 
		(id, user_id, assignee_id, title, status, priority, category, created_at, updated_at)
		VALUES (:id, :user_id, :assignee_id, :title, :status, :priority, :category, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, ticket)
	return err
}

func (r *ForumSupportRepository) GetUserTickets(ctx context.Context, userID string, page, pageSize int) ([]models.UserSupportTicket, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM support_tickets WHERE user_id = $1`
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	query := `
		SELECT 
			t.*,
			u.id as "user.id",
			u.username as "user.username",
			u2.id as "assignee.id",
			u2.username as "assignee.username"
		FROM support_tickets t
		JOIN users u ON u.id = t.user_id
		LEFT JOIN users u2 ON u2.id = t.assignee_id
		WHERE t.user_id = $1
		ORDER BY t.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	var tickets []models.UserSupportTicket
	err = r.db.SelectContext(ctx, &tickets, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get message counts
	for i := range tickets {
		err = r.db.GetContext(ctx, &tickets[i].MessageCount,
			"SELECT COUNT(*) FROM support_messages WHERE ticket_id = $1", tickets[i].ID)
		if err != nil {
			return nil, 0, err
		}
	}

	return tickets, total, nil
}

func (r *ForumSupportRepository) AddSupportMessage(ctx context.Context, message *models.SupportTicketMessage) error {
	query := `
		INSERT INTO support_messages 
		(id, ticket_id, user_id, content, is_internal_note, created_at)
		VALUES (:id, :ticket_id, :user_id, :content, :is_internal_note, :created_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, message)
	if err != nil {
		return err
	}

	// Update ticket's updated_at
	_, err = r.db.ExecContext(ctx, `
		UPDATE support_tickets 
		SET updated_at = $1 
		WHERE id = $2
	`, time.Now(), message.TicketID)

	return err
}

// Helper function to generate UUID
func generateID() string {
	// This is a simplified version. In production, use a proper UUID generator
	// like github.com/google/uuid
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
