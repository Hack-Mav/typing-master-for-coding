package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// GetForums returns available community forums
func GetForums(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		forumType := c.Query("type")

		forums, err := database.GetForums(db, category, forumType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forums"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"forums": forums,
			"total":  len(forums),
		})
	}
}

// GetForum returns a specific forum
func GetForum(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		forumID := c.Param("id")

		forum, err := database.GetForumByID(db, forumID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
			return
		}

		c.JSON(http.StatusOK, forum)
	}
}

// CreateForumPost creates a new forum post
func CreateForumPost(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			ForumID string   `json:"forum_id" binding:"required"`
			Title   string   `json:"title" binding:"required"`
			Content string   `json:"content" binding:"required"`
			Type    string   `json:"type"`
			Tags    []string `json:"tags"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if forum exists
		forum, err := database.GetForumByID(db, req.ForumID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
			return
		}

		if !forum.IsActive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Forum is not active"})
			return
		}

		post := &models.ForumPost{
			ID:        generateID(),
			ForumID:   req.ForumID,
			UserID:    userID,
			Title:     req.Title,
			Content:   req.Content,
			Type:      req.Type,
			Status:    "open",
			Tags:      req.Tags,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.CreateForumPost(db, post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create forum post"})
			return
		}

		// Update forum activity
		forum.PostCount++
		forum.LastActivity = time.Now()
		database.UpdateForum(db, forum)

		c.JSON(http.StatusCreated, gin.H{
			"post":    post,
			"message": "Forum post created successfully",
		})
	}
}

// GetForumPosts returns posts in a forum
func GetForumPosts(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		forumID := c.Param("forum_id")
		status := c.Query("status")
		tag := c.Query("tag")
		limit := c.DefaultQuery("limit", "20")

		posts, err := database.GetForumPosts(db, forumID, status, tag, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forum posts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"posts": posts,
			"total": len(posts),
		})
	}
}

// GetForumPost returns a specific forum post
func GetForumPost(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")

		post, err := database.GetForumPostByID(db, postID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum post not found"})
			return
		}

		// Increment view count
		post.ViewCount++
		database.UpdateForumPost(db, post)

		c.JSON(http.StatusOK, post)
	}
}

// UpdateForumPost updates a forum post
func UpdateForumPost(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		postID := c.Param("id")

		var req struct {
			Title   string   `json:"title"`
			Content string   `json:"content"`
			Tags    []string `json:"tags"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		post, err := database.GetForumPostByID(db, postID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum post not found"})
			return
		}

		// Check if user owns the post or is admin
		if post.UserID != userID {
			role := c.GetString("role")
			if role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}

		// Update fields
		if req.Title != "" {
			post.Title = req.Title
		}
		if req.Content != "" {
			post.Content = req.Content
		}
		if len(req.Tags) > 0 {
			post.Tags = req.Tags
		}
		post.UpdatedAt = time.Now()

		if err := database.UpdateForumPost(db, post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update forum post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"post":    post,
			"message": "Forum post updated successfully",
		})
	}
}

// DeleteForumPost deletes a forum post
func DeleteForumPost(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		postID := c.Param("id")

		post, err := database.GetForumPostByID(db, postID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum post not found"})
			return
		}

		// Check if user owns the post or is admin
		if post.UserID != userID {
			role := c.GetString("role")
			if role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}

		if err := database.DeleteForumPost(db, postID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete forum post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Forum post deleted successfully"})
	}
}

