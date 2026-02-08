package scoring

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/typing-master-for-coding-backend/internal/models"

	"cloud.google.com/go/datastore"
)

// TournamentService handles tournament creation, management, and results
type TournamentService struct {
	datastoreClient DatastoreClient
	antiCheatSvc    AntiCheatService
	leaderboardSvc  LeaderboardService
}

// NewTournamentService creates a new tournament service
func NewTournamentService(datastoreClient DatastoreClient, antiCheatSvc AntiCheatService, leaderboardSvc LeaderboardService) *TournamentService {
	return &TournamentService{
		datastoreClient: datastoreClient,
		antiCheatSvc:    antiCheatSvc,
		leaderboardSvc:  leaderboardSvc,
	}
}

// CreateTournament creates a new tournament
func (s *TournamentService) CreateTournament(ctx context.Context, req *CreateTournamentRequest) (*models.Tournament, error) {
	tournament := &models.Tournament{
		Title:                req.Title,
		Description:          req.Description,
		LanguageID:           req.LanguageID,
		Mode:                 req.Mode,
		VerificationRequired: req.VerificationRequired,
		VerificationMethod:   req.VerificationMethod,
		AntiCheatLevel:       req.AntiCheatLevel,
		RegistrationStart:    req.RegistrationStart,
		RegistrationEnd:      req.RegistrationEnd,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		DurationMinutes:      req.DurationMinutes,
		MaxParticipants:      req.MaxParticipants,
		PrizePool:            req.PrizePool,
		PrizeDistribution:    req.PrizeDistribution,
		BadgeReward:          req.BadgeReward,
		Status:               "upcoming",
		CreatedBy:            req.CreatedBy,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Generate tournament ID
	tournament.ID = fmt.Sprintf("tournament_%d", time.Now().Unix())

	// Validate tournament configuration
	err := s.validateTournament(tournament)
	if err != nil {
		return nil, fmt.Errorf("invalid tournament configuration: %w", err)
	}

	// Store tournament
	key := datastore.NameKey("Tournament", tournament.ID, nil)
	_, err = s.datastoreClient.Put(ctx, key, tournament)
	if err != nil {
		return nil, fmt.Errorf("failed to store tournament: %w", err)
	}

	log.Printf("Created tournament: %s", tournament.ID)
	return tournament, nil
}

// RegisterParticipant registers a user for a tournament
func (s *TournamentService) RegisterParticipant(ctx context.Context, tournamentID, userID string) (*models.TournamentParticipant, error) {
	// Get tournament
	tournament, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found: %w", err)
	}

	// Check if tournament is in registration phase
	if tournament.Status != "registration" {
		return nil, fmt.Errorf("tournament is not accepting registrations")
	}

	// Check if user is already registered
	existing, _ := s.getParticipant(ctx, tournamentID, userID)
	if existing != nil {
		return nil, fmt.Errorf("user already registered for this tournament")
	}

	// Check participant limit
	if tournament.CurrentParticipants >= tournament.MaxParticipants {
		return nil, fmt.Errorf("tournament is full")
	}

	// Create participant record
	participant := &models.TournamentParticipant{
		TournamentID:       tournamentID,
		UserID:             userID,
		RegisteredAt:       time.Now(),
		VerificationStatus: "pending",
		Status:             "registered",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Perform verification if required
	if tournament.VerificationRequired {
		verificationResult, err := s.antiCheatSvc.VerifyTournamentParticipant(ctx, tournamentID, userID, tournament.VerificationMethod)
		if err != nil {
			return nil, fmt.Errorf("verification failed: %w", err)
		}

		participant.VerificationStatus = verificationResult.Status
		participant.VerificationData = verificationResult.Details

		if verificationResult.Status != "verified" {
			participant.Status = "verification_failed"
		}
	} else {
		participant.VerificationStatus = "not_required"
	}

	// Store participant
	err = s.storeParticipant(ctx, participant)
	if err != nil {
		return nil, fmt.Errorf("failed to store participant: %w", err)
	}

	// Update tournament participant count
	tournament.CurrentParticipants++
	tournament.ParticipantIDs = append(tournament.ParticipantIDs, userID)
	tournament.UpdatedAt = time.Now()

	err = s.updateTournament(ctx, tournament)
	if err != nil {
		log.Printf("Failed to update tournament participant count: %v", err)
	}

	log.Printf("User %s registered for tournament %s", userID, tournamentID)
	return participant, nil
}

// StartTournament starts a tournament
func (s *TournamentService) StartTournament(ctx context.Context, tournamentID string) error {
	tournament, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Check if tournament can be started
	if tournament.Status != "registration" {
		return fmt.Errorf("tournament cannot be started in current status: %s", tournament.Status)
	}

	now := time.Now()
	if now.Before(tournament.StartTime) {
		return fmt.Errorf("tournament cannot be started before scheduled time")
	}

	// Update tournament status
	tournament.Status = "active"
	tournament.UpdatedAt = time.Now()

	err = s.updateTournament(ctx, tournament)
	if err != nil {
		return fmt.Errorf("failed to update tournament status: %w", err)
	}

	// Schedule tournament end
	go s.scheduleTournamentEnd(ctx, tournamentID)

	log.Printf("Started tournament: %s", tournamentID)
	return nil
}

// SubmitTournamentResult submits a result for a tournament participant
func (s *TournamentService) SubmitTournamentResult(ctx context.Context, tournamentID, userID, sessionID string, score int) error {
	// Get tournament and participant
	tournament, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	participant, err := s.getParticipant(ctx, tournamentID, userID)
	if err != nil {
		return fmt.Errorf("participant not found: %w", err)
	}

	// Validate tournament is active
	if tournament.Status != "active" {
		return fmt.Errorf("tournament is not active")
	}

	// Check if tournament has ended
	if time.Now().After(tournament.EndTime) {
		return fmt.Errorf("tournament has ended")
	}

	// Update participant result
	participant.SessionID = sessionID
	participant.Score = score
	participant.CompletedAt = &time.Time{}
	*participant.CompletedAt = time.Now()
	participant.Status = "completed"
	participant.UpdatedAt = time.Now()

	err = s.updateParticipant(ctx, participant)
	if err != nil {
		return fmt.Errorf("failed to update participant result: %w", err)
	}

	// Update leaderboard immediately for this tournament
	err = s.updateTournamentLeaderboard(ctx, tournamentID)
	if err != nil {
		log.Printf("Failed to update tournament leaderboard: %v", err)
	}

	log.Printf("Submitted result for user %s in tournament %s: score %d", userID, tournamentID, score)
	return nil
}

// EndTournament ends a tournament and calculates final results
func (s *TournamentService) EndTournament(ctx context.Context, tournamentID string) error {
	tournament, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Update tournament status
	tournament.Status = "completed"
	tournament.UpdatedAt = time.Now()

	// Calculate final rankings
	err = s.calculateFinalRankings(ctx, tournament)
	if err != nil {
		log.Printf("Failed to calculate final rankings: %v", err)
	}

	err = s.updateTournament(ctx, tournament)
	if err != nil {
		return fmt.Errorf("failed to update tournament: %w", err)
	}

	// Award prizes and badges
	err = s.awardPrizesAndBadges(ctx, tournament)
	if err != nil {
		log.Printf("Failed to award prizes and badges: %v", err)
	}

	log.Printf("Ended tournament: %s", tournamentID)
	return nil
}

// GetTournament retrieves tournament information
func (s *TournamentService) GetTournament(ctx context.Context, tournamentID string) (*models.Tournament, error) {
	return s.getTournament(ctx, tournamentID)
}

// GetTournamentLeaderboard retrieves the leaderboard for a specific tournament
func (s *TournamentService) GetTournamentLeaderboard(ctx context.Context, tournamentID string) (*TournamentLeaderboard, error) {
	tournament, err := s.getTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found: %w", err)
	}

	// Get participants and their scores
	participants, err := s.getTournamentParticipants(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	// Sort by score (highest first)
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Score > participants[j].Score
	})

	// Convert to leaderboard entries
	var entries []LeaderboardEntry
	for i, participant := range participants {
		if participant.Status == "completed" && participant.Score > 0 {
			entry := LeaderboardEntry{
				UserID:     participant.UserID,
				Username:   s.getUsername(participant.UserID), // TODO: Implement user lookup
				Score:      participant.Score,
				Rank:       i + 1,
				IsVerified: participant.VerificationStatus == "verified",
			}
			entries = append(entries, entry)
		}
	}

	return &TournamentLeaderboard{
		TournamentID: tournamentID,
		Title:        tournament.Title,
		Entries:      entries,
		Status:       tournament.Status,
		EndTime:      tournament.EndTime,
	}, nil
}

