import React from 'react';

export const Sparkline = ({ data = [], width = 120, height = 28 }) => {
  if (!data || data.length === 0) {
    return <span className="text-xs text-slate-400 font-mono">No data</span>;
  }

  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max === min ? 1 : max - min;
  const padding = 2;
  const chartHeight = height - padding * 2;

  const points = data.map((val, idx) => {
    const x = padding + (idx / Math.max(1, data.length - 1)) * (width - padding * 2);
    const y = height - padding - ((val - min) / range) * chartHeight;
    return `${x.toFixed(1)},${y.toFixed(1)}`;
  });

  const pathD = `M ${points.join(' L ')}`;

  return (
    <div className="flex items-center gap-2">
      <svg
        width={width}
        height={height}
        className="overflow-visible"
        aria-label="Response time trend sparkline"
      >
        <path
          d={pathD}
          fill="none"
          stroke="#6366f1"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        {data.length > 0 && (
          <circle
            cx={points[points.length - 1].split(',')[0]}
            cy={points[points.length - 1].split(',')[1]}
            r="2"
            fill="#4f46e5"
          />
        )}
      </svg>
    </div>
  );
};
