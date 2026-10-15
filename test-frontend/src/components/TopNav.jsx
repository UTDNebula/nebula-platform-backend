export function TopNav({ activeTab, isAdmin }) {
  const options = [
    { label: 'Profile', value: '/profile' },
    ...(isAdmin ? [{ label: 'Admin', value: '/admin/users' }] : []),
  ]

  const selected = options.find(
    (option) => option.value === (activeTab === 'admin' ? '/admin/users' : '/profile')
  )

  return (
    <nav className="top-nav" aria-label="Main navigation">
      <label className="nav-select-wrap">
        <span className="sr-only">Switch view</span>
        <select
          className="nav-select"
          value={selected?.value ?? '/profile'}
          onChange={(event) => {
            window.location.href = event.target.value
          }}
        >
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
    </nav>
  )
}
