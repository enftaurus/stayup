import { apiFetch, setAccessToken } from './client';

export const getMe = async () => {
  const res = await apiFetch('/api/me');
  if (!res.ok) throw new Error('Failed to load user profile');
  return res.json();
};

export const logout = async () => {
  try {
    await fetch('/auth/logout', { method: 'POST', credentials: 'include' });
  } finally {
    setAccessToken(null);
  }
};

export const devLogin = async () => {
  const res = await fetch('/auth/dev-login', {
    method: 'POST',
    credentials: 'include',
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error?.message || 'Dev login failed');
  }
  const data = await res.json();
  setAccessToken(data.access_token);
  return data;
};
