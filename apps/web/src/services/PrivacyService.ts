/**
 * PrivacyService - Handles privacy settings, consent management, and GDPR compliance
 */

import { authService } from './AuthService';

export interface ConsentStatus {
  isAnonymous: boolean;
  privacyMode: boolean;
  telemetryConsent: boolean;
  dataProcessingConsent: boolean;
}

export interface PrivacySettings {
  privacyMode: boolean;
  telemetryConsent: boolean;
  dataProcessingConsent: boolean;
}

export interface DataExportOptions {
  format: 'json' | 'csv';
}

export interface TelemetryEvent {
  eventType: string;
  timestamp: number;
  data: Record<string, any>;
}

class PrivacyService {
  private telemetryQueue: TelemetryEvent[] = [];
  private telemetryEnabled: boolean = false;
  private readonly MAX_QUEUE_SIZE = 100;

  /**
   * Get current consent status
   */
  async getConsentStatus(): Promise<ConsentStatus> {
    return await authService.authenticatedRequest<ConsentStatus>(
      '/privacy/consent',
      { method: 'GET' }
    );
  }

  /**
   * Update privacy settings
   */
  async updatePrivacySettings(
    settings: PrivacySettings
  ): Promise<PrivacySettings> {
    const response = await authService.authenticatedRequest<PrivacySettings>(
      '/privacy/settings',
      {
        method: 'PUT',
        body: JSON.stringify(settings),
      }
    );

    // Update local telemetry state
    this.telemetryEnabled = settings.telemetryConsent;

    // Clear telemetry queue if consent is revoked
    if (!settings.telemetryConsent) {
      this.telemetryQueue = [];
    }

    return response;
  }

  /**
   * Export user data (GDPR compliance)
   */
  async exportUserData(
    options: DataExportOptions = { format: 'json' }
  ): Promise<Blob> {
    const response = await fetch(
      `${authService['apiBaseUrl']}/privacy/export?format=${options.format}`,
      {
        method: 'GET',
        headers: authService.getAuthHeader(),
      }
    );

    if (!response.ok) {
      throw new Error('Failed to export user data');
    }

    return await response.blob();
  }

  /**
   * Delete all user data (GDPR compliance - Right to be forgotten)
   */
  async deleteUserData(confirm: boolean = false): Promise<void> {
    if (!confirm) {
      throw new Error('Data deletion must be explicitly confirmed');
    }

    await authService.authenticatedRequest<void>('/privacy/delete', {
      method: 'POST',
      body: JSON.stringify({ confirm: true }),
    });

    // Logout after deletion
    authService.logout();
  }

  /**
   * Record a telemetry event (with consent check)
   */
  recordTelemetryEvent(eventType: string, data: Record<string, any>): void {
    // Check if telemetry is enabled
    if (!this.telemetryEnabled) {
      return;
    }

    // Check if user has given consent
    const user = authService.getCurrentUser();
    if (!user || !user.telemetryConsent) {
      return;
    }

    // Anonymize data if in privacy mode
    const anonymizedData = user.privacyMode ? this.anonymizeData(data) : data;

    const event: TelemetryEvent = {
      eventType,
      timestamp: Date.now(),
      data: anonymizedData,
    };

    this.telemetryQueue.push(event);

    // Prevent queue from growing too large
    if (this.telemetryQueue.length > this.MAX_QUEUE_SIZE) {
      this.telemetryQueue.shift();
    }

    // Auto-flush if queue is getting full
    if (this.telemetryQueue.length >= this.MAX_QUEUE_SIZE * 0.8) {
      this.flushTelemetry();
    }
  }

  /**
   * Anonymize telemetry data by removing PII
   */
  private anonymizeData(data: Record<string, any>): Record<string, any> {
    const anonymized: Record<string, any> = {};

    // Only include non-identifying metrics
    const allowedFields = [
      'mode',
      'language',
      'duration',
      'cpm',
      'wpm',
      'accuracy',
      'errorCount',
      'sessionType',
    ];

    for (const field of allowedFields) {
      if (field in data) {
        anonymized[field] = data[field];
      }
    }

    return anonymized;
  }

  /**
   * Flush telemetry queue to server
   */
  async flushTelemetry(): Promise<void> {
    if (this.telemetryQueue.length === 0) {
      return;
    }

    const eventsToSend = [...this.telemetryQueue];
    this.telemetryQueue = [];

    try {
      // Send telemetry data to server (implement endpoint as needed)
      await authService.authenticatedRequest('/telemetry/events', {
        method: 'POST',
        body: JSON.stringify({ events: eventsToSend }),
      });
    } catch (error) {
      console.error('Failed to send telemetry:', error);
      // Re-queue events on failure (up to max size)
      this.telemetryQueue = [
        ...eventsToSend.slice(-this.MAX_QUEUE_SIZE / 2),
        ...this.telemetryQueue,
      ].slice(0, this.MAX_QUEUE_SIZE);
    }
  }

  /**
   * Initialize privacy service with user consent status
   */
  async initialize(): Promise<void> {
    try {
      const user = authService.getCurrentUser();
      if (user && !user.isAnonymous) {
        const consent = await this.getConsentStatus();
        this.telemetryEnabled = consent.telemetryConsent;
      }
    } catch (error) {
      console.error('Failed to initialize privacy service:', error);
      this.telemetryEnabled = false;
    }
  }

  /**
   * Check if telemetry is enabled
   */
  isTelemetryEnabled(): boolean {
    return this.telemetryEnabled;
  }

  /**
   * Download exported data as file
   */
  async downloadUserData(format: 'json' | 'csv' = 'json'): Promise<void> {
    const blob = await this.exportUserData({ format });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `user_data_${new Date().toISOString().split('T')[0]}.${format}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  }
}

// Export singleton instance
export const privacyService = new PrivacyService();
export default privacyService;
