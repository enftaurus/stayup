import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { AlertOctagon, CheckCircle2, ArrowRight } from 'lucide-react';
import { listIncidents } from '../api/incidents';
import { Skeleton } from '../components/Skeleton';

export const IncidentsPage = () => {
  const [tab, setTab] = useState('open'); // 'open' | 'resolved'

  const { data, isLoading } = useQuery({
    queryKey: ['incidents', tab],
    queryFn: () => listIncidents(tab),
    refetchInterval: 15000,
  });

  const incidents = data?.incidents || [];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Incidents</h1>
        <p className="text-sm text-slate-500 mt-0.5">
          History and ongoing outages detected across all your monitors
        </p>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-200 pb-3">
        <button
          type="button"
          onClick={() => setTab('open')}
          className={`px-3.5 py-1.5 text-xs font-semibold rounded-lg transition-colors flex items-center gap-1.5 ${
            tab === 'open'
              ? 'bg-rose-50 text-rose-700 border border-rose-200'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
          }`}
        >
          <AlertOctagon className="w-3.5 h-3.5" />
          <span>Active / Open</span>
        </button>

        <button
          type="button"
          onClick={() => setTab('resolved')}
          className={`px-3.5 py-1.5 text-xs font-semibold rounded-lg transition-colors flex items-center gap-1.5 ${
            tab === 'resolved'
              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
          }`}
        >
          <CheckCircle2 className="w-3.5 h-3.5" />
          <span>Resolved</span>
        </button>
      </div>

      {/* Incidents Table / List */}
      {isLoading ? (
        <div className="space-y-3">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full rounded-xl" />
          ))}
        </div>
      ) : incidents.length === 0 ? (
        <div className="bg-white rounded-2xl border border-slate-200/80 p-12 text-center shadow-xs">
          <div className="w-12 h-12 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center mx-auto mb-4">
            <CheckCircle2 className="w-6 h-6" />
          </div>
          <h3 className="text-base font-semibold text-slate-900">
            {tab === 'open' ? 'All systems operational' : 'No past resolved incidents'}
          </h3>
          <p className="mt-1 text-sm text-slate-500 max-w-sm mx-auto">
            {tab === 'open'
              ? 'No open downtime incidents detected across any active monitors.'
              : 'Resolved downtime events will be archived here for reference.'}
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs divide-y divide-slate-100 overflow-hidden">
          {incidents.map((inc) => (
            <div
              key={inc.id}
              className="p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-slate-50/60 transition-colors"
            >
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${
                    inc.resolved_at
                      ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                      : 'bg-rose-50 text-rose-700 border border-rose-200 animate-pulse'
                  }`}>
                    {inc.resolved_at ? 'Resolved' : 'Ongoing Outage'}
                  </span>
                  <Link
                    to={`/monitors/${inc.monitor_id}`}
                    className="text-sm font-semibold text-slate-900 hover:text-indigo-600 flex items-center gap-1 transition-colors"
                  >
                    <span>{inc.monitor_name || 'Monitor'}</span>
                    <ArrowRight className="w-3.5 h-3.5 text-slate-400" />
                  </Link>
                </div>
                <div className="text-xs text-slate-500">
                  Started: {new Date(inc.started_at).toLocaleString()}
                </div>
              </div>

              <div className="text-right text-xs">
                {inc.resolved_at ? (
                  <div className="text-slate-600">
                    <div>Resolved: {new Date(inc.resolved_at).toLocaleString()}</div>
                    <div className="text-slate-400 mt-0.5">Duration: {inc.duration_seconds}s</div>
                  </div>
                ) : (
                  <span className="text-rose-600 font-semibold font-mono">
                    Investigating Downtime
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
