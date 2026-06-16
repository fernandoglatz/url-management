# url-management

A Go service for managing URL redirects with support for three redirect strategies: transparent proxy, HTTP redirect, and iframe embedding. Redirects can be triggered by ID or matched automatically based on the incoming request's hostname (DNS-based routing).

## Tech Stack

- **Go 1.25** with [Gin](https://github.com/gin-gonic/gin)
- **MongoDB** — persistent storage for redirect rules
- **Redis** — cache layer with configurable TTL (default 24h)
- **Swagger** — auto-generated API docs via swaggo
- **Docker / Docker Compose**

## Redirect Types

| Type | Behavior |
|------|----------|
| `PROXY` | Transparent reverse proxy — forwards requests to the destination, rewrites domain and external URL references in headers/body, proxies WebSocket upgrades, and strips browser-blocking headers (CSP, HSTS, X-Frame-Options) |
| `REDIRECT` | HTTP 307 Temporary Redirect to the destination URL |
| `IFRAME` | Serves a full-page iframe wrapping the destination URL |

### Proxy mode details

When `type` is `PROXY`, the service acts as a full reverse proxy:

- **Domain rewriting** — replaces the destination domain with the proxy domain in response headers and text-based bodies (HTML, CSS, JS, JSON, XML, etc.)
- **External URL rewriting** — rewrites external URLs in `src`, `href`, `url()`, `srcset`, `@import`, Module Federation remotes, and HTML entity-encoded values through the built-in CDN proxy (`/__cdnp/`)
- **WebSocket** — transparently tunnels WebSocket connections (plain and TLS) to the upstream host
- **Cookie rewriting** — adjusts `Set-Cookie` `Domain`, `Secure`, and `SameSite` attributes to match the proxy host
- **Hop-by-hop header filtering** — strips `Connection`, `Transfer-Encoding`, `Upgrade`, etc. per RFC 7230
- **SRI stripping** — removes `integrity` attributes from `<script>` and `<link>` tags whose content has been rewritten

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Run with Docker Compose

```bash
docker compose up -d
```

This starts the application on port `8080` along with MongoDB and Redis.

### Run locally

```bash
# Start dependencies
docker compose up -d mongo redis

# Run the app
go run main.go
```

The server listens on `0.0.0.0:8080` with context path `/url-management`.

## Configuration

Configuration is loaded from `conf/application.yml`:

```yaml
server:
  listening: "0.0.0.0:8080"
  context-path: "/url-management"

data:
  mongo:
    uri: "mongodb://mongo:27017"
    database: "url-management"
  redis:
    address: "redis:6379"
    password: ""
    db: 0
    ttl:
      redirect: 24h

log:
  level: TRACE
  format: TEXT
  colored: true
```

Environment variables are loaded from `.env` at startup.

## API

Base path: `/url-management`

Swagger UI is available at `/url-management/swagger-ui/index.html`.

### Redirect management

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/redirect` | Create a redirect |
| `GET` | `/redirect` | List all redirects |
| `GET` | `/redirect/{id}` | Get a redirect by ID |
| `PUT` | `/redirect/{id}` | Update a redirect |
| `POST` | `/redirect/{id}` | Update a redirect |
| `DELETE` | `/redirect/{id}` | Delete a redirect |

### Execute a redirect

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/?to={id}` | Execute redirect by ID |
| `GET` | `/*` | DNS-based redirect (matches the request hostname) |

### CDN proxy (used internally by PROXY mode)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/__cdn?url={url}` | Proxy an external resource by URL query parameter |
| `GET` | `/__cdnp/{host}/{path}` | Proxy an external resource with target host and path encoded in the URL path (preserves webpack `publicPath` detection) |

Both endpoints rewrite external URLs in CSS responses so that nested assets are also routed through the proxy.

### Health

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |

### Request body

```json
{
  "dns": "example.com",
  "destination": "https://target.example.com",
  "type": "PROXY"
}
```

### Example

```bash
# Create a redirect
curl -X PUT http://localhost:8080/url-management/redirect \
  -H 'Content-Type: application/json' \
  -d '{"dns": "short.example.com", "destination": "https://www.example.com", "type": "REDIRECT"}'

# Execute redirect by ID
curl -L http://localhost:8080/url-management/?to=<id>
```

## Database Migrations

MongoDB migrations are stored in `scripts/mongo/migrations/` and run automatically on startup via `golang-migrate`.

## License

Apache 2.0 — see [LICENSE](LICENSE).
