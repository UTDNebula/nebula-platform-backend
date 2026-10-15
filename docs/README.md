# API documentation

Start the backend using its normal database and provider configuration, then open
`/swagger/index.html` on the backend (for example, `http://localhost:8081/swagger/index.html`).
The generated specification is also served at `/swagger/doc.json`.

The specification covers authentication, the current user profile, and the admin
user list. Authentication routes are browser redirect flows: start login in a browser and let the provider return to the callback
with the same session cookie. The callback is not a standalone login request.

After editing Swaggo annotations, regenerate the committed documentation from the
repository root:

```sh
go generate .
```

The generator is tracked as a Go tool in `go.mod`, so no separate global `swag`
installation is required. Edit the comments in `server.go` and
`internal/controllers/auth.go` and `internal/controllers/user.go`; `docs.go`, `swagger.json`, and `swagger.yaml` are generated.

Protected user routes require the `nebula_session` cookie. Complete login in the
same browser and use Swagger UI on the backend origin so the browser sends it
automatically. Swagger 2.0 has no native cookie security scheme, so the spec uses
a header API-key scheme to describe it; entering Cookie in Authorize cannot set
that forbidden header from browser JavaScript.

The profile endpoint currently returns HTTP 202. The admin list returns HTTP 401
for both unauthenticated and non-admin sessions. The documentation preserves these
existing behaviors.
