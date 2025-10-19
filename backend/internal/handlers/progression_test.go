package handlers

import (
	"net/http"
	"testing"

	"typing-master-backend/internal/testutil"
	"typing-master-backend/internal/models"
	"typing-master-backend/internal/middleware"

	"github.com/stretchr/testify/assert"
)

// TestGetLessonProgress tests lesson progress retrieval
func TestGetLessonProgress(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()
	cfg := testutil.CreateTestConfig()
	
	router.GET("/api/v1/lessons/:id/progress", middleware.AuthMiddleware(cfg.JWTSecret), GetLessonProgress(mockDB))
	
	// Requirement 2: Lesson progression tracking
	t.Run("Get Existing Progress", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		lesson := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro to Functions")
		
		// Create progress
		progress := &models.LessonProgress{
			ID:               "progress1",
			UserID:           user.ID,
			LessonID:         lesson.ID,
			CurrentStage:     "core",
			CompletedStages:  []string{"intro"},
			ProgressPercent:  50.0,
			TokensCovered:    []string{"function", "return"},
			IsUnlocked:       true,
			IsCompleted:      false,
			BestScore:        85,
			AttemptsCount:    3,
		}
		mockDB.PutLessonProgress("progress1", progress)
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons/lesson1/progress", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, user.ID, response["user_id"])
		assert.Equal(t, lesson.ID, response["lesson_id"])
		assert.Equal(t, "core", response["current_stage"])
		assert.Equal(t, 50.0, response["progress_percent"])
		assert.Equal(t, float64(85), response["best_score"])
	})
	
	t.Run("No Progress Found", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro to Functions")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons/lesson1/progress", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Contains(t, response, "message")
		assert.Contains(t, response["message"], "No progress found")
	})
}

// TestUpdateLessonProgress tests lesson progress update
func TestUpdateLessonProgress(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()
	cfg := testutil.CreateTestConfig()
	
	router.PUT("/api/v1/lessons/:id/progress", middleware.AuthMiddleware(cfg.JWTSecret), UpdateLessonProgress(mockDB))
	
	// Requirement 2: Progression from Intro → Core → Idioms → Advanced → Review
	t.Run("Create New Progress", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro to Functions")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		reqBody := map[string]interface{}{
			"current_stage":    "intro",
			"progress_percent": 25.0,
			"tokens_covered":   []string{"function"},
			"score":            75,
		}
		
		w := testutil.MakeRequest(router, "PUT", "/api/v1/lessons/lesson1/progress", reqBody, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, "intro", response["current_stage"])
		assert.Equal(t, 25.0, response["progress_percent"])
		assert.Equal(t, float64(75), response["best_score"])
		assert.Equal(t, float64(1), response["attempts_count"])
	})
	
	t.Run("Update Existing Progress", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		lesson := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro to Functions")
		
		// Create initial progress
		progress := &models.LessonProgress{
			ID:               "progress1",
			UserID:           user.ID,
			LessonID:         lesson.ID,
			CurrentStage:     "intro",
			ProgressPercent:  25.0,
			BestScore:        75,
			AttemptsCount:    1,
		}
		mockDB.PutLessonProgress("progress1", progress)
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		reqBody := map[string]interface{}{
			"current_stage":    "core",
			"progress_percent": 50.0,
			"score":            85,
		}
		
		w := testutil.MakeRequest(router, "PUT", "/api/v1/lessons/lesson1/progress", reqBody, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, "core", response["current_stage"])
		assert.Equal(t, 50.0, response["progress_percent"])
		assert.Equal(t, float64(85), response["best_score"]) // Updated to higher score
		assert.Equal(t, float64(2), response["attempts_count"]) // Incremented
	})
	
	t.Run("Complete Lesson", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro to Functions")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		reqBody := map[string]interface{}{
			"current_stage":    "review",
			"progress_percent": 100.0,
			"score":            95,
		}
		
		w := testutil.MakeRequest(router, "PUT", "/api/v1/lessons/lesson1/progress", reqBody, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, true, response["is_completed"])
		assert.Equal(t, 100.0, response["progress_percent"])
	})
	
	t.Run("Track Best Score", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		lesson := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro to Functions")
		
		// Create progress with high score
		progress := &models.LessonProgress{
			ID:            "progress1",
			UserID:        user.ID,
			LessonID:      lesson.ID,
			BestScore:     90,
			AttemptsCount: 1,
		}
		mockDB.PutLessonProgress("progress1", progress)
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		// Try to update with lower score
		reqBody := map[string]interface{}{
			"score": 80,
		}
		
		w := testutil.MakeRequest(router, "PUT", "/api/v1/lessons/lesson1/progress", reqBody, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		// Best score should remain unchanged
		assert.Equal(t, float64(90), response["best_score"])
		assert.Equal(t, float64(2), response["attempts_count"])
	})
}

