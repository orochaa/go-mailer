# 📬 Mailing Microservice (Go)

A lightweight, **high-concurrency** email-dispatch service written in **Go**.  
It exposes a single HTTP endpoint, queues incoming requests in memory, and delivers mail through any SMTP-compatible server.

## 🚀 Why Go?

| Go trait                   | How it helps this service                                                                                                                                                          |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Goroutines & channels**  | Every incoming request is queued in a channel and processed by a pool of worker goroutines (10 by default). This lets the service handle bursts without blocking the HTTP handler. |
| **Compiled speed**         | The executable starts in ~20 ms and stays under 15 MiB RSS in typical Docker containers.                                                                                           |
| **Small, static binaries** | Deploy once, copy the single binary into any scratch/alpine image, and you’re done.                                                                                                |

## 🗄️ Env-File Configuration

Create a `.env` file

```env
# SMTP credentials
MAILER_HOST=smtp.example.com
MAILER_PORT=587
MAILER_USER=noreply@example.com
MAILER_PASS=super_secret_password

# HTTP server
PORT=3000 # optional – defaults to 3000
WORKERS=10 # optional – defaults to 10
```

## 🛠️ Running

```bash
go run main.go
```

**Docker**

```bash
docker compose up
```

## 📨 HTTP API

### `POST /`

#### Request body

```json
{
  "subject": "Welcome to ACME",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "message": "Thanks for signing up!"
}
```

#### Successful response `200 OK`

```json
{ "message": "Mail added to mail queue" }
```

## TODO

- **Graceful shutdown**: Wrap server in `http.Server`, listen for OS signals, call `Shutdown(ctx)` and close the `mailChannel` so workers exit cleanly.
- **Validation** : Add lightweight validation (regex/email check, length limits) and return `400`.
- **Logging** : Switch to structured logger (`log/slog`, `zap`, or `zerolog`) with fields `{listener: i, email: message.Email}`.
- **HTML escaping** :`template.HTMLEscapeString` the user-supplied `message.Message` to avoid XSS in forwarded email.
