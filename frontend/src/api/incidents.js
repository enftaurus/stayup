import { apiFetch } from './client';

export const listIncidents = async (status = '', cursor = '') => {
  const params = new URLSearchParams();
  if (status) params.append('status', status);
  if (cursor) params.append('cursor', cursor);

  const query = params.toString() ? `?${params.toString()}` : '';
  const res = await apiFetch(`/api/incidents${query}`);
  if (!res.ok) throw new Error('Failed to fetch incidents');
  return res.json();
};
