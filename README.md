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
| `PROXY` | Transparent reverse proxy — forwards the full request to the destination and rewrites domain references in response headers/body |
| `REDIRECT` | HTTP 307 Temporary Redirect to the destination URL |
| `IFRAME` | Serves a full-page iframe wrapping the destination URL |

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

Swagger UI is available at `/url-management/swagger/index.html`.

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
