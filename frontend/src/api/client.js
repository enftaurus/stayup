// API Client managing in-memory access token and automatic refresh on 401

let inMemoryToken = null;
let refreshPromise = null;

export const setAccessToken = (token) => {
  inMemoryToken = token;
};

export const getAccessToken = () => {
  return inMemoryToken;
};

export const refreshSession = async () => {
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    try {
      const res = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      });

      if (!res.ok) {
        inMemoryToken = null;
        return null;
      }

      const data = await res.json();
      inMemoryToken = data.access_token;
      return data;
    } catch (err) {
      inMemoryToken = null;
      return null;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
};

export const apiFetch = async (url, options = {}) => {
  options.headers = options.headers || {};
  if (inMemoryToken) {
    options.headers['Authorization'] = `Bearer ${inMemoryToken}`;
  }
  options.credentials = 'include';

  let res = await fetch(url, options);

  // If 401 unauthorized, attempt single refresh
  if (res.status === 401) {
    const refreshed = await refreshSession();
    if (refreshed && refreshed.access_token) {
      options.headers['Authorization'] = `Bearer ${refreshed.access_token}`;
      res = await fetch(url, options);
    }
  }

  return res;
};
