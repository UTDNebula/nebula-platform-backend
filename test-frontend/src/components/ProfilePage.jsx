import { useEffect, useState } from 'react'

import { TopNav } from './TopNav'

export function ProfilePage() {
  const [user, setUser] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchProfile() {
      try {
        const response = await fetch('/api/users/profile', {
          credentials: 'include',
        })

        if (!response.ok) {
          throw new Error('Not authenticated')
        }

        const payload = await response.json()
        const userData = payload?.Data ?? payload?.data ?? payload

        setUser(userData)
      } catch (err) {
        setError('Unable to load profile. Please sign in again.')
      } finally {
        setLoading(false)
      }
    }

    fetchProfile()
  }, [])

  if (loading) {
    return (
      <main className="auth-shell">
        <div className="card loading">Loading profile...</div>
      </main>
    )
  }

  if (error) {
    return (
      <main className="auth-shell">
        <div className="card error-box">
          <h2>Profile unavailable</h2>
          <p>{error}</p>
          <button className="secondary-button" onClick={() => (window.location.href = '/')}>
            Back to login
          </button>
        </div>
      </main>
    )
  }

  return (
    <main className="profile-shell">
      <TopNav activeTab="profile" isAdmin={user?.role === 'admin'} />

      <div className="profile-card fade-in">
        <div className="avatar-ring">
          {user?.avatar_url ? (
            <img
              src={user.avatar_url}
              alt={user.display_name || user.email}
              referrerPolicy="no-referrer"
            />
          ) : (
            <span>{(user?.display_name || user?.email || 'N').slice(0, 1).toUpperCase()}</span>
          )}
        </div>

        <div className="profile-header">
          <p className="eyebrow">User profile</p>
          <h1>{user?.display_name || 'Nebula User'}</h1>
          <span className={`role-chip ${user?.role === 'admin' ? 'admin' : 'user'}`}>
            {(user?.role || 'user').toUpperCase()}
          </span>
        </div>

        <div className="profile-grid">
          <div className="info-item">
            <label>Email</label>
            <strong>{user?.email || 'Not provided'}</strong>
          </div>
          <div className="info-item">
            <label>Created</label>
            <strong>{user?.created_at ? new Date(user.created_at).toLocaleString() : 'N/A'}</strong>
          </div>
          <div className="info-item">
            <label>Updated</label>
            <strong>{user?.updated_at ? new Date(user.updated_at).toLocaleString() : 'N/A'}</strong>
          </div>
        </div>

        <button
          className="secondary-button"
          onClick={() => {
            window.location.href = 'http://localhost:8081/api/auth/google/logout'
          }}
        >
          Logout
        </button>
      </div>
    </main>
  )
}