// CreateForumReply creates a reply to a forum post
func CreateForumReply(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		postID := c.Param("id")

		var req struct {
			Content string `json:"content" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if post exists and is not locked
		post, err := database.GetForumPostByID(db, postID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum post not found"})
			return
		}

		if post.IsLocked {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Post is locked for replies"})
			return
		}

		reply := &models.ForumReply{
			ID:        generateID(),
			PostID:    postID,
			UserID:    userID,
			Content:   req.Content,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.CreateForumReply(db, reply); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reply"})
			return
		}

		// Update post
		post.ReplyCount++
		now := time.Now()
		post.LastReplyAt = &now
		post.UpdatedAt = time.Now()
		database.UpdateForumPost(db, post)

		c.JSON(http.StatusCreated, gin.H{
			"reply":   reply,
			"message": "Reply created successfully",
		})
	}
}

// GetForumReplies returns replies to a forum post
func GetForumReplies(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")

		replies, err := database.GetForumReplies(db, postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch replies"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"replies": replies,
			"total":   len(replies),
		})
	}
}

// LikeForumPost likes a forum post
func LikeForumPost(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		postID := c.Param("id")

		post, err := database.GetForumPostByID(db, postID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum post not found"})
			return
		}

		// Check if user already liked the post
		alreadyLiked, err := database.HasUserLikedPost(db, userID, postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check like status"})
			return
		}

		if alreadyLiked {
			// Unlike the post
			if err := database.UnlikeForumPost(db, userID, postID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike post"})
				return
			}
			post.LikeCount--
		} else {
			// Like the post
			if err := database.LikeForumPost(db, userID, postID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like post"})
				return
			}
			post.LikeCount++
		}

		if err := database.UpdateForumPost(db, post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"like_count": post.LikeCount,
			"liked":      !alreadyLiked,
			"message":    "Post like status updated",
		})
	}
}

// LikeForumReply likes a forum reply
func LikeForumReply(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		replyID := c.Param("reply_id")

		reply, err := database.GetForumReplyByID(db, replyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum reply not found"})
			return
		}

		// Check if user already liked the reply
		alreadyLiked, err := database.HasUserLikedReply(db, userID, replyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check like status"})
			return
		}

		if alreadyLiked {
			// Unlike the reply
			if err := database.UnlikeForumReply(db, userID, replyID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike reply"})
				return
			}
			reply.LikeCount--
		} else {
			// Like the reply
			if err := database.LikeForumReply(db, userID, replyID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like reply"})
				return
			}
			reply.LikeCount++
		}

		if err := database.UpdateForumReply(db, reply); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reply"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"like_count": reply.LikeCount,
			"liked":      !alreadyLiked,
			"message":    "Reply like status updated",
		})
	}
}

// MarkReplyAsAnswer marks a reply as the best answer
func MarkReplyAsAnswer(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		postID := c.Param("id")
		replyID := c.Param("reply_id")

		post, err := database.GetForumPostByID(db, postID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum post not found"})
			return
		}

		// Check if user owns the post or is admin
		if post.UserID != userID {
			role := c.GetString("role")
			if role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}

		reply, err := database.GetForumReplyByID(db, replyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Forum reply not found"})
			return
		}

		// Unmark any existing answers
		if err := database.UnmarkAllAnswers(db, postID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmark existing answers"})
			return
		}

		// Mark this reply as answer
		reply.IsAnswer = true
		if err := database.UpdateForumReply(db, reply); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark answer"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"reply":   reply,
			"message": "Reply marked as answer",
		})
	}
}

// GetUserPosts returns posts created by the current user
func GetUserPosts(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		posts, err := database.GetUserForumPosts(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user posts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"posts": posts,
			"total": len(posts),
		})
	}
}

// SearchForumPosts searches forum posts
func SearchForumPosts(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		forumID := c.Query("forum_id")
		tag := c.Query("tag")
		limit := c.DefaultQuery("limit", "20")

		posts, err := database.SearchForumPosts(db, query, forumID, tag, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search forum posts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"posts": posts,
			"total": len(posts),
			"query": query,
		})
	}
}

// GetPopularPosts returns popular forum posts
func GetPopularPosts(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		period := c.DefaultQuery("period", "week") // "day", "week", "month"
		limit := c.DefaultQuery("limit", "10")

		posts, err := database.GetPopularForumPosts(db, period, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch popular posts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"posts":  posts,
			"total":  len(posts),
			"period": period,
		})
	}
}

// GetForumStats returns forum statistics
func GetForumStats(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := database.GetForumStats(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forum stats"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
