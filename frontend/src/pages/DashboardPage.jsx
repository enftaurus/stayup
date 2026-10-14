import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Plus, Activity, CheckCircle2, AlertCircle, PauseCircle, 
  HelpCircle, ExternalLink, ArrowUpRight, Clock, RefreshCw, AlertTriangle
} from 'lucide-react';
import { listMonitors, createMonitor } from '../api/monitors';
import { getDashboardSummary } from '../api/dashboard';
import { StatusBadge } from '../components/StatusBadge';
import { Sparkline } from '../components/Sparkline';
import { MonitorModal } from '../components/MonitorModal';
import { CardSkeleton, MonitorRowSkeleton } from '../components/Skeleton';

const timeAgo = (dateStr) => {
  if (!dateStr) return 'Never checked';
  const seconds = Math.floor((new Date() - new Date(dateStr)) / 1000);
  if (seconds < 5) return 'Just now';
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
};

export const DashboardPage = () => {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [isAddOpen, setIsAddOpen] = useState(false);

  // Poll summary and monitors every 15 seconds
  const { data: summary, isLoading: isSummaryLoading, refetch: refetchSummary } = useQuery({
    queryKey: ['dashboard-summary'],
    queryFn: getDashboardSummary,
    refetchInterval: 15000,
  });

  const { data: monitors = [], isLoading: isMonitorsLoading, isFetching, refetch: refetchMonitors } = useQuery({
    queryKey: ['monitors', statusFilter],
    queryFn: () => listMonitors(statusFilter),
    refetchInterval: 15000,
  });

  const createMutation = useMutation({
    mutationFn: createMonitor,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard-summary'] });
    },
  });

  const handleCreateMonitor = async (payload) => {
    await createMutation.mutateAsync(payload);
  };

  const filterTabs = [
    { label: 'All', value: '' },
    { label: 'Up', value: 'UP' },
    { label: 'Down', value: 'DOWN' },
    { label: 'Paused', value: 'PAUSED' },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Top Banner & Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Overview</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Real-time status and health across all monitored services
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => {
              refetchMonitors();
              refetchSummary();
            }}
            title="Refresh now"
            className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg border border-slate-200 bg-white transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin text-indigo-600' : ''}`} />
          </button>
          <button
            type="button"
            onClick={() => setIsAddOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold rounded-xl shadow-xs shadow-indigo-600/30 transition-all cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            <span>Add Monitor</span>
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      {isSummaryLoading ? (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          {[...Array(6)].map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
          className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4"
        >
          {/* Total Monitors */}
          <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
            <span className="text-xs font-medium text-slate-500">Monitors</span>
            <div className="mt-1 text-2xl font-bold text-slate-900">
              {summary?.total_monitors ?? 0}
            </div>
            <span className="text-[11px] text-slate-400 mt-1 block">Total configured</span>
          </div>

          {/* Operational */}
          <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
            <span className="text-xs font-medium text-emerald-600 flex items-center gap-1.5">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
              <span>Up</span>
            </span>
            <div className="mt-1 text-2xl font-bold text-emerald-600">
              {summary?.up_count ?? 0}
            </div>
            <span className="text-[11px] text-slate-400 mt-1 block">Healthy services</span>
          </div>

          {/* Down */}
          <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
            <span className="text-xs font-medium text-rose-600 flex items-center gap-1.5">
              <span className="h-1.5 w-1.5 rounded-full bg-rose-500" />
              <span>Down</span>
            </span>
            <div className="mt-1 text-2xl font-bold text-rose-600">
              {summary?.down_count ?? 0}
            </div>
            <span className="text-[11px] text-slate-400 mt-1 block">Failing checks</span>
          </div>

          {/* Paused */}
          <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
            <span className="text-xs font-medium text-slate-500">Paused</span>
            <div className="mt-1 text-2xl font-bold text-slate-700">
              {summary?.paused_count ?? 0}
            </div>
            <span className="text-[11px] text-slate-400 mt-1 block">Monitoring disabled</span>
          </div>

          {/* 24h Uptime */}
          <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
            <span className="text-xs font-medium text-slate-500">24h Uptime</span>
            <div className="mt-1 text-2xl font-bold text-slate-900">
              {summary?.uptime_24h != null ? `${summary.uptime_24h}%` : '100%'}
            </div>
            <span className="text-[11px] text-slate-400 mt-1 block">Overall reliability</span>
          </div>

          {/* Open Incidents */}
          <div className="bg-white p-4 rounded-xl border border-slate-200/80 shadow-xs">
            <span className="text-xs font-medium text-amber-600 flex items-center gap-1.5">
              <AlertTriangle className="w-3.5 h-3.5" />
              <span>Open Incidents</span>
            </span>
            <div className="mt-1 text-2xl font-bold text-amber-600">
              {summary?.open_incidents ?? 0}
            </div>
            <span className="text-[11px] text-slate-400 mt-1 block">Active investigations</span>
          </div>
        </motion.div>
      )}

      {/* Monitor List Filter Tabs */}
      <div className="flex items-center justify-between border-b border-slate-200 pb-3">
        <div className="flex items-center gap-1">
          {filterTabs.map((tab) => (
            <button
              key={tab.value}
              type="button"
              onClick={() => setStatusFilter(tab.value)}
              className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors ${
                statusFilter === tab.value
                  ? 'bg-indigo-50 text-indigo-700 font-semibold border border-indigo-200'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
        <span className="text-xs text-slate-400 font-mono">
          Auto-refreshing every 15s
        </span>
      </div>

      {/* Monitor List Items */}
      {isMonitorsLoading ? (
        <div className="space-y-3">
          {[...Array(4)].map((_, i) => (
            <MonitorRowSkeleton key={i} />
          ))}
        </div>
      ) : monitors.length === 0 ? (
        <div className="bg-white rounded-2xl border border-slate-200/80 p-12 text-center shadow-xs">
          <div className="w-12 h-12 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center mx-auto mb-4">
            <Activity className="w-6 h-6" />
          </div>
          <h3 className="text-base font-semibold text-slate-900">No monitors found</h3>
          <p className="mt-1.5 text-sm text-slate-500 max-w-sm mx-auto">
            {statusFilter
              ? `No monitors matching status "${statusFilter}". Try another filter.`
              : 'Add your first monitor to start pinging your Render app or website.'}
          </p>
          <div className="mt-6">
            <button
              type="button"
              onClick={() => setIsAddOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold rounded-xl shadow-xs transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Add Your First Monitor</span>
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs divide-y divide-slate-100 overflow-hidden">
          {monitors.map((m, idx) => (
            <motion.div
              key={m.id}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.15, delay: idx * 0.03 }}
              className="p-4 sm:p-5 hover:bg-slate-50/80 transition-colors flex flex-col sm:flex-row sm:items-center justify-between gap-4"
            >
              {/* Left Details */}
              <div className="flex items-start gap-3.5 min-w-0">
                <div className="mt-1">
                  <StatusBadge status={m.status} size="sm" />
                </div>
                <div className="min-w-0">
                  <Link
                    to={`/monitors/${m.id}`}
                    className="text-sm font-semibold text-slate-900 hover:text-indigo-600 transition-colors flex items-center gap-1.5 truncate"
                  >
                    <span>{m.name}</span>
                    <ArrowUpRight className="w-3.5 h-3.5 text-slate-400 group-hover:text-indigo-600 shrink-0" />
                  </Link>
                  <div className="flex items-center gap-3 mt-1 text-xs text-slate-500">
                    <span className="truncate max-w-[200px] sm:max-w-xs">{m.url}</span>
                    <span>•</span>
                    <span className="flex items-center gap-1 text-slate-400">
                      <Clock className="w-3 h-3" />
                      <span>Every {m.interval_seconds / 60}m</span>
                    </span>
                  </div>
                </div>
              </div>

              {/* Right Metrics & Sparkline */}
              <div className="flex items-center justify-between sm:justify-end gap-6 sm:gap-8 pt-2 sm:pt-0 border-t sm:border-0 border-slate-100">
                {/* Last Response Time */}
                <div className="text-left sm:text-right">
                  <div className="text-xs font-semibold text-slate-900 font-mono">
                    {m.last_response_time_ms != null ? `${m.last_response_time_ms} ms` : '—'}
                  </div>
                  <div className="text-[11px] text-slate-400">
                    {timeAgo(m.last_checked_at)}
                  </div>
                </div>

                {/* Sparkline Trend */}
                <div className="hidden md:block">
                  <Sparkline data={m.sparkline} width={100} height={24} />
                </div>

                {/* 24h Uptime */}
                <div className="text-right min-w-[50px]">
                  <div className={`text-xs font-semibold font-mono ${
                    m.uptime_24h >= 99 ? 'text-emerald-600' : m.uptime_24h >= 90 ? 'text-amber-600' : 'text-rose-600'
                  }`}>
                    {m.uptime_24h != null ? `${m.uptime_24h}%` : '100%'}
                  </div>
                  <div className="text-[11px] text-slate-400">24h Uptime</div>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      )}

      {/* Add Monitor Modal */}
      <MonitorModal
        isOpen={isAddOpen}
        onClose={() => setIsAddOpen(false)}
        onSubmit={handleCreateMonitor}
      />
    </div>
  );
};
