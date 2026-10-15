import { useEffect, useState } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'

import { AdminUsersPage } from './components/AdminUsersPage'
import { AuthErrorPage } from './components/AuthErrorPage'
import { LoginPage } from './components/LoginPage'
import { ProfilePage } from './components/ProfilePage'

function RootRoute() {
  const [isAuthenticated, setIsAuthenticated] = useState(null)

  useEffect(() => {
    async function checkAuth() {
      try {
        const response = await fetch('/api/users/profile', {
          credentials: 'include',
        })
        setIsAuthenticated(response.ok)
      } catch {
        setIsAuthenticated(false)
      }
    }

    checkAuth()
  }, [])

  if (isAuthenticated === null) {
    return (
      <main className="auth-shell">
        <div className="card loading">Checking session...</div>
      </main>
    )
  }

  return isAuthenticated ? <Navigate to="/profile" replace /> : <LoginPage />
}

function ProfileRoute() {
  const [isAuthenticated, setIsAuthenticated] = useState(null)

  useEffect(() => {
    async function checkAuth() {
      try {
        const response = await fetch('/api/users/profile', {
          credentials: 'include',
        })
        setIsAuthenticated(response.ok)
      } catch {
        setIsAuthenticated(false)
      }
    }

    checkAuth()
  }, [])

  if (isAuthenticated === null) {
    return (
      <main className="auth-shell">
        <div className="card loading">Checking session...</div>
      </main>
    )
  }

  return isAuthenticated ? <ProfilePage /> : <Navigate to="/" replace />
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<RootRoute />} />
      <Route path="/login/error" element={<AuthErrorPage />} />
      <Route path="/profile" element={<ProfileRoute />} />
      <Route path="/admin/users" element={<AdminUsersPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
