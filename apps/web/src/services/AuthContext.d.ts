import React from 'react';

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

export interface AuthContextType {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  isAnonymous: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (handle: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
}

export const AuthProvider: React.FC<{ children: React.ReactNode }>;
export const useAuth: () => AuthContextType;
