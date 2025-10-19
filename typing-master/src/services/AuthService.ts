/**
 * AuthService - Handles user authentication, JWT token management, and anonymous sessions
 */

export interface User {
  id: string;
  handle: string;
  email: string;
  isAnonymous: boolean;
  locale: string;
  keyboardLayout: string;
  privacyMode: boolean;
  telemetryConsent: boolean;
  dataProcessingConsent: boolean;
  role?: string;
  settings: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface AuthResponse {
  user: User;
  tokens: TokenPair;
}

export interface RegisterRequest {
  handle: string;
  email: string;
  password: string;
  locale?: string;
  keyboardLayout?: string;
  telemetryConsent?: boolean;
  dataProcessingConsent?: boolean;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AnonymousSessionRequest {
  deviceId: string;
  keyboardLayout?: string;
  locale?: string;
}

class AuthService {
  private apiBaseUrl: string;
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  private currentUser: User | null = null;
  private tokenRefreshTimeout: NodeJS.Timeout | null = null;

  constructor() {
    this.apiBaseUrl =
      process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
    this.loadFromStorage();
  }

  /**
   * Load authentication state from localStorage
   */
  private loadFromStorage(): void {
    try {
      const accessToken = localStorage.getItem('access_token');
      const refreshToken = localStorage.getItem('refresh_token');
      const userJson = localStorage.getItem('user');

      if (accessToken && refreshToken && userJson) {
        this.accessToken = accessToken;
        this.refreshToken = refreshToken;
        this.currentUser = JSON.parse(userJson);
        this.scheduleTokenRefresh();
      }
    } catch (error) {
      console.error('Failed to load auth state from storage:', error);
      this.clearStorage();
    }
  }

  /**
   * Save authentication state to localStorage
   */
  private saveToStorage(user: User, tokens: TokenPair): void {
    localStorage.setItem('access_token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);
    localStorage.setItem('user', JSON.stringify(user));
  }

  /**
   * Clear authentication state from localStorage
   */
  private clearStorage(): void {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
  }

  /**
   * Schedule automatic token refresh before expiration
   */
  private scheduleTokenRefresh(): void {
    if (this.tokenRefreshTimeout) {
      clearTimeout(this.tokenRefreshTimeout);
    }

    // Refresh token 1 minute before expiration (default 15 min - 1 min = 14 min)
    const refreshTime = 14 * 60 * 1000; // 14 minutes in milliseconds

    this.tokenRefreshTimeout = setTimeout(async () => {
      try {
        await this.refreshAccessToken();
      } catch (error) {
        console.error('Failed to refresh token:', error);
        this.logout();
      }
    }, refreshTime);
  }

  /**
   * Register a new user
   */
  async register(request: RegisterRequest): Promise<AuthResponse> {
    const response = await fetch(`${this.apiBaseUrl}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    const data: AuthResponse = await response.json();
    this.accessToken = data.tokens.access_token;
    this.refreshToken = data.tokens.refresh_token;
    this.currentUser = data.user;

    this.saveToStorage(data.user, data.tokens);
    this.scheduleTokenRefresh();

    return data;
  }

  /**
   * Login with email and password
   */
  async login(request: LoginRequest): Promise<AuthResponse> {
    const response = await fetch(`${this.apiBaseUrl}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    const data: AuthResponse = await response.json();
    this.accessToken = data.tokens.access_token;
    this.refreshToken = data.tokens.refresh_token;
    this.currentUser = data.user;

    this.saveToStorage(data.user, data.tokens);
    this.scheduleTokenRefresh();

    return data;
  }

  /**
   * Create an anonymous session (no server-side storage)
   */
  async createAnonymousSession(
    request: AnonymousSessionRequest
  ): Promise<AuthResponse> {
    const response = await fetch(`${this.apiBaseUrl}/auth/anonymous`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create anonymous session');
    }

    const data: AuthResponse = await response.json();
    this.accessToken = data.tokens.access_token;
    this.refreshToken = data.tokens.refresh_token;
    this.currentUser = data.user;

    this.saveToStorage(data.user, data.tokens);
    this.scheduleTokenRefresh();

    return data;
  }

  /**
   * Refresh the access token using the refresh token
   */
  async refreshAccessToken(): Promise<TokenPair> {
    if (!this.refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await fetch(`${this.apiBaseUrl}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: this.refreshToken }),
    });

    if (!response.ok) {
      throw new Error('Failed to refresh token');
    }

    const tokens: TokenPair = await response.json();
    this.accessToken = tokens.access_token;
    this.refreshToken = tokens.refresh_token;

    if (this.currentUser) {
      this.saveToStorage(this.currentUser, tokens);
    }

    this.scheduleTokenRefresh();

    return tokens;
  }

  /**
   * Logout and clear authentication state
   */
  logout(): void {
    this.accessToken = null;
    this.refreshToken = null;
    this.currentUser = null;

    if (this.tokenRefreshTimeout) {
      clearTimeout(this.tokenRefreshTimeout);
      this.tokenRefreshTimeout = null;
    }

    this.clearStorage();
  }

  /**
   * Get the current authenticated user
   */
  getCurrentUser(): User | null {
    return this.currentUser;
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    return this.accessToken !== null && this.currentUser !== null;
  }

  /**
   * Check if current user is anonymous
   */
  isAnonymous(): boolean {
    return this.currentUser?.isAnonymous ?? false;
  }

  /**
   * Get the current access token
   */
  getAccessToken(): string | null {
    return this.accessToken;
  }

  /**
   * Get authorization header for API requests
   */
  getAuthHeader(): Record<string, string> {
    if (!this.accessToken) {
      return {};
    }
    return {
      Authorization: `Bearer ${this.accessToken}`,
    };
  }

  /**
   * Make an authenticated API request
   */
  async authenticatedRequest<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers = {
      'Content-Type': 'application/json',
      ...this.getAuthHeader(),
      ...options.headers,
    };

    const response = await fetch(`${this.apiBaseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    // Handle token expiration
    if (response.status === 401) {
      try {
        await this.refreshAccessToken();
        // Retry the request with new token
        const retryHeaders = {
          'Content-Type': 'application/json',
          ...this.getAuthHeader(),
          ...options.headers,
        };
        const retryResponse = await fetch(`${this.apiBaseUrl}${endpoint}`, {
          ...options,
          headers: retryHeaders,
        });

        if (!retryResponse.ok) {
          throw new Error('Request failed after token refresh');
        }

        return await retryResponse.json();
      } catch (error) {
        this.logout();
        throw new Error('Authentication expired. Please login again.');
      }
    }

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Request failed');
    }

    return await response.json();
  }

  /**
   * Get user profile
   */
  async getProfile(): Promise<User> {
    const user = await this.authenticatedRequest<User>('/profile', {
      method: 'GET',
    });
    this.currentUser = user;
    if (this.refreshToken && this.accessToken) {
      localStorage.setItem('user', JSON.stringify(user));
    }
    return user;
  }

  /**
   * Update user profile
   */
  async updateProfile(updates: Partial<User>): Promise<User> {
    const user = await this.authenticatedRequest<User>('/profile', {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
    this.currentUser = user;
    if (this.refreshToken && this.accessToken) {
      localStorage.setItem('user', JSON.stringify(user));
    }
    return user;
  }
}

// Export singleton instance
export const authService = new AuthService();
export default authService;