// GetUserTournaments retrieves tournaments for a specific user
func (s *TournamentService) GetUserTournaments(ctx context.Context, userID string, status string) ([]*models.Tournament, error) {
	// Query tournaments where user is a participant
	query := datastore.NewQuery("TournamentParticipant").
		FilterField("UserID", "=", userID)

	var participants []*models.TournamentParticipant
	_, err := s.datastoreClient.GetAll(ctx, query, &participants)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tournaments: %w", err)
	}

	var tournaments []*models.Tournament
	for _, participant := range participants {
		tournament, err := s.getTournament(ctx, participant.TournamentID)
		if err != nil {
			continue // Skip if tournament not found
		}

		// Filter by status if specified
		if status != "" && tournament.Status != status {
			continue
		}

		tournaments = append(tournaments, tournament)
	}

	return tournaments, nil
}

// Helper methods

func (s *TournamentService) validateTournament(tournament *models.Tournament) error {
	if tournament.Title == "" {
		return fmt.Errorf("tournament title is required")
	}

	if tournament.LanguageID == "" {
		return fmt.Errorf("language ID is required")
	}

	if tournament.StartTime.After(tournament.EndTime) {
		return fmt.Errorf("start time must be before end time")
	}

	if tournament.RegistrationStart.After(tournament.RegistrationEnd) {
		return fmt.Errorf("registration start must be before registration end")
	}

	if tournament.MaxParticipants <= 0 {
		return fmt.Errorf("max participants must be greater than 0")
	}

	return nil
}

