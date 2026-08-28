import React, { useState, useEffect } from 'react';

export default function App() {
  const [userData, setUserData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const state = urlParams.get('state');

    if (code && state) {
      handleCallback(code, state);
    }
  }, []);

  const handleCallback = async (code, state) => {
    setLoading(true);
    setError(null);
    try {
      // Clean up the URL query params without reloading the page
      window.history.replaceState({}, document.title, window.location.pathname);

      // Call backend callback endpoint
      const response = await fetch(`/auth/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`, {
        method: 'GET',
        headers: {
          'Accept': 'application/json',
        },
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP ${response.status}: ${errorText || response.statusText}`);
      }

      const data = await response.json();
      setUserData(data);
    } catch (err) {
      console.error('OAuth Callback Error:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleLogin = () => {
    // Redirect browser to Go backend OAuth endpoint
    window.location.href = 'http://localhost:8080/auth/url';
  };

  const handleCopyJson = () => {
    if (userData) {
      navigator.clipboard.writeText(JSON.stringify(userData, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleReset = () => {
    setUserData(null);
    setError(null);
  };

  return (
    <div className="container">
      <div className="header">
        <div className="logo-group">
          <svg className="logo-icon" viewBox="0 0 16 16" version="1.1" aria-hidden="true">
            <path fillRule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.28.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
          </svg>
          <div>
            <h1>StayUp</h1>
            <p className="subtitle">GitHub OAuth JSON Inspector</p>
          </div>
        </div>

        {userData && (
          <button className="btn btn-secondary" onClick={handleReset}>
            Clear & Reset
          </button>
        )}
      </div>

      {error && (
        <div className="error-box">
          <strong>Authentication Error:</strong> {error}
        </div>
      )}

      {loading ? (
        <div className="loader">
          <div className="spinner"></div>
          <p>Exchanging OAuth code and fetching GitHub JSON...</p>
        </div>
      ) : userData ? (
        <div>
          {/* GitHub User Profile Card */}
          <div className="profile-section">
            {userData.avatar_url && (
              <img src={userData.avatar_url} alt={userData.name || userData.login} className="avatar" />
            )}
            <div className="profile-info">
              <h3>{userData.name || userData.login || 'GitHub User'}</h3>
              <p className="username">@{userData.login || 'user'}</p>
              
              <div className="badges">
                {userData.public_repos !== undefined && (
                  <span className="badge">📦 {userData.public_repos} Repos</span>
                )}
                {userData.followers !== undefined && (
                  <span className="badge">👥 {userData.followers} Followers</span>
                )}
                {userData.location && (
                  <span className="badge">📍 {userData.location}</span>
                )}
                {userData.email && (
                  <span className="badge">✉️ {userData.email}</span>
                )}
              </div>
            </div>
          </div>

          {/* Raw JSON Inspector */}
          <div className="json-header">
            <h4>Raw Response Payload (`/user`)</h4>
            <button className="btn btn-secondary" onClick={handleCopyJson} style={{ padding: '6px 12px', fontSize: '0.85rem' }}>
              {copied ? '✓ Copied!' : 'Copy JSON'}
            </button>
          </div>
          <pre className="json-box">{JSON.stringify(userData, null, 2)}</pre>
        </div>
      ) : (
        <div className="login-card">
          <h2>Test GitHub OAuth Response</h2>
          <p>
            Click below to initiate GitHub OAuth flow via your Go backend. 
            Once authorized, the JSON received from GitHub's <code>/user</code> endpoint will be rendered here.
          </p>
          <button className="btn" onClick={handleLogin}>
            <svg style={{ width: 18, height: 18, fill: 'currentColor' }} viewBox="0 0 16 16">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.28.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
            </svg>
            Login with GitHub
          </button>
        </div>
      )}
    </div>
  );
}
