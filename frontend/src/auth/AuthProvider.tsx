import {
  createContext,
  useEffect,
  useState,
  type PropsWithChildren
} from "react";
import { authApi } from "../api/endpoints";
import {
  clearAccessToken,
  getAccessToken,
  setAccessToken
} from "../api/client";
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  User
} from "../api/types";

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (payload: LoginRequest) => Promise<void>;
  register: (payload: RegisterRequest) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
  authenticateWithToken: (token: string) => Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | undefined>(
  undefined
);

export function AuthProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  function applyAuthResponse(response: AuthResponse) {
    setAccessToken(response.accessToken);
    setUser(response.user);
  }

  async function refreshUser() {
    const token = getAccessToken();

    if (!token) {
      setUser(null);
      setIsLoading(false);
      return;
    }

    try {
      const currentUser = await authApi.getMe();
      setUser(currentUser);
    } catch {
      clearAccessToken();
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }

  async function authenticateWithToken(token: string) {
    setAccessToken(token);
    setIsLoading(true);

    try {
      const currentUser = await authApi.getMe();
      setUser(currentUser);
    } catch {
      clearAccessToken();
      setUser(null);
      throw new Error("Не удалось создать сессию после OAuth-ответа.");
    } finally {
      setIsLoading(false);
    }
  }

  async function login(payload: LoginRequest) {
    const response = await authApi.login(payload);
    applyAuthResponse(response);
  }

  async function register(payload: RegisterRequest) {
    const response = await authApi.register(payload);
    applyAuthResponse(response);
  }

  function logout() {
    clearAccessToken();
    setUser(null);
  }

  useEffect(() => {
    void refreshUser();
  }, []);

  useEffect(() => {
    function handleLogout() {
      setUser(null);
    }

    window.addEventListener("auth:logout", handleLogout);
    return () => {
      window.removeEventListener("auth:logout", handleLogout);
    };
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: Boolean(user),
        isLoading,
        login,
        register,
        logout,
        refreshUser,
        authenticateWithToken
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