// TestCheckLessonPrerequisites tests prerequisite checking
func TestCheckLessonPrerequisites(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()
	cfg := testutil.CreateTestConfig()
	
	router.GET("/api/v1/lessons/:id/prerequisites", middleware.AuthMiddleware(cfg.JWTSecret), CheckLessonPrerequisites(mockDB))
	
	// Requirement 2: Prerequisites system
	t.Run("Lesson Unlocked - All Prerequisites Met", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		
		// Create prerequisite lesson
		prereqLesson := testutil.CreateTestLesson(mockDB, "lesson_prereq", "javascript", "Basics")
		
		// Create target lesson with prerequisite
		targetLesson := &models.Lesson{
			ID:            "lesson_advanced",
			LanguageID:    "javascript",
			Title:         "Advanced Topics",
			Prerequisites: []string{prereqLesson.ID},
		}
		mockDB.PutLesson(targetLesson.ID, targetLesson)
		
		// Mark prerequisite as completed
		progress := &models.LessonProgress{
			ID:          "progress1",
			UserID:      user.ID,
			LessonID:    prereqLesson.ID,
			IsCompleted: true,
		}
		mockDB.PutLessonProgress("progress1", progress)
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons/lesson_advanced/prerequisites", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, true, response["is_unlocked"])
		
		missingPrereqs := response["missing_prereqs"].([]interface{})
		assert.Equal(t, 0, len(missingPrereqs))
	})
	
	t.Run("Lesson Locked - Missing Prerequisites", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		
		// Create prerequisite lesson
		prereqLesson := testutil.CreateTestLesson(mockDB, "lesson_prereq", "javascript", "Basics")
		
		// Create target lesson with prerequisite
		targetLesson := &models.Lesson{
			ID:            "lesson_advanced",
			LanguageID:    "javascript",
			Title:         "Advanced Topics",
			Prerequisites: []string{prereqLesson.ID},
		}
		mockDB.PutLesson(targetLesson.ID, targetLesson)
		
		// No progress for prerequisite
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons/lesson_advanced/prerequisites", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, false, response["is_unlocked"])
		
		missingPrereqs := response["missing_prereqs"].([]interface{})
		assert.Equal(t, 1, len(missingPrereqs))
		assert.Equal(t, prereqLesson.ID, missingPrereqs[0])
	})
	
	t.Run("Lesson with No Prerequisites", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		
		// Create lesson with no prerequisites
		lesson := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons/"+lesson.ID+"/prerequisites", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, true, response["is_unlocked"])
	})
}

// TestGetLessonProgressionFlow tests lesson flow retrieval
func TestGetLessonProgressionFlow(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()
	cfg := testutil.CreateTestConfig()
	
	router.GET("/api/v1/lessons/:id/flow", middleware.AuthMiddleware(cfg.JWTSecret), GetLessonProgressionFlow(mockDB))
	
	// Requirement 2: Structured progression flow
	t.Run("Get Lesson Progression Flow", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		lesson := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Functions")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons/"+lesson.ID+"/flow", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, lesson.ID, response["lesson_id"])
		assert.Contains(t, response, "flow")
		
		flow := response["flow"].([]interface{})
		assert.Equal(t, 5, len(flow)) // intro, core, idioms, advanced, review
		
		// Verify stages
		stages := []string{}
		for _, stage := range flow {
			stageMap := stage.(map[string]interface{})
			stages = append(stages, stageMap["stage"].(string))
		}
		
		assert.Contains(t, stages, "intro")
		assert.Contains(t, stages, "core")
		assert.Contains(t, stages, "idioms")
		assert.Contains(t, stages, "advanced")
		assert.Contains(t, stages, "review")
	})
	
	t.Run("Flow Includes Estimated Times", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		lesson := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Functions")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons/"+lesson.ID+"/flow", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		flow := response["flow"].([]interface{})
		
		for _, stage := range flow {
			stageMap := stage.(map[string]interface{})
			assert.Contains(t, stageMap, "estimated_minutes")
			assert.Contains(t, stageMap, "description")
		}
	})
}