func (s *TournamentService) scheduleTournamentEnd(ctx context.Context, tournamentID string) {
	tournament, _ := s.getTournament(ctx, tournamentID)
	if tournament == nil {
		return
	}

	// Calculate duration until end time
	duration := time.Until(tournament.EndTime)
	if duration <= 0 {
		s.EndTournament(ctx, tournamentID)
		return
	}

	// Schedule end
	time.AfterFunc(duration, func() {
		s.EndTournament(context.Background(), tournamentID)
	})
}

func (s *TournamentService) calculateFinalRankings(ctx context.Context, tournament *models.Tournament) error {
	participants, err := s.getTournamentParticipants(ctx, tournament.ID)
	if err != nil {
		return err
	}

	// Sort by score
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Score > participants[j].Score
	})

	// Update rankings
	rank := 1
	for _, participant := range participants {
		if participant.Status == "completed" && participant.Score > 0 {
			participant.Rank = rank
			participant.UpdatedAt = time.Now()
			s.updateParticipant(ctx, participant)
			rank++
		}
	}

	// Update tournament final rankings
	tournament.FinalRankings = make([]string, 0, len(participants))
	for _, participant := range participants {
		if participant.Status == "completed" && participant.Score > 0 {
			tournament.FinalRankings = append(tournament.FinalRankings, participant.UserID)
		}
	}

	return nil
}

func (s *TournamentService) awardPrizesAndBadges(ctx context.Context, tournament *models.Tournament) error {
	// TODO: Implement prize distribution and badge awarding
	// This would integrate with payment systems and user badge system
	return nil
}

