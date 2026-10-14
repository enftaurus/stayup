import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Sparkles, Clock, Globe, ShieldCheck, ChevronDown, ChevronUp } from 'lucide-react';

const INTERVAL_OPTIONS = [
  { label: '1m', seconds: 60, popular: true },
  { label: '2m', seconds: 120 },
  { label: '5m', seconds: 300, popular: true },
  { label: '10m', seconds: 600, hint: 'Render Safe' },
  { label: '15m', seconds: 900 },
  { label: '30m', seconds: 1800 },
  { label: '60m', seconds: 3600 },
];

export const MonitorModal = ({ isOpen, onClose, onSubmit, initialData = null }) => {
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [intervalSeconds, setIntervalSeconds] = useState(60);
  const [expectedStatus, setExpectedStatus] = useState(200);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [fieldErrors, setFieldErrors] = useState({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (initialData) {
      setName(initialData.name || '');
      setUrl(initialData.url || '');
      setIntervalSeconds(initialData.interval_seconds || 60);
      setExpectedStatus(initialData.expected_status || 200);
      setShowAdvanced(initialData.expected_status !== 200);
    } else {
      setName('');
      setUrl('');
      setIntervalSeconds(60);
      setExpectedStatus(200);
      setShowAdvanced(false);
    }
    setFieldErrors({});
  }, [initialData, isOpen]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFieldErrors({});
    setSubmitting(true);

    try {
      await onSubmit({
        name,
        url,
        interval_seconds: intervalSeconds,
        expected_status: Number(expectedStatus),
      });
      onClose();
    } catch (err) {
      if (err?.error?.field) {
        setFieldErrors({ [err.error.field]: err.error.message });
      } else {
        setFieldErrors({ general: err?.error?.message || 'Failed to save monitor' });
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 bg-slate-900/40 backdrop-blur-xs"
          />

          {/* Modal Container */}
          <motion.div
            initial={{ scale: 0.95, opacity: 0, y: 15 }}
            animate={{ scale: 1, opacity: 1, y: 0 }}
            exit={{ scale: 0.95, opacity: 0, y: 15 }}
            transition={{ type: 'spring', damping: 25, stiffness: 300 }}
            className="relative w-full max-w-lg bg-white rounded-2xl shadow-2xl border border-slate-100 overflow-hidden"
          >
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
              <h2 className="text-lg font-semibold text-slate-900">
                {initialData ? 'Edit Monitor' : 'Add New Monitor'}
              </h2>
              <button
                type="button"
                onClick={onClose}
                className="p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="p-6 space-y-5">
              {fieldErrors.general && (
                <div className="p-3 text-xs text-rose-600 bg-rose-50 border border-rose-200 rounded-lg">
                  {fieldErrors.general}
                </div>
              )}

              {/* Name */}
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                  Monitor Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. My Render API or Production Website"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className={`w-full px-3.5 py-2 text-sm rounded-lg border ${
                    fieldErrors.name ? 'border-rose-300 ring-2 ring-rose-100' : 'border-slate-200'
                  } focus:outline-hidden focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all`}
                />
                {fieldErrors.name && (
                  <p className="mt-1 text-xs text-rose-600">{fieldErrors.name}</p>
                )}
              </div>

              {/* URL */}
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                  Website URL
                </label>
                <div className="relative">
                  <Globe className="absolute left-3 top-2.5 w-4 h-4 text-slate-400" />
                  <input
                    type="url"
                    required
                    placeholder="https://my-app.onrender.com"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    className={`w-full pl-9 pr-3.5 py-2 text-sm rounded-lg border ${
                      fieldErrors.url ? 'border-rose-300 ring-2 ring-rose-100' : 'border-slate-200'
                    } focus:outline-hidden focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all`}
                  />
                </div>
                {fieldErrors.url && (
                  <p className="mt-1 text-xs text-rose-600">{fieldErrors.url}</p>
                )}
                <p className="mt-1.5 text-xs text-slate-500 flex items-center gap-1.5">
                  <ShieldCheck className="w-3.5 h-3.5 text-indigo-500 shrink-0" />
                  <span>StayUp checks with lightweight <strong>HTTP HEAD</strong> requests.</span>
                </p>
              </div>

              {/* Interval Chips */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider">
                    Check Interval
                  </label>
                  <span className="text-xs text-slate-400">Currently: {intervalSeconds / 60} min</span>
                </div>
                <div className="grid grid-cols-4 sm:grid-cols-7 gap-1.5">
                  {INTERVAL_OPTIONS.map((opt) => (
                    <button
                      key={opt.seconds}
                      type="button"
                      onClick={() => setIntervalSeconds(opt.seconds)}
                      className={`relative px-2 py-2 text-xs font-medium rounded-lg border transition-all text-center ${
                        intervalSeconds === opt.seconds
                          ? 'border-indigo-600 bg-indigo-50/70 text-indigo-700 font-semibold ring-1 ring-indigo-600'
                          : 'border-slate-200 hover:border-slate-300 text-slate-700 bg-white'
                      }`}
                    >
                      {opt.label}
                      {opt.hint && (
                        <span className="absolute -top-2 left-1/2 transform -translate-x-1/2 bg-amber-500 text-white text-[9px] px-1 rounded-full whitespace-nowrap">
                          {opt.hint}
                        </span>
                      )}
                    </button>
                  ))}
                </div>
                {fieldErrors.interval_seconds && (
                  <p className="mt-1 text-xs text-rose-600">{fieldErrors.interval_seconds}</p>
                )}

                {/* Render keep-awake callout */}
                <div className="mt-2.5 p-2.5 bg-amber-50/80 border border-amber-200/80 rounded-lg text-xs text-amber-800 flex items-start gap-2">
                  <Sparkles className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
                  <span>
                    <strong>Render sleep prevention:</strong> Free instances spin down after ~15 minutes without traffic. Setting an interval of 10 minutes or less guarantees continuous wakefulness.
                  </span>
                </div>
              </div>

              {/* Advanced toggle */}
              <div className="pt-1">
                <button
                  type="button"
                  onClick={() => setShowAdvanced(!showAdvanced)}
                  className="flex items-center gap-1.5 text-xs text-slate-600 hover:text-slate-900 font-medium"
                >
                  {showAdvanced ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                  <span>Advanced HTTP Settings</span>
                </button>

                {showAdvanced && (
                  <div className="mt-3 p-3 bg-slate-50 border border-slate-200 rounded-lg">
                    <label className="block text-xs font-medium text-slate-700 mb-1">
                      Expected HTTP Status Code
                    </label>
                    <input
                      type="number"
                      min="100"
                      max="599"
                      value={expectedStatus}
                      onChange={(e) => setExpectedStatus(e.target.value)}
                      className="w-32 px-3 py-1.5 text-sm bg-white rounded-md border border-slate-200 focus:outline-hidden focus:ring-1 focus:ring-indigo-500"
                    />
                    <p className="text-[11px] text-slate-500 mt-1">
                      Defaults to 200. Any other response status will be classified as an incident (DOWN).
                    </p>
                  </div>
                )}
              </div>

              {/* Actions */}
              <div className="flex items-center justify-end gap-3 pt-4 border-t border-slate-100">
                <button
                  type="button"
                  onClick={onClose}
                  className="px-4 py-2 text-sm font-medium text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="px-5 py-2 text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 rounded-lg shadow-xs shadow-indigo-600/30 transition-all"
                >
                  {submitting ? 'Saving...' : initialData ? 'Update Monitor' : 'Create Monitor'}
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
};
