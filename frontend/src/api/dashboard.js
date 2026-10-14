import { apiFetch } from './client';

export const getDashboardSummary = async () => {
  const res = await apiFetch('/api/dashboard/summary');
  if (!res.ok) throw new Error('Failed to fetch dashboard summary');
  return res.json();
};
