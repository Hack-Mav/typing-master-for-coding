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

  const value: AuthContextType = {
    token: authService.getAccessToken(),
    user,
    isAuthenticated: authService.isAuthenticated(),
    isAnonymous: authService.isAnonymous(),
    login,
    register,
    logout,
    loading,
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
