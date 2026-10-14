import React, { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { 
  ArrowLeft, ExternalLink, Play, Pause, Edit3, Trash2, 
  Clock, ShieldAlert, CheckCircle, RefreshCw, AlertTriangle
} from 'lucide-react';
import { 
  getMonitor, updateMonitor, deleteMonitor, pauseMonitor, resumeMonitor, 
  getMonitorChecks, getRecentChecks, getMonitorIncidents 
} from '../api/monitors';
import { StatusBadge } from '../components/StatusBadge';
import { Chart } from '../components/Chart';
import { MonitorModal } from '../components/MonitorModal';
import { ConfirmModal } from '../components/ConfirmModal';
import { Skeleton } from '../components/Skeleton';

const timeAgo = (dateStr) => {
  if (!dateStr) return 'Never';
  const seconds = Math.floor((new Date() - new Date(dateStr)) / 1000);
  if (seconds < 5) return 'Just now';
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
};

export const MonitorDetailPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [range, setRange] = useState('24h');
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);

  // Queries
  const { data: monitor, isLoading: isMonitorLoading, refetch: refetchMonitor } = useQuery({
    queryKey: ['monitor', id],
    queryFn: () => getMonitor(id),
    refetchInterval: 15000,
  });

  const { data: chartData, isLoading: isChartLoading } = useQuery({
    queryKey: ['monitor-checks', id, range],
    queryFn: () => getMonitorChecks(id, range),
    refetchInterval: 15000,
  });

  const { data: recentData, isLoading: isRecentLoading } = useQuery({
    queryKey: ['monitor-recent', id],
    queryFn: () => getRecentChecks(id, 50),
    refetchInterval: 15000,
  });

  const { data: incidentsData } = useQuery({
    queryKey: ['monitor-incidents', id],
    queryFn: () => getMonitorIncidents(id),
    refetchInterval: 15000,
  });

  // Mutations
  const updateMutation = useMutation({
    mutationFn: (payload) => updateMonitor(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitor', id] });
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });

  const pauseMutation = useMutation({
    mutationFn: () => pauseMonitor(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitor', id] });
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });

  const resumeMutation = useMutation({
    mutationFn: () => resumeMonitor(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitor', id] });
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteMonitor(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
      navigate('/dashboard');
    },
  });

  if (isMonitorLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-12 w-full" />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-24 w-full rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  if (!monitor) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-16 text-center">
        <h2 className="text-xl font-bold text-slate-900">Monitor not found</h2>
        <p className="mt-2 text-sm text-slate-500">The requested monitor does not exist or was removed.</p>
        <Link to="/dashboard" className="mt-4 inline-flex items-center text-sm text-indigo-600 font-semibold">
          Return to Dashboard
        </Link>
      </div>
    );
  }

  const isPaused = !monitor.enabled || monitor.state.status === 'PAUSED';

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Top Navigation */}
      <div>
        <Link
          to="/dashboard"
          className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-500 hover:text-slate-900 transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" />
          <span>Back to Dashboard</span>
        </Link>
      </div>

      {/* Header Info & Actions */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6 bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs">
        <div className="space-y-2">
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-2xl font-bold text-slate-900 tracking-tight">{monitor.name}</h1>
            <StatusBadge status={monitor.state.status} size="md" />
          </div>
          <div className="flex items-center gap-3 text-xs text-slate-500">
            <a
              href={monitor.url}
              target="_blank"
              rel="noreferrer"
              className="hover:text-indigo-600 flex items-center gap-1 font-mono"
            >
              <span>{monitor.url}</span>
              <ExternalLink className="w-3 h-3" />
            </a>
            <span>•</span>
            <span>Check interval: {monitor.interval_seconds / 60}m</span>
            <span>•</span>
            <span>Expected status: {monitor.expected_status}</span>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2.5 flex-wrap">
          {isPaused ? (
            <button
              type="button"
              onClick={() => resumeMutation.mutate()}
              disabled={resumeMutation.isPending}
              className="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-lg bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-100 transition-colors"
            >
              <Play className="w-3.5 h-3.5" />
              <span>Resume</span>
            </button>
          ) : (
            <button
              type="button"
              onClick={() => pauseMutation.mutate()}
              disabled={pauseMutation.isPending}
              className="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-lg bg-slate-100 text-slate-700 hover:bg-slate-200 transition-colors"
            >
              <Pause className="w-3.5 h-3.5" />
              <span>Pause</span>
            </button>
          )}

          <button
            type="button"
            onClick={() => setIsEditOpen(true)}
            className="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-lg bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 transition-colors"
          >
            <Edit3 className="w-3.5 h-3.5 text-slate-400" />
            <span>Edit</span>
          </button>

          <button
            type="button"
            onClick={() => setIsDeleteOpen(true)}
            className="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-lg bg-white border border-rose-200 text-rose-600 hover:bg-rose-50 transition-colors"
          >
            <Trash2 className="w-3.5 h-3.5 text-rose-500" />
            <span>Delete</span>
          </button>
        </div>
      </div>

      {/* Stat Tiles */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        {/* Current status & last check */}
        <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
          <span className="text-xs text-slate-500">Status</span>
          <div className="mt-1">
            <StatusBadge status={monitor.state.status} size="sm" />
          </div>
          <span className="text-[11px] text-slate-400 mt-2 block">
            Last check: {timeAgo(monitor.state.last_checked_at)}
          </span>
        </div>

        {/* 24h Uptime */}
        <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
          <span className="text-xs text-slate-500">24-Hour Uptime</span>
          <div className="mt-1 text-xl font-bold font-mono text-slate-900">
            {monitor.uptime_24h != null ? `${monitor.uptime_24h}%` : '100%'}
          </div>
          <span className="text-[11px] text-slate-400 mt-1 block">Past 24 hours</span>
        </div>

        {/* 7d Uptime */}
        <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
          <span className="text-xs text-slate-500">7-Day Uptime</span>
          <div className="mt-1 text-xl font-bold font-mono text-slate-900">
            {monitor.uptime_7d != null ? `${monitor.uptime_7d}%` : '100%'}
          </div>
          <span className="text-[11px] text-slate-400 mt-1 block">Full retention</span>
        </div>

        {/* Average Latency */}
        <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
          <span className="text-xs text-slate-500">Last Latency</span>
          <div className="mt-1 text-xl font-bold font-mono text-indigo-600">
            {monitor.state.last_response_time_ms != null ? `${monitor.state.last_response_time_ms} ms` : '—'}
          </div>
          <span className="text-[11px] text-slate-400 mt-1 block">
            Code: {monitor.state.last_status_code || '—'}
          </span>
        </div>

        {/* Next Scheduled Check */}
        <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
          <span className="text-xs text-slate-500">Next Due Check</span>
          <div className="mt-1 text-base font-semibold text-slate-900 truncate">
            {monitor.state.next_check_at ? timeAgo(monitor.state.next_check_at).replace('ago', 'due') : '—'}
          </div>
          <span className="text-[11px] text-slate-400 mt-1 block">
            Consecutive: {monitor.state.status === 'UP' ? `${monitor.state.consecutive_successes} up` : `${monitor.state.consecutive_failures} down`}
          </span>
        </div>
      </div>

      {/* Chart Section */}
      <div>
        <Chart
          buckets={chartData?.buckets || []}
          range={range}
          onRangeChange={(newRange) => setRange(newRange)}
        />
      </div>

      {/* Incidents Section */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs overflow-hidden">
        <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ShieldAlert className="w-4 h-4 text-rose-500" />
            <h3 className="text-sm font-semibold text-slate-900">Incidents History</h3>
          </div>
          <span className="text-xs text-slate-400">
            {(incidentsData?.incidents || []).length} recorded
          </span>
        </div>

        {(incidentsData?.incidents || []).length === 0 ? (
          <div className="p-8 text-center text-xs text-slate-500">
            No incidents reported for this monitor. Everything has been operating smoothly!
          </div>
        ) : (
          <div className="divide-y divide-slate-100">
            {incidentsData.incidents.map((inc) => (
              <div key={inc.id} className="p-4 flex items-center justify-between text-xs">
                <div className="flex items-center gap-3">
                  <span className={`px-2 py-0.5 rounded-full font-semibold ${
                    inc.resolved_at ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700 animate-pulse'
                  }`}>
                    {inc.resolved_at ? 'Resolved' : 'Ongoing'}
                  </span>
                  <span className="text-slate-600">
                    Started {new Date(inc.started_at).toLocaleString()}
                  </span>
                </div>
                <div className="text-right text-slate-500">
                  {inc.resolved_at ? (
                    <span>Duration: {inc.duration_seconds}s</span>
                  ) : (
                    <span className="text-rose-600 font-semibold">Incident Active</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Recent 50 Checks Table */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs overflow-hidden">
        <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-indigo-500" />
            <h3 className="text-sm font-semibold text-slate-900">Recent Checks</h3>
          </div>
          <span className="text-xs text-slate-400">Latest 50 checks</span>
        </div>

        {(recentData?.checks || []).length === 0 ? (
          <div className="p-8 text-center text-xs text-slate-500">
            Waiting for first scheduled check execution (about 1 minute after creation)...
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-50/70 border-b border-slate-100 text-slate-500 font-medium">
                <tr>
                  <th className="px-6 py-3">Checked At</th>
                  <th className="px-6 py-3">Result</th>
                  <th className="px-6 py-3">HTTP Code</th>
                  <th className="px-6 py-3">Latency</th>
                  <th className="px-6 py-3">Details / Error</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 font-mono">
                {recentData.checks.map((chk) => (
                  <tr key={chk.id} className="hover:bg-slate-50/50 transition-colors">
                    <td className="px-6 py-3 text-slate-600 whitespace-nowrap">
                      {new Date(chk.checked_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                    </td>
                    <td className="px-6 py-3">
                      <span className={`inline-flex px-2 py-0.5 rounded-full text-[10px] font-sans font-semibold ${
                        chk.status === 'UP' ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'
                      }`}>
                        {chk.status}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-slate-700">
                      {chk.status_code ?? '—'}
                    </td>
                    <td className="px-6 py-3 text-slate-700">
                      {chk.response_time_ms != null ? `${chk.response_time_ms} ms` : '—'}
                    </td>
                    <td className="px-6 py-3 font-sans text-slate-500 truncate max-w-xs">
                      {chk.error_message || 'OK'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Edit Modal */}
      <MonitorModal
        isOpen={isEditOpen}
        onClose={() => setIsEditOpen(false)}
        onSubmit={async (payload) => {
          await updateMutation.mutateAsync(payload);
        }}
        initialData={monitor}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        onConfirm={() => deleteMutation.mutate()}
        title="Delete Monitor"
        message={`Are you sure you want to delete "${monitor.name}"? All associated check history and incident logs will be permanently deleted.`}
        confirmText="Delete Monitor"
      />
    </div>
  );
};
