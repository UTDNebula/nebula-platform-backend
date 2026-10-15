export function LoginPage() {
  return (
    <main className="auth-shell">
      <div className="card fade-in">
        <p className="eyebrow">Nebula Platform</p>
        <h1>Sign in to continue</h1>
        <p className="subtext">Use your Google account</p>
        <button
          className="google-button"
          onClick={() => {
            window.location.href = 'http://localhost:8081/api/auth/google/login'
          }}
          type="button"
        >
          <span className="google-mark">G</span>
          Continue with Google
        </button>
      </div>
    </main>
  )
}
