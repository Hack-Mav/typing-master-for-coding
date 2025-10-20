import { authService } from './AuthService';

/**
 * LeaderboardService - Handles leaderboard and tournament functionality
 */
export interface LeaderboardEntry {
  userId: string;
  handle: string;
  score: number;
  rank: number;
  avatar?: string;
  country?: string;
  lastActive: Date;
}

export interface Tournament {
  id: string;
  name: string;
  description: string;
  startDate: Date;
  endDate: Date;
  status: 'upcoming' | 'active' | 'completed';
  participantCount: number;
  maxParticipants?: number;
  prizePool?: number;
  rules: string[];
  language?: string;
  difficulty?: number;
}

export interface TournamentParticipant {
  userId: string;
  handle: string;
  joinedAt: Date;
  score: number;
  rank: number;
}

export interface TournamentResult {
  tournamentId: string;
  userId: string;
  finalScore: number;
  finalRank: number;
  completedAt: Date;
  prize?: number;
}

class LeaderboardService {
  private apiBaseUrl: string;

  constructor() {
    this.apiBaseUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
  }

  /**
   * Get global leaderboard
   */
  async getLeaderboard(
    limit: number = 50,
    timeRange: 'daily' | 'weekly' | 'monthly' | 'allTime' = 'allTime'
  ): Promise<LeaderboardEntry[]> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/leaderboards?limit=${limit}&timeRange=${timeRange}`,
        {
          headers: authService.getAuthHeader(),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch leaderboard: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching leaderboard:', error);
      return [];
    }
  }

  /**
   * Get user's rank in leaderboard
   */
  async getUserRank(userId: string): Promise<number> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/leaderboards/rank/${userId}`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch user rank: ${response.statusText}`);
      }

      const data = await response.json();
      return data.rank;
    } catch (error) {
      console.error('Error fetching user rank:', error);
      return 0;
    }
  }

  /**
   * Get leaderboard trends over time
   */
  async getLeaderboardTrends(
    userId: string,
    days: number = 30
  ): Promise<{ date: string; rank: number; score: number }[]> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/leaderboards/trends?userId=${userId}&days=${days}`,
        {
          headers: authService.getAuthHeader(),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch leaderboard trends: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching leaderboard trends:', error);
      return [];
    }
  }

  /**
   * Get available tournaments
   */
  async getTournaments(status?: 'upcoming' | 'active' | 'completed'): Promise<Tournament[]> {
    try {
      const url = status
        ? `${this.apiBaseUrl}/tournaments?status=${status}`
        : `${this.apiBaseUrl}/tournaments`;

      const response = await fetch(url, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch tournaments: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching tournaments:', error);
      return [];
    }
  }

  /**
   * Get specific tournament details
   */
  async getTournament(tournamentId: string): Promise<Tournament | null> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments/${tournamentId}`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        if (response.status === 404) {
          return null;
        }
        throw new Error(`Failed to fetch tournament: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching tournament:', error);
      return null;
    }
  }

  /**
   * Register for a tournament
   */
  async registerForTournament(tournamentId: string): Promise<void> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments/${tournamentId}/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to register for tournament: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error registering for tournament:', error);
      throw error;
    }
  }

  /**
   * Get tournament leaderboard
   */
  async getTournamentLeaderboard(tournamentId: string): Promise<TournamentParticipant[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments/${tournamentId}/leaderboard`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch tournament leaderboard: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching tournament leaderboard:', error);
      return [];
    }
  }

  /**
   * Submit tournament result
   */
  async submitTournamentResult(
    tournamentId: string,
    sessionId: string,
    score: number
  ): Promise<void> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments/${tournamentId}/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify({
          sessionId,
          score,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to submit tournament result: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error submitting tournament result:', error);
      throw error;
    }
  }

  /**
   * Get user's tournaments
   */
  async getUserTournaments(userId: string): Promise<Tournament[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments/user/${userId}`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch user tournaments: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching user tournaments:', error);
      return [];
    }
  }

  /**
   * Create a new tournament (admin only)
   */
  async createTournament(tournament: Omit<Tournament, 'id' | 'participantCount'>): Promise<Tournament> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify(tournament),
      });

      if (!response.ok) {
        throw new Error(`Failed to create tournament: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error creating tournament:', error);
      throw error;
    }
  }

  /**
   * Start a tournament (admin only)
   */
  async startTournament(tournamentId: string): Promise<void> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments/${tournamentId}/start`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to start tournament: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error starting tournament:', error);
      throw error;
    }
  }

  /**
   * End a tournament (admin only)
   */
  async endTournament(tournamentId: string): Promise<void> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/tournaments/${tournamentId}/end`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to end tournament: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error ending tournament:', error);
      throw error;
    }
  }
}

// Export singleton instance
export const leaderboardService = new LeaderboardService();
export default leaderboardService;
