package models

import (
	"time"
)

// ForumCategory represents a category in the forum
type ForumCategory struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description,omitempty" db:"description"`
	Slug         string    `json:"slug" db:"slug"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	IsPublic     bool      `json:"is_public" db:"is_public"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	TopicCount   int       `json:"topic_count,omitempty" db:"-"`
	PostCount    int       `json:"post_count,omitempty" db:"-"`
}

// ForumTopic represents a topic in the forum
type ForumTopic struct {
	ID         string    `json:"id" db:"id"`
	CategoryID string    `json:"category_id" db:"category_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Title      string    `json:"title" db:"title"`
	Slug       string    `json:"slug" db:"slug"`
	Content    string    `json:"content" db:"content"`
	IsPinned   bool      `json:"is_pinned" db:"is_pinned"`
	IsLocked   bool      `json:"is_locked" db:"is_locked"`
	ViewCount  int       `json:"view_count" db:"view_count"`
	LastPostAt time.Time `json:"last_post_at,omitempty" db:"last_post_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Relations
	Category  *ForumCategory `json:"category,omitempty" db:"-"`
	User      *User          `json:"user,omitempty" db:"-"`
	PostCount int            `json:"post_count,omitempty" db:"-"`
	LastPost  *ForumPost     `json:"last_post,omitempty" db:"-"`
}

// ForumTopicPost represents a post in a forum topic
type ForumTopicPost struct {
	ID          string    `json:"id" db:"id"`
	TopicID     string    `json:"topic_id" db:"topic_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Content     string    `json:"content" db:"content"`
	IsFirstPost bool      `json:"is_first_post" db:"is_first_post"`
	IsEdited    bool      `json:"is_edited" db:"is_edited"`
	EditedAt    time.Time `json:"edited_at,omitempty" db:"edited_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relations
	Topic     *ForumTopic `json:"topic,omitempty" db:"-"`
	User      *User       `json:"user,omitempty" db:"-"`
	LikeCount int         `json:"like_count,omitempty" db:"-"`
	IsLiked   bool        `json:"is_liked,omitempty" db:"-"`
}

// ForumPostLike represents a like on a forum post
type ForumPostLike struct {
	PostID    string    `json:"post_id" db:"post_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserSupportTicket represents a user support ticket
type UserSupportTicket struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	AssigneeID *string   `json:"assignee_id,omitempty" db:"assignee_id"`
	Title      string    `json:"title" db:"title"`
	Status     string    `json:"status" db:"status"`
	Priority   string    `json:"priority" db:"priority"`
	Category   string    `json:"category" db:"category"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Relations
	User         *User                   `json:"user,omitempty" db:"-"`
	Assignee     *User                   `json:"assignee,omitempty" db:"-"`
	Messages     []*SupportTicketMessage `json:"messages,omitempty" db:"-"`
	MessageCount int                     `json:"message_count,omitempty" db:"-"`
}

// SupportTicketMessage represents a message in a support ticket
type SupportTicketMessage struct {
	ID             string    `json:"id" db:"id"`
	TicketID       string    `json:"ticket_id" db:"ticket_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Content        string    `json:"content" db:"content"`
	IsInternalNote bool      `json:"is_internal_note" db:"is_internal_note"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`

	// Relations
	User   *User              `json:"user,omitempty" db:"-"`
	Ticket *UserSupportTicket `json:"-" db:"-"`
}

// Request/Response models

// CreateTopicRequest represents the request to create a new forum topic
type CreateTopicRequest struct {
	CategoryID string `json:"category_id" binding:"required,uuid"`
	Title      string `json:"title" binding:"required,min=5,max=255"`
	Content    string `json:"content" binding:"required,min=10"`
}

// CreatePostRequest represents the request to create a new forum post
type CreatePostRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

// CreateTicketRequest represents the request to create a new support ticket
type CreateTicketRequest struct {
	Title    string `json:"title" binding:"required,min=5,max=255"`
	Content  string `json:"content" binding:"required,min=10"`
	Category string `json:"category" binding:"required,oneof=bug feature account billing other"`
	Priority string `json:"priority" binding:"oneof=low medium high critical"`
}

// CreateMessageRequest represents the request to add a message to a support ticket
type CreateMessageRequest struct {
	Content        string `json:"content" binding:"required,min=1"`
	IsInternalNote bool   `json:"is_internal_note"`
}

// UpdateTicketRequest represents the request to update a support ticket
type UpdateTicketRequest struct {
	Status     *string `json:"status,omitempty"`
	Priority   *string `json:"priority,omitempty"`
	AssigneeID *string `json:"assignee_id,omitempty"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
