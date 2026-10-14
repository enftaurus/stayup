import React from 'react';

export const Skeleton = ({ className = '' }) => {
  return <div className={`skeleton-shimmer rounded-md ${className}`} />;
};

export const CardSkeleton = () => (
  <div className="bg-white rounded-xl border border-slate-200/80 p-5 shadow-xs space-y-3">
    <div className="flex justify-between items-center">
      <Skeleton className="h-4 w-24" />
      <Skeleton className="h-6 w-16 rounded-full" />
    </div>
    <Skeleton className="h-7 w-20" />
    <Skeleton className="h-3 w-32" />
  </div>
);

export const MonitorRowSkeleton = () => (
  <div className="bg-white rounded-xl border border-slate-200/80 p-4 shadow-xs flex items-center justify-between gap-4">
    <div className="flex items-center gap-3">
      <Skeleton className="h-8 w-8 rounded-full" />
      <div className="space-y-1.5">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-3 w-56" />
      </div>
    </div>
    <div className="hidden sm:flex items-center gap-8">
      <Skeleton className="h-5 w-24" />
      <Skeleton className="h-5 w-16" />
      <Skeleton className="h-5 w-14" />
    </div>
  </div>
);
