export function AuthErrorPage() {
  return (
    <main className="auth-shell">
      <div className="card error-box">
        <p className="eyebrow">Authentication error</p>
        <h1>We couldn't sign you in</h1>
        <p className="subtext">Something went wrong.</p>
        <button
          className="google-button"
          onClick={() => {
            window.location.href = 'http://localhost:8081/api/auth/google/login'
          }}
          type="button"
        >
          <span className="google-mark">G</span>
          Try again with Google
        </button>
      </div>
    </main>
  )
}
