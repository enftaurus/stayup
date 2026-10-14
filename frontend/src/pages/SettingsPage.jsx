import React from 'react';
import { useAuth } from '../context/AuthContext';
import { User, Mail, Github, LogOut } from 'lucide-react';

export const SettingsPage = () => {
  const { user, logoutUser } = useAuth();

  if (!user) return null;

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10 space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Account Settings</h1>
        <p className="text-sm text-slate-500 mt-0.5">
          Manage your profile and authentication session
        </p>
      </div>

      <div className="bg-white rounded-2xl border border-slate-200/80 p-6 sm:p-8 shadow-xs space-y-6">
        <div className="flex items-center gap-5 pb-6 border-b border-slate-100">
          {user.avatar ? (
            <img
              src={user.avatar}
              alt={user.name || user.username}
              className="w-16 h-16 rounded-full border-2 border-indigo-100 object-cover"
            />
          ) : (
            <div className="w-16 h-16 rounded-full bg-indigo-50 text-indigo-700 font-bold text-xl flex items-center justify-center border border-indigo-200">
              {(user.username || 'U')[0].toUpperCase()}
            </div>
          )}
          <div>
            <h2 className="text-lg font-bold text-slate-900">{user.name || user.username}</h2>
            <p className="text-xs text-slate-500 flex items-center gap-1.5 mt-0.5">
              <Github className="w-3.5 h-3.5 text-slate-400" />
              <span>@{user.username}</span>
            </p>
          </div>
        </div>

        <div className="space-y-4 text-sm">
          <div className="flex items-center justify-between py-2 border-b border-slate-100">
            <span className="text-slate-500 flex items-center gap-2">
              <User className="w-4 h-4 text-slate-400" />
              <span>Full Name</span>
            </span>
            <span className="font-medium text-slate-800">{user.name || '—'}</span>
          </div>

          <div className="flex items-center justify-between py-2 border-b border-slate-100">
            <span className="text-slate-500 flex items-center gap-2">
              <Mail className="w-4 h-4 text-slate-400" />
              <span>Email Address</span>
            </span>
            <span className="font-medium text-slate-800">{user.email || '—'}</span>
          </div>

          <div className="flex items-center justify-between py-2 border-b border-slate-100">
            <span className="text-slate-500 flex items-center gap-2">
              <Github className="w-4 h-4 text-slate-400" />
              <span>GitHub Handle</span>
            </span>
            <span className="font-medium text-slate-800">@{user.username}</span>
          </div>
        </div>

        <div className="pt-4 flex justify-end">
          <button
            type="button"
            onClick={logoutUser}
            className="inline-flex items-center gap-2 px-4 py-2 text-xs font-semibold rounded-lg bg-rose-50 text-rose-600 hover:bg-rose-100 border border-rose-200 transition-colors"
          >
            <LogOut className="w-4 h-4" />
            <span>Sign Out</span>
          </button>
        </div>
      </div>
    </div>
  );
};
