import type { ErrorResponse } from "./types";

export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "") ||
  "http://localhost:8080";

export const USE_MOCKS =
  String(import.meta.env.VITE_USE_MOCKS ?? "false").toLowerCase() === "true";

export const ACCESS_TOKEN_KEY = "goalmind.accessToken";

export class ApiClientError extends Error {
  status: number;
  code?: string;
  details?: Record<string, unknown> | null;

  constructor(
    message: string,
    status: number,
    code?: string,
    details?: Record<string, unknown> | null
  ) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

export function getAccessToken() {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function setAccessToken(token: string) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(ACCESS_TOKEN_KEY, token);
  window.dispatchEvent(new CustomEvent("auth:token-updated"));
}

export function clearAccessToken() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(ACCESS_TOKEN_KEY);
  window.dispatchEvent(new CustomEvent("auth:logout"));
}

export function buildApiUrl(path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${API_BASE_URL}${normalizedPath}`;
}

function redirectToLogin() {
  if (typeof window === "undefined") {
    return;
  }

  const publicPaths = ["/login", "/register", "/auth/callback"];
  if (publicPaths.includes(window.location.pathname)) {
    return;
  }

  window.location.replace("/login");
}

async function parseResponseBody(response: Response) {
  const text = await response.text();
  if (!text) {
    return null;
  }

  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

export async function apiRequest<T>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Accept", "application/json");

  if (init?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(buildApiUrl(path), {
    ...init,
    headers
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const body = await parseResponseBody(response);

  if (!response.ok) {
    const errorPayload = body as ErrorResponse | null;

    if (response.status === 401 && errorPayload?.error.code === "unauthorized") {
      clearAccessToken();
      redirectToLogin();
    }

    throw new ApiClientError(
      errorPayload?.error.message || `Request failed with status ${response.status}`,
      response.status,
      errorPayload?.error.code,
      errorPayload?.error.details
    );
  }

  return body as T;
}
