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

export interface AnonymousUserResponse {
  id: string;
  handle: string;
  email: string;
  is_anonymous: boolean;
  keyboard_layout: string;
  locale: string;
  privacy_mode: boolean;
  telemetry_consent: boolean;
  data_processing_consent: boolean;
  settings: Record<string, any>;
  created_at: string;
  updated_at: string;
  role: string;
}

class AuthService {
  private apiBaseUrl: string;
  private currentUser: User | null = null;
  private accessToken: string | null = null;

  constructor() {
    this.apiBaseUrl =
      process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
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
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    const data: AuthResponse = await response.json();
    this.currentUser = data.user;
    this.accessToken = data.tokens?.access_token || null;

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
      credentials: 'include',
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
    this.currentUser = data.user;
    this.accessToken = data.tokens?.access_token || null;

    return data;
  }

  /**
   * Create an anonymous session by calling the backend API
   */
  async createAnonymousSession(
    request: AnonymousSessionRequest
  ): Promise<AuthResponse> {
    const response = await fetch(`${this.apiBaseUrl}/auth/anonymous`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create anonymous session');
    }

    const data = (await response.json()) as {
      user: AnonymousUserResponse;
      tokens: TokenPair;
      requires_mfa?: boolean;
    };

    // Convert the response to match our User interface
    const user: User = {
      id: data.user.id,
      handle: data.user.handle,
      email: data.user.email || '',
      isAnonymous: data.user.is_anonymous || true,
      locale: data.user.locale,
      keyboardLayout: data.user.keyboard_layout,
      privacyMode: data.user.privacy_mode || true,
      telemetryConsent: data.user.telemetry_consent || false,
      dataProcessingConsent: data.user.data_processing_consent || false,
      role: data.user.role || 'user',
      settings: data.user.settings || {},
      createdAt: data.user.created_at,
      updatedAt: data.user.updated_at,
      mfaEnabled: false, // Anonymous users don't have MFA
    };

    // Set current user and token
    this.currentUser = user;
    this.accessToken = data.tokens?.access_token || null;

    return {
      user: user,
      tokens: data.tokens,
      requires_mfa: false,
    } as AuthResponse;
  }

  /**
   * Logout and clear authentication state
   */
  async logout(): Promise<void> {
    const response = await fetch(`${this.apiBaseUrl}/auth/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      console.error('Logout request failed');
    }

    this.currentUser = null;
    this.accessToken = null;
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
    return this.currentUser !== null;
  }

  /**
   * Check if current user is anonymous
   */
  isAnonymous(): boolean {
    return this.currentUser?.isAnonymous ?? false;
  }

  /**
   * Get authorization headers for authenticated API requests
   */
  getAuthHeader(): Record<string, string> {
    if (!this.accessToken) {
      return {};
    }
    return { Authorization: `Bearer ${this.accessToken}` };
  }

  /**
   * Get the current access token
   */
  getToken(): string | null {
    return this.accessToken;
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
      ...options.headers,
    };

    const response = await fetch(`${this.apiBaseUrl}${endpoint}`, {
      ...options,
      headers,
      credentials: 'include',
    });

    // Handle token expiration - HttpOnly cookies are managed by browser
    if (response.status === 401) {
      await this.logout();
      throw new Error('Authentication expired. Please login again.');
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
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'MFA verification failed');
    }

    const data: AuthResponse = await response.json();
    this.currentUser = data.user;
    this.accessToken = data.tokens?.access_token || null;

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