func (s *TournamentService) updateTournamentLeaderboard(ctx context.Context, tournamentID string) error {
	// Get current tournament results
	leaderboard, err := s.GetTournamentLeaderboard(ctx, tournamentID)
	if err != nil {
		return err
	}

	// TODO: Update real-time tournament leaderboard (could use pub/sub or websockets)
	log.Printf("Updated tournament leaderboard for %s: %d participants", tournamentID, len(leaderboard.Entries))

	return nil
}

// Database operations

func (s *TournamentService) getTournament(ctx context.Context, tournamentID string) (*models.Tournament, error) {
	key := datastore.NameKey("Tournament", tournamentID, nil)
	var tournament models.Tournament
	err := s.datastoreClient.Get(ctx, key, &tournament)
	if err != nil {
		return nil, err
	}
	return &tournament, nil
}

func (s *TournamentService) updateTournament(ctx context.Context, tournament *models.Tournament) error {
	key := datastore.NameKey("Tournament", tournament.ID, nil)
	_, err := s.datastoreClient.Put(ctx, key, tournament)
	return err
}

func (s *TournamentService) getParticipant(ctx context.Context, tournamentID, userID string) (*models.TournamentParticipant, error) {
	query := datastore.NewQuery("TournamentParticipant").
		FilterField("TournamentID", "=", tournamentID).
		FilterField("UserID", "=", userID)

	var participants []*models.TournamentParticipant
	_, err := s.datastoreClient.GetAll(ctx, query, &participants)
	if err != nil {
		return nil, err
	}

	if len(participants) == 0 {
		return nil, nil
	}

	return participants[0], nil
}

func (s *TournamentService) storeParticipant(ctx context.Context, participant *models.TournamentParticipant) error {
	key := datastore.NameKey("TournamentParticipant",
		fmt.Sprintf("%s_%s", participant.TournamentID, participant.UserID), nil)
	_, err := s.datastoreClient.Put(ctx, key, participant)
	return err
}

func (s *TournamentService) updateParticipant(ctx context.Context, participant *models.TournamentParticipant) error {
	key := datastore.NameKey("TournamentParticipant",
		fmt.Sprintf("%s_%s", participant.TournamentID, participant.UserID), nil)
	_, err := s.datastoreClient.Put(ctx, key, participant)
	return err
}

func (s *TournamentService) getTournamentParticipants(ctx context.Context, tournamentID string) ([]*models.TournamentParticipant, error) {
	query := datastore.NewQuery("TournamentParticipant").
		FilterField("TournamentID", "=", tournamentID)

	var participants []*models.TournamentParticipant
	_, err := s.datastoreClient.GetAll(ctx, query, &participants)
	return participants, err
}

func (s *TournamentService) getUsername(userID string) string {
	// TODO: Implement user lookup
	return fmt.Sprintf("user_%s", userID[:8])
}

// Request/Response types

type CreateTournamentRequest struct {
	Title                string             `json:"title"`
	Description          string             `json:"description"`
	LanguageID           string             `json:"language_id"`
	Mode                 string             `json:"mode"`
	VerificationRequired bool               `json:"verification_required"`
	VerificationMethod   string             `json:"verification_method"`
	AntiCheatLevel       string             `json:"anti_cheat_level"`
	RegistrationStart    time.Time          `json:"registration_start"`
	RegistrationEnd      time.Time          `json:"registration_end"`
	StartTime            time.Time          `json:"start_time"`
	EndTime              time.Time          `json:"end_time"`
	DurationMinutes      int                `json:"duration_minutes"`
	MaxParticipants      int                `json:"max_participants"`
	PrizePool            float64            `json:"prize_pool"`
	PrizeDistribution    map[string]float64 `json:"prize_distribution"`
	BadgeReward          string             `json:"badge_reward"`
	CreatedBy            string             `json:"created_by"`
}
