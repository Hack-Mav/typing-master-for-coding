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

  // MFA fields
  mfaEnabled?: boolean;
  mfaSetupAt?: string;
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
  requires_mfa?: boolean;
  user_id?: string;
  message?: string;
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

export interface LoginWithMFARequest {
  email: string;
  password: string;
  code: string;
  rememberDevice?: boolean;
}

export interface MFASetupRequest {
  // No additional fields needed
}

export interface MFASetupResponse {
  secret: string;
  qr_code_url: string;
  backup_codes: string[];
}

export interface MFAVerifyRequest {
  code: string;
}

export interface MFAVerifyResponse {
  success: boolean;
  message: string;
}

export interface MFAStatusResponse {
  enabled: boolean;
  setup_at?: string;
  has_backup_codes: boolean;
}

export interface MFADisableRequest {
  password: string;
  code: string;
}

export interface AnonymousSessionRequest {
  device_id: string;
  keyboard_layout: string;
  locale: string;
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

    // Check if MFA is required
    if (data.requires_mfa) {
      return data; // Return early with MFA requirement info
    }

    // MFA not required - complete login
    this.accessToken = data.tokens.access_token;
    this.refreshToken = data.tokens.refresh_token;
    this.currentUser = data.user;

    this.saveToStorage(data.user, data.tokens);
    this.scheduleTokenRefresh();

    return data;
  }

  /**
   * Create an anonymous session (handled entirely in frontend)
   */
  async createAnonymousSession(
    request: AnonymousSessionRequest
  ): Promise<AuthResponse> {
    // Generate anonymous user ID based on device ID
    const anonymousID = `anon_${request.device_id}`;
    
    // Create mock user object
    const anonymousUser: User = {
      id: anonymousID,
      handle: 'Anonymous',
      email: '',
      isAnonymous: true,
      locale: request.locale,
      keyboardLayout: request.keyboard_layout,
      privacyMode: true, // Anonymous users get privacy mode by default
      telemetryConsent: false,
      dataProcessingConsent: false,
      settings: {},
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    // Generate mock tokens (for frontend-only use)
    const mockTokens: TokenPair = {
      access_token: `anon_token_${anonymousID}_${Date.now()}`,
      refresh_token: `anon_refresh_${anonymousID}_${Date.now()}`,
      expires_in: 86400, // 24 hours
      token_type: 'Bearer',
    };

    // Set current user and tokens
    this.currentUser = anonymousUser;
    this.accessToken = mockTokens.access_token;
    this.refreshToken = mockTokens.refresh_token;

    // Save to localStorage for persistence
    this.saveToStorage(anonymousUser, mockTokens);
    
    // No need to schedule token refresh for anonymous users
    if (this.tokenRefreshTimeout) {
      clearTimeout(this.tokenRefreshTimeout);
      this.tokenRefreshTimeout = null;
    }

    return {
      user: anonymousUser,
      tokens: mockTokens,
    };
  }

  /**
   * Refresh the access token using the refresh token
   */
  async refreshAccessToken(): Promise<TokenPair> {
    if (!this.refreshToken) {
      throw new Error('No refresh token available');
    }

    // Anonymous users don't need to refresh tokens
    if (this.currentUser?.isAnonymous) {
      // Just return the current mock tokens
      return {
        access_token: this.accessToken!,
        refresh_token: this.refreshToken,
        expires_in: 86400,
        token_type: 'Bearer',
      };
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
    // Anonymous users don't make real API requests
    if (this.currentUser?.isAnonymous) {
      throw new Error('Anonymous users cannot access server-side features');
    }

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
    // For anonymous users, return the stored profile
    if (this.currentUser?.isAnonymous) {
      return this.currentUser;
    }

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
    // Anonymous users cannot update profiles
    if (this.currentUser?.isAnonymous) {
      throw new Error('Anonymous users cannot update profiles');
    }

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

  /**
   * Complete login with MFA verification
   */
  async loginWithMFA(request: LoginWithMFARequest): Promise<AuthResponse> {
    const response = await fetch(`${this.apiBaseUrl}/auth/login/mfa`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'MFA verification failed');
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
   * Setup MFA for the current user
   */
  async setupMFA(): Promise<MFASetupResponse> {
    const response = await this.authenticatedRequest<MFASetupResponse>(
      '/auth/mfa/setup',
      {
        method: 'POST',
        body: JSON.stringify({}),
      }
    );

    return response;
  }

  /**
   * Verify MFA setup with TOTP code
   */
  async verifyMFASetup(code: string): Promise<MFAVerifyResponse> {
    const response = await this.authenticatedRequest<MFAVerifyResponse>(
      '/auth/mfa/verify-setup',
      {
        method: 'POST',
        body: JSON.stringify({ code }),
      }
    );

    return response;
  }

  /**
   * Get MFA status for the current user
   */
  async getMFAStatus(): Promise<MFAStatusResponse> {
    const response = await this.authenticatedRequest<MFAStatusResponse>(
      '/mfa/status',
      {
        method: 'GET',
      }
    );

    return response;
  }

  /**
   * Disable MFA for the current user
   */
  async disableMFA(
    password: string,
    code: string
  ): Promise<{ message: string }> {
    const response = await this.authenticatedRequest<{ message: string }>(
      '/auth/mfa/disable',
      {
        method: 'POST',
        body: JSON.stringify({ password, code }),
      }
    );

    return response;
  }

  /**
   * Regenerate MFA backup codes
   */
  async regenerateMFABackupCodes(
    code: string
  ): Promise<{ backup_codes: string[]; message: string }> {
    const response = await this.authenticatedRequest<{
      backup_codes: string[];
      message: string;
    }>('/auth/mfa/backup-codes/regenerate', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });

    return response;
  }
}

// Export singleton instance
export const authService = new AuthService();
export default authService;
