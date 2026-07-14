import React, { createContext, useContext, useState, ReactNode } from 'react';
import { authService, User } from './AuthService';

export interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isAnonymous: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (handle: string, email: string, password: string) => Promise<void>;
  anonymousLogin: (
    deviceId: string,
    keyboardLayout: string,
    locale: string
  ) => Promise<void>;
  logout: () => void;
  loading: boolean;

  // MFA methods
  loginWithMFA: (
    email: string,
    password: string,
    code: string
  ) => Promise<void>;
  setupMFA: () => Promise<any>;
  verifyMFASetup: (code: string) => Promise<void>;
  getMFAStatus: () => Promise<any>;
  disableMFA: (password: string, code: string) => Promise<void>;
  regenerateMFABackupCodes: (code: string) => Promise<any>;
}

interface AuthProviderProps {
  children: ReactNode;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setCurrentUser] = useState<User | null>(
    authService.getCurrentUser()
  );
  const [loading, setLoading] = useState(false);

  const login = async (email: string, password: string) => {
    setLoading(true);
    try {
      await authService.login({ email, password });
      setCurrentUser(authService.getCurrentUser());
    } finally {
      setLoading(false);
    }
  };

  const register = async (handle: string, email: string, password: string) => {
    setLoading(true);
    try {
      await authService.register({ handle, email, password });
      setCurrentUser(authService.getCurrentUser());
    } finally {
      setLoading(false);
    }
  };

  const anonymousLogin = async (
    deviceId: string,
    keyboardLayout: string,
    locale: string
  ) => {
    setLoading(true);
    try {
      await authService.createAnonymousSession({
        device_id: deviceId,
        keyboard_layout: keyboardLayout,
        locale,
      });
      setCurrentUser(authService.getCurrentUser());
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    setLoading(true);
    try {
      await authService.logout();
      setCurrentUser(null);
    } finally {
      setLoading(false);
    }
  };

  const loginWithMFA = async (
    email: string,
    password: string,
    code: string
  ) => {
    setLoading(true);
    try {
      await authService.loginWithMFA({ email, password, code });
      setCurrentUser(authService.getCurrentUser());
    } finally {
      setLoading(false);
    }
  };

  const setupMFA = async () => {
    setLoading(true);
    try {
      return await authService.setupMFA();
    } finally {
      setLoading(false);
    }
  };

  const verifyMFASetup = async (code: string) => {
    setLoading(true);
    try {
      await authService.verifyMFASetup(code);
      setCurrentUser(authService.getCurrentUser());
    } finally {
      setLoading(false);
    }
  };

  const getMFAStatus = async () => {
    return await authService.getMFAStatus();
  };

  const disableMFA = async (password: string, code: string) => {
    setLoading(true);
    try {
      await authService.disableMFA(password, code);
      setCurrentUser(authService.getCurrentUser());
    } finally {
      setLoading(false);
    }
  };

  const regenerateMFABackupCodes = async (code: string) => {
    return await authService.regenerateMFABackupCodes(code);
  };

  const value: AuthContextType = {
    user,
    token: authService.getToken(),
    isAuthenticated: authService.isAuthenticated(),
    isAnonymous: authService.isAnonymous(),
    login,
    register,
    anonymousLogin,
    logout,
    loading,

    // MFA methods
    loginWithMFA,
    setupMFA,
    verifyMFASetup,
    getMFAStatus,
    disableMFA,
    regenerateMFABackupCodes,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
