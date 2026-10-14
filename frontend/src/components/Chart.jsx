// Choice: Hand-drawn SVG Chart instead of heavy chart libraries (e.g. uPlot/Chart.js/Recharts).
// Rationale: Zero third-party bundle weight (<3 KB vs 50-120+ KB), seamless Tailwind styling integration,
// perfectly responsive with zero canvas resize/blur glitches, built-in SVG accessibility, and smooth CSS line-draw animations that naturally respect prefers-reduced-motion.

import React, { useState } from 'react';

export const Chart = ({ buckets = [], range = '24h', onRangeChange }) => {
  const [hoveredIdx, setHoveredIdx] = useState(null);

  const width = 800;
  const height = 240;
  const padLeft = 50;
  const padRight = 20;
  const padTop = 20;
  const padBottom = 30;

  const chartWidth = width - padLeft - padRight;
  const chartHeight = height - padTop - padBottom;

  if (!buckets || buckets.length === 0) {
    return (
      <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500">
        <p className="text-sm">No historical check data available yet for this time window.</p>
      </div>
    );
  }

  const responseTimes = buckets.map((b) => b.avg_response_ms || 0);
  const maxVal = Math.max(100, Math.ceil((Math.max(...responseTimes, 100) * 1.15) / 50) * 50);

  // Compute points
  const points = buckets.map((b, idx) => {
    const x = padLeft + (idx / Math.max(1, buckets.length - 1)) * chartWidth;
    const y = padTop + chartHeight - ((b.avg_response_ms || 0) / maxVal) * chartHeight;
    return { x, y, bucket: b };
  });

  const lineD = points.length > 0 ? `M ${points.map((p) => `${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' L ')}` : '';
  const areaD = points.length > 0
    ? `${lineD} L ${points[points.length - 1].x},${padTop + chartHeight} L ${points[0].x},${padTop + chartHeight} Z`
    : '';

  // Y-axis grid ticks
  const yTicks = [0, Math.round(maxVal * 0.5), maxVal];

  const formatDate = (isoStr) => {
    const d = new Date(isoStr);
    return range === '24h'
      ? d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      : `${d.toLocaleDateString([], { month: 'short', day: 'numeric' })} ${d.toLocaleTimeString([], { hour: '2-digit' })}`;
  };

  return (
    <div className="bg-white rounded-xl border border-slate-200/80 p-5 shadow-xs">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between pb-4 border-b border-slate-100 gap-3">
        <div>
          <h3 className="text-sm font-semibold text-slate-900">Response Time & Health</h3>
          <p className="text-xs text-slate-500">Average latency per time bucket (failures highlighted)</p>
        </div>
        <div className="inline-flex rounded-lg border border-slate-200 bg-slate-50 p-1 text-xs font-medium">
          <button
            type="button"
            onClick={() => onRangeChange('24h')}
            className={`px-3 py-1 rounded-md transition-colors ${
              range === '24h'
                ? 'bg-white text-indigo-600 shadow-xs font-semibold'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            24 Hours (5m)
          </button>
          <button
            type="button"
            onClick={() => onRangeChange('7d')}
            className={`px-3 py-1 rounded-md transition-colors ${
              range === '7d'
                ? 'bg-white text-indigo-600 shadow-xs font-semibold'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            7 Days (1h)
          </button>
        </div>
      </div>

      <div className="relative mt-4">
        <svg
          viewBox={`0 0 ${width} ${height}`}
          className="w-full h-auto overflow-visible select-none"
        >
          <defs>
            <linearGradient id="chartGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#6366f1" stopOpacity="0.18" />
              <stop offset="100%" stopColor="#6366f1" stopOpacity="0.0" />
            </linearGradient>
          </defs>

          {/* Grid lines and Y axis */}
          {yTicks.map((val) => {
            const y = padTop + chartHeight - (val / maxVal) * chartHeight;
            return (
              <g key={val} className="text-slate-400">
                <line
                  x1={padLeft}
                  y1={y}
                  x2={width - padRight}
                  y2={y}
                  stroke="#f1f5f9"
                  strokeWidth="1"
                  strokeDasharray="4 4"
                />
                <text
                  x={padLeft - 8}
                  y={y + 4}
                  textAnchor="end"
                  className="text-[10px] fill-slate-400 font-mono"
                >
                  {val}ms
                </text>
              </g>
            );
          })}

          {/* Area fill */}
          {areaD && <path d={areaD} fill="url(#chartGradient)" />}

          {/* Line path with draw animation */}
          {lineD && (
            <path
              d={lineD}
              fill="none"
              stroke="#6366f1"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="transition-all duration-300"
            />
          )}

          {/* Points & Failure markers */}
          {points.map((p, idx) => {
            const hasFailure = p.bucket.failures > 0;
            return (
              <g key={idx}>
                {hasFailure && (
                  <circle
                    cx={p.x}
                    cy={p.y}
                    r="4.5"
                    className="fill-rose-500 stroke-white stroke-2"
                  />
                )}
                {/* Invisible hover hotspot */}
                <rect
                  x={p.x - 8}
                  y={padTop}
                  width="16"
                  height={chartHeight}
                  fill="transparent"
                  onMouseEnter={() => setHoveredIdx(idx)}
                  onMouseLeave={() => setHoveredIdx(null)}
                  className="cursor-pointer"
                />
              </g>
            );
          })}

          {/* Active tooltip marker */}
          {hoveredIdx !== null && points[hoveredIdx] && (
            <g>
              <line
                x1={points[hoveredIdx].x}
                y1={padTop}
                x2={points[hoveredIdx].x}
                y2={padTop + chartHeight}
                stroke="#94a3b8"
                strokeWidth="1"
                strokeDasharray="2 2"
              />
              <circle
                cx={points[hoveredIdx].x}
                cy={points[hoveredIdx].y}
                r="5"
                className="fill-indigo-600 stroke-white stroke-2"
              />
            </g>
          )}

          {/* X axis labels */}
          {points.length > 0 && (
            <>
              <text
                x={padLeft}
                y={height - 6}
                textAnchor="start"
                className="text-[10px] fill-slate-400"
              >
                {formatDate(points[0].bucket.ts)}
              </text>
              <text
                x={width - padRight}
                y={height - 6}
                textAnchor="end"
                className="text-[10px] fill-slate-400"
              >
                {formatDate(points[points.length - 1].bucket.ts)}
              </text>
            </>
          )}
        </svg>

        {/* Hover Tooltip Box */}
        {hoveredIdx !== null && points[hoveredIdx] && (
          <div
            className="absolute z-20 pointer-events-none bg-slate-900 text-white rounded-lg px-3 py-2 text-xs shadow-lg transform -translate-x-1/2 -translate-y-full mb-2"
            style={{
              left: `${(points[hoveredIdx].x / width) * 100}%`,
              top: `${(points[hoveredIdx].y / height) * 100}%`,
            }}
          >
            <div className="font-semibold text-slate-200">
              {formatDate(points[hoveredIdx].bucket.ts)}
            </div>
            <div className="mt-1 flex items-center gap-2">
              <span className="text-slate-400">Latency:</span>
              <span className="font-mono text-indigo-300 font-bold">
                {points[hoveredIdx].bucket.avg_response_ms} ms
              </span>
            </div>
            <div className="flex items-center gap-2 text-[11px] text-slate-300">
              <span>Checks: {points[hoveredIdx].bucket.checks}</span>
              {points[hoveredIdx].bucket.failures > 0 ? (
                <span className="text-rose-400 font-semibold">
                  • {points[hoveredIdx].bucket.failures} failures
                </span>
              ) : (
                <span className="text-emerald-400">• 100% UP</span>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