// TestGetUserProgressionSummary tests overall progression summary
func TestGetUserProgressionSummary(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()
	
	router.GET("/api/v1/progression/summary", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetUserProgressionSummary(mockDB))
	
	// Requirement 3: Comprehensive metrics tracking
	t.Run("Get Progression Summary", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		
		// Create lessons
		lesson1 := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Basics")
		lesson2 := testutil.CreateTestLesson(mockDB, "lesson2", "javascript", "Advanced")
		testutil.CreateTestLesson(mockDB, "lesson3", "python", "Intro")
		
		// Create progress for some lessons
		progress1 := &models.LessonProgress{
			ID:              "progress1",
			UserID:          user.ID,
			LessonID:        lesson1.ID,
			ProgressPercent: 100.0,
			IsCompleted:     true,
			BestScore:       90,
		}
		mockDB.PutLessonProgress("progress1", progress1)
		
		progress2 := &models.LessonProgress{
			ID:              "progress2",
			UserID:          user.ID,
			LessonID:        lesson2.ID,
			ProgressPercent: 50.0,
			IsCompleted:     false,
			BestScore:       75,
		}
		mockDB.PutLessonProgress("progress2", progress2)
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/progression/summary", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Contains(t, response, "total_lessons")
		assert.Contains(t, response, "completed_lessons")
		assert.Contains(t, response, "overall_progress")
		assert.Contains(t, response, "average_score")
		assert.Contains(t, response, "completion_rate")
		
		assert.Equal(t, float64(3), response["total_lessons"])
		assert.Equal(t, float64(1), response["completed_lessons"])
		assert.Greater(t, response["average_score"], float64(0))
	})
	
	t.Run("Summary for User with No Progress", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		
		// Create lessons but no progress
		testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Basics")
		testutil.CreateTestLesson(mockDB, "lesson2", "javascript", "Advanced")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/progression/summary", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		assert.Equal(t, float64(0), response["completed_lessons"])
		assert.Equal(t, float64(0), response["overall_progress"])
		assert.Equal(t, float64(0), response["completion_rate"])
	})
	
	t.Run("Summary Calculates Completion Rate", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		
		// Create 4 lessons
		for i := 1; i <= 4; i++ {
			lesson := testutil.CreateTestLesson(mockDB, "lesson"+string(rune(i)), "javascript", "Lesson")
			
			// Complete 2 of them
			if i <= 2 {
				progress := &models.LessonProgress{
					ID:          "progress" + string(rune(i)),
					UserID:      user.ID,
					LessonID:    lesson.ID,
					IsCompleted: true,
					BestScore:   85,
				}
				mockDB.PutLessonProgress("progress"+string(rune(i)), progress)
			}
		}
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		w := testutil.MakeRequest(router, "GET", "/api/v1/progression/summary", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)
		
		// 2 out of 4 = 50% completion rate
		assert.Equal(t, float64(50), response["completion_rate"])
	})
}

// TestProgressionStageTransitions tests stage transition logic
func TestProgressionStageTransitions(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()
	
	router.PUT("/api/v1/lessons/:id/progress", middleware.AuthMiddleware("test-secret-key-for-testing-only"), UpdateLessonProgress(mockDB))
	
	// Requirement 2: Progression stages
	t.Run("Progress Through All Stages", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Complete Journey")
		
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}
		
		stages := []struct {
			stage    string
			progress float64
		}{
			{"intro", 20.0},
			{"core", 40.0},
			{"idioms", 60.0},
			{"advanced", 80.0},
			{"review", 100.0},
		}
		
		for _, s := range stages {
			reqBody := map[string]interface{}{
				"current_stage":    s.stage,
				"progress_percent": s.progress,
				"score":            85,
			}
			
			w := testutil.MakeRequest(router, "PUT", "/api/v1/lessons/lesson1/progress", reqBody, headers)
			assert.Equal(t, http.StatusOK, w.Code)
			
			var response map[string]interface{}
			testutil.ParseJSON(w.Body.Bytes(), &response)
			
			assert.Equal(t, s.stage, response["current_stage"])
			assert.Equal(t, s.progress, response["progress_percent"])
		}
	})
}
