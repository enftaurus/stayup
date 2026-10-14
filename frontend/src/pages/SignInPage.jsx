import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { motion } from 'framer-motion';
import { Globe, Zap, AlertTriangle, ArrowRight, ShieldCheck, Terminal } from 'lucide-react';

export const SignInPage = () => {
  const { user, loading, loginWithGithub, loginDev } = useAuth();

  if (!loading && user) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div className="min-h-screen bg-[#fbfbfb] flex flex-col justify-between">
      {/* Header */}
      <header className="max-w-7xl mx-auto w-full px-6 py-6 flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-indigo-600 flex items-center justify-center text-white font-bold text-lg shadow-xs shadow-indigo-600/30">
            S
          </div>
          <span className="font-bold text-slate-900 tracking-tight text-xl">StayUp</span>
          <span className="relative flex h-2 w-2 ml-1">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
          </span>
        </div>
      </header>

      {/* Hero Section */}
      <main className="max-w-4xl mx-auto px-6 py-12 text-center my-auto">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
          className="space-y-6"
        >
          {/* Badge */}
          <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-indigo-50 border border-indigo-100/80 text-indigo-700 text-xs font-semibold">
            <Zap className="w-3.5 h-3.5 text-indigo-600" />
            <span>Zero Cold Starts • Real-Time Uptime Monitoring</span>
          </div>

          {/* Heading */}
          <h1 className="text-4xl sm:text-5xl font-extrabold text-slate-900 tracking-tight leading-tight">
            Keep your Render apps awake.<br />
            <span className="text-indigo-600">Know the second they go down.</span>
          </h1>

          {/* Subtitle */}
          <p className="max-w-2xl mx-auto text-base sm:text-lg text-slate-600 leading-relaxed">
            StayUp sends scheduled HTTP HEAD checks to prevent free-tier apps on Render from spinning down after 15 minutes of inactivity, while recording uptime metrics and opening incidents the moment a site fails.
          </p>

          {/* Sign In Buttons */}
          <div className="pt-4 flex flex-col sm:flex-row items-center justify-center gap-3">
            <button
              type="button"
              onClick={loginWithGithub}
              className="w-full sm:w-auto px-6 py-3 bg-slate-900 hover:bg-black text-white font-semibold text-sm rounded-xl shadow-md hover:shadow-lg transition-all flex items-center justify-center gap-3"
            >
              <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
              </svg>
              <span>Continue with GitHub</span>
              <ArrowRight className="w-4 h-4 text-slate-400" />
            </button>

            {/* Quick Dev Sign In button for instant testing */}
            <button
              type="button"
              onClick={loginDev}
              className="w-full sm:w-auto px-6 py-3 bg-white hover:bg-slate-50 text-indigo-700 font-semibold text-sm rounded-xl border border-indigo-200 shadow-xs hover:border-indigo-300 transition-all flex items-center justify-center gap-2"
            >
              <Terminal className="w-4 h-4 text-indigo-600" />
              <span>Quick Dev Sign In (Bypass)</span>
            </button>
          </div>
        </motion.div>

        {/* 3-Step Explainer */}
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.15 }}
          className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-6 text-left"
        >
          <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold text-sm mb-4">
              1
            </div>
            <h3 className="text-base font-semibold text-slate-900">Add Your URL</h3>
            <p className="mt-2 text-sm text-slate-600 leading-relaxed">
              Enter your Render app or website URL and select an interval from 1 to 60 minutes.
            </p>
          </div>

          <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold text-sm mb-4">
              2
            </div>
            <h3 className="text-base font-semibold text-slate-900">Queue-Based Checks</h3>
            <p className="mt-2 text-sm text-slate-600 leading-relaxed">
              A high-throughput scheduler dispatches checks to Kafka workers that ping your endpoint with fast HEAD requests.
            </p>
          </div>

          <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold text-sm mb-4">
              3
            </div>
            <h3 className="text-base font-semibold text-slate-900">Never Sleeps, Instant Alerts</h3>
            <p className="mt-2 text-sm text-slate-600 leading-relaxed">
              Your free instance never spins down, while downtime immediately opens an incident on your dashboard.
            </p>
          </div>
        </motion.div>
      </main>

      {/* Footer */}
      <footer className="max-w-7xl mx-auto w-full px-6 py-6 border-t border-slate-200 text-center text-xs text-slate-400">
        StayUp • High-reliability uptime monitoring with queue-based scheduling
      </footer>
    </div>
  );
};
