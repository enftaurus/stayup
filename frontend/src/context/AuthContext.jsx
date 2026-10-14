import React, { createContext, useContext, useEffect, useState } from 'react';
import { refreshSession } from '../api/client';
import { getMe, logout as apiLogout, devLogin as apiDevLogin } from '../api/auth';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const checkAuth = async () => {
    try {
      const refreshed = await refreshSession();
      if (refreshed) {
        const profile = await getMe();
        setUser(profile);
      } else {
        setUser(null);
      }
    } catch (err) {
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  const loginWithGithub = () => {
    window.location.href = '/auth/url';
  };

  const loginDev = async () => {
    setLoading(true);
    try {
      const data = await apiDevLogin();
      setUser(data.user);
      return data;
    } finally {
      setLoading(false);
    }
  };

  const logoutUser = async () => {
    await apiLogout();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, loginWithGithub, loginDev, logoutUser, checkAuth }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
