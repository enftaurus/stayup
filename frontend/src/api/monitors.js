import { apiFetch } from './client';

export const listMonitors = async (status = '') => {
  const query = status ? `?status=${encodeURIComponent(status)}` : '';
  const res = await apiFetch(`/api/monitors${query}`);
  if (!res.ok) throw new Error('Failed to fetch monitors');
  const data = await res.json();
  return data.monitors;
};

export const getMonitor = async (id) => {
  const res = await apiFetch(`/api/monitors/${id}`);
  if (!res.ok) throw new Error('Failed to fetch monitor details');
  return res.json();
};

export const createMonitor = async (payload) => {
  const res = await apiFetch('/api/monitors', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw errorData;
  }
  return res.json();
};

export const updateMonitor = async (id, payload) => {
  const res = await apiFetch(`/api/monitors/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw errorData;
  }
  return res.json();
};

export const deleteMonitor = async (id) => {
  const res = await apiFetch(`/api/monitors/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete monitor');
};

export const pauseMonitor = async (id) => {
  const res = await apiFetch(`/api/monitors/${id}/pause`, { method: 'POST' });
  if (!res.ok) throw new Error('Failed to pause monitor');
  return res.json();
};

export const resumeMonitor = async (id) => {
  const res = await apiFetch(`/api/monitors/${id}/resume`, { method: 'POST' });
  if (!res.ok) throw new Error('Failed to resume monitor');
  return res.json();
};

export const getMonitorChecks = async (id, range = '24h') => {
  const res = await apiFetch(`/api/monitors/${id}/checks?range=${range}`);
  if (!res.ok) throw new Error('Failed to fetch check buckets');
  return res.json();
};

export const getRecentChecks = async (id, limit = 50) => {
  const res = await apiFetch(`/api/monitors/${id}/checks/recent?limit=${limit}`);
  if (!res.ok) throw new Error('Failed to fetch recent checks');
  return res.json();
};

export const getMonitorIncidents = async (id, cursor = '') => {
  const query = cursor ? `?cursor=${encodeURIComponent(cursor)}` : '';
  const res = await apiFetch(`/api/monitors/${id}/incidents${query}`);
  if (!res.ok) throw new Error('Failed to fetch incidents');
  return res.json();
};
