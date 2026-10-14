import React from 'react';
import { CheckCircle2, AlertCircle, PauseCircle, HelpCircle } from 'lucide-react';

export const StatusBadge = ({ status, size = 'md' }) => {
  const normStatus = (status || 'UNKNOWN').toUpperCase();

  const configs = {
    UP: {
      bg: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      dot: 'bg-emerald-500',
      pulse: true,
      icon: CheckCircle2,
      label: 'Operational',
    },
    DOWN: {
      bg: 'bg-rose-50 text-rose-700 border-rose-200',
      dot: 'bg-rose-500',
      pulse: false,
      icon: AlertCircle,
      label: 'Downtime',
    },
    PAUSED: {
      bg: 'bg-slate-100 text-slate-600 border-slate-200',
      dot: 'bg-slate-400',
      pulse: false,
      icon: PauseCircle,
      label: 'Paused',
    },
    UNKNOWN: {
      bg: 'bg-amber-50 text-amber-700 border-amber-200',
      dot: 'bg-amber-500',
      pulse: false,
      icon: HelpCircle,
      label: 'Pending',
    },
  };

  const conf = configs[normStatus] || configs.UNKNOWN;
  const Icon = conf.icon;

  const sizeClasses = size === 'sm' 
    ? 'text-xs px-2 py-0.5 gap-1.5' 
    : 'text-sm px-2.5 py-1 gap-2';

  return (
    <span className={`inline-flex items-center font-medium rounded-full border ${conf.bg} ${sizeClasses}`}>
      <span className="relative flex h-2 w-2">
        {conf.pulse && (
          <span className={`animate-ping absolute inline-flex h-full w-full rounded-full ${conf.dot} opacity-75`} />
        )}
        <span className={`relative inline-flex rounded-full h-2 w-2 ${conf.dot}`} />
      </span>
      <span>{conf.label}</span>
    </span>
  );
};
