import { useEffect, useState } from 'react'

import { TopNav } from './TopNav'

export function AdminUsersPage() {
  const [users, setUsers] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchUsers() {
      try {
        const response = await fetch('/api/admin/users/', {
          credentials: 'include',
        })

        if (!response.ok) {
          throw new Error('Admin access required')
        }

        const payload = await response.json()
        const userList = payload?.Data ?? payload?.data ?? []
        setUsers(userList)
      } catch (err) {
        setError('Unable to load admin users. You may not have admin access.')
      } finally {
        setLoading(false)
      }
    }

    fetchUsers()
  }, [])

  if (loading) {
    return (
      <main className="auth-shell">
        <div className="card loading">Loading users...</div>
      </main>
    )
  }

  if (error) {
    return (
      <main className="auth-shell">
        <div className="card error-box">
          <h2>Admin access required</h2>
          <p>{error}</p>
          <button className="secondary-button" onClick={() => (window.location.href = '/profile')}>
            Back to profile
          </button>
        </div>
      </main>
    )
  }

  return (
    <main className="profile-shell">
      <TopNav activeTab="admin" isAdmin={true} />

      <div className="profile-card fade-in admin-panel">
        <div className="profile-header">
          <p className="eyebrow">Admin panel</p>
          <h1>All users</h1>
        </div>

        <div className="table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>Email</th>
                <th>Role</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td>{user.email || 'N/A'}</td>
                  <td>
                    <span className={`role-chip ${user.role === 'admin' ? 'admin' : 'user'}`}>
                      {(user.role || 'user').toUpperCase()}
                    </span>
                  </td>
                  <td>{user.created_at ? new Date(user.created_at).toLocaleString() : 'N/A'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <button className="secondary-button" onClick={() => (window.location.href = '/profile')}>
          Back to profile
        </button>
      </div>
    </main>
  )
}
