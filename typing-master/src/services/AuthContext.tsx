import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from 'react';
import { authService, User } from './AuthService';

export interface AuthContextType {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  isAnonymous: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (handle: string, email: string, password: string) => Promise<void>;
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

  useEffect(() => {
    // Listen for auth state changes (e.g., when tokens are refreshed)
    const handleStorageChange = () => {
      setCurrentUser(authService.getCurrentUser());
    };

    // Check for auth state changes periodically
    const interval = setInterval(handleStorageChange, 1000);

    return () => clearInterval(interval);
  }, []);

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

  const logout = () => {
    authService.logout();
    setCurrentUser(null);
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
    token: authService.getAccessToken(),
    user,
    isAuthenticated: authService.isAuthenticated(),
    isAnonymous: authService.isAnonymous(),
    login,
    register,
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
