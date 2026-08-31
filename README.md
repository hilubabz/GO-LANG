# Chirpy API

A RESTful chirp (short-message) API built with **Go**, **PostgreSQL**, **JWT authentication**, and **sqlc**.

The application provides user registration and authentication, chirp creation and retrieval, refresh-token management, user updates, chirp deletion, an admin metrics endpoint, and a Polka webhook for upgrading users to Chirpy Red.

## Features

- User registration with password hashing
- JWT-based access-token authentication
- Refresh-token authentication and revocation
- Create and validate chirps
- Automatic filtering of blocked words in chirps
- Retrieve all chirps
- Filter chirps by author
- Sort chirps by creation time
- Retrieve a chirp by ID
- Delete chirps with ownership authorization
- Update authenticated user information
- Polka webhook integration for Chirpy Red upgrades
- Admin metrics for file-server hits
- Development-only database reset endpoint
- PostgreSQL database access through generated `sqlc` queries
- Request logging middleware
- Health-check endpoint

## Tech Stack

- **Go**
- **net/http**
- **PostgreSQL**
- **sqlc**
- **JWT** — `github.com/golang-jwt/jwt/v5`
- **UUID** — `github.com/google/uuid`
- **bcrypt/password hashing**
- **godotenv**
- **lib/pq**

The application uses a PostgreSQL-backed `database.Queries` instance and separates authentication functionality into an `internal/auth` package. fileciteturn0file0L15-L20

## Project Structure

```text
.
├── main.go
├── internal/
│   ├── auth/
│   │   └── ...
│   └── database/
│       └── ...
├── middleware/
│   └── ...
├── sql/
│   └── ...
├── .env
├── go.mod
└── README.md
```

> The exact contents of the `internal`, `middleware`, and SQL directories are not shown in the provided source file, but the application imports and uses these packages.

## Environment Variables

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
JWT_SECRET=your-secret-key
POLKA_KEY=your-polka-api-key
PLATFORM=dev
```

### Variables

| Variable | Description |
|---|---|
| `DB_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret used to create and validate JWT access tokens |
| `POLKA_KEY` | API key used to authenticate Polka webhooks |
| `PLATFORM` | Application platform/environment. `dev` enables the reset endpoint |

The application loads `.env` with `godotenv`, reads `DB_URL`, and initializes the PostgreSQL connection when starting. fileciteturn0file0L513-L526

## Installation

### 1. Clone the repository

```bash
git clone <your-repository-url>
cd <project-directory>
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Configure PostgreSQL

Create a PostgreSQL database and set the connection string in `.env`.

### 4. Run database migrations

Run the SQL migrations used by the project before starting the API.

If you are using `sqlc`, regenerate the database access code after modifying SQL queries:

```bash
sqlc generate
```

### 5. Start the server

```bash
go run .
```

The server starts on:

```text
http://localhost:8080
```

The configured HTTP server listens on port `8080` and uses a logging middleware. fileciteturn0file0L560-L570

## API Endpoints

### Health Check

```http
GET /api/healthz
```

Returns:

```text
OK
```

---

### Create User

```http
POST /api/users
```

Request:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Creates a user after hashing the supplied password.

Response:

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

---

### Login

```http
POST /api/login
```

Request:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Response includes:

- User information
- JWT access token
- Refresh token

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "jwt-access-token",
  "refresh_token": "refresh-token"
}
```

The login handler verifies the password, creates a JWT, generates a refresh token, and stores the refresh token in PostgreSQL. fileciteturn0file0L259-L340

---

### Create Chirp

```http
POST /api/chirps
Authorization: Bearer <access-token>
```

Request:

```json
{
  "body": "Hello, Chirpy!"
}
```

A chirp cannot exceed **140 characters**.

The API also replaces the following words with `****`:

```text
kerfuffle
fornax
sharbert
```

Both lowercase and capitalized versions are handled. fileciteturn0file0L124-L149

Example response:

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "Hello, Chirpy!",
  "user_id": "uuid"
}
```

---

### Get Chirps

```http
GET /api/chirps
```

Returns all chirps.

#### Filter by author

```http
GET /api/chirps?author_id=<user-uuid>
```

#### Sort chirps

Ascending:

```http
GET /api/chirps?sort=asc
```

Descending:

```http
GET /api/chirps?sort=desc
```

The endpoint supports both author filtering and `asc`/`desc` sorting. fileciteturn0file0L175-L228

---

### Get Chirp by ID

```http
GET /api/chirps/{chirpId}
```

Example:

```http
GET /api/chirps/550e8400-e29b-41d4-a716-446655440000
```

Returns the requested chirp.

---

### Delete Chirp

```http
DELETE /api/chirps/{chirpId}
Authorization: Bearer <access-token>
```

Only the user who owns the chirp can delete it.

The handler validates the JWT, retrieves the chirp, compares the chirp's `UserID` with the authenticated user ID, and deletes it only when they match. fileciteturn0file0L443-L474

---

### Update User

```http
PUT /api/users
Authorization: Bearer <access-token>
```

Request:

```json
{
  "email": "new@example.com",
  "password": "new-password"
}
```

The authenticated user's JWT identifies which account should be updated.

---

### Refresh Access Token

```http
POST /api/refresh
Authorization: Bearer <refresh-token>
```

Returns a new access token:

```json
{
  "token": "new-jwt-access-token"
}
```

The refresh endpoint looks up the refresh token in the database and creates a new JWT for its associated user. fileciteturn0file0L346-L368

---

### Revoke Refresh Token

```http
POST /api/revoke
Authorization: Bearer <refresh-token>
```

Revokes the supplied refresh token.

Response:

```http
204 No Content
```

---

### Polka Webhook

```http
POST /api/polka/webhooks
Authorization: ApiKey <polka-api-key>
```

Expected request:

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid"
  }
}
```

When the event is `user.upgraded`, the associated user is upgraded to Chirpy Red.

Other events are acknowledged without performing an upgrade. fileciteturn0file0L477-L510

---

## Admin Endpoints

### Metrics

```http
GET /admin/metrics
```

Returns an HTML page showing the number of times the `/app/` file server has been visited.

Example:

```text
Welcome, Chirpy Admin
Chirpy has been visited 10 times!
```

The hit counter is implemented using `atomic.Int32`. fileciteturn0file0L24-L46

### Reset

```http
POST /admin/reset
```

This endpoint is available only when:

```env
PLATFORM=dev
```

It deletes users and resets the file-server hit counter.

> **Warning:** This is a destructive development endpoint. Do not expose it in production.

---

## Static File Server

The application serves files through:

```http
/app/
```

The `/app/` prefix is stripped before serving files from the project directory. Each request increments the file-server hit counter. fileciteturn0file0L529-L547

---

## Authentication

Protected endpoints use the `Authorization` header:

```http
Authorization: Bearer <token>
```

The application:

1. Extracts the bearer token.
2. Validates the JWT using `JWT_SECRET`.
3. Obtains the authenticated user's UUID.
4. Performs the requested operation.

JWT validation is used when creating chirps and updating/deleting authenticated resources. fileciteturn0file0L108-L122

### Token Types

**Access Token**

Used to authenticate API requests.

**Refresh Token**

Stored in PostgreSQL and used to obtain a new access token. Refresh tokens can also be revoked.

---

## HTTP Status Codes

Common responses include:

| Status | Meaning |
|---|---|
| `200 OK` | Successful request |
| `201 Created` | Resource successfully created |
| `204 No Content` | Successful operation with no response body |
| `400 Bad Request` | Invalid request/data |
| `401 Unauthorized` | Missing or invalid authentication |
| `403 Forbidden` | Authenticated but not permitted |
| `404 Not Found` | Requested resource does not exist |
| `500 Internal Server Error` | Server-side failure |

---

## Running in Development

Set:

```env
PLATFORM=dev
```

Then run:

```bash
go run .
```

The server will be available at:

```text
http://localhost:8080
```

Useful endpoints during development:

```text
GET  /api/healthz
GET  /admin/metrics
POST /admin/reset
```

## Example Authentication Flow

```text
Register
   │
   ▼
POST /api/users
   │
   ▼
Login
   │
   ▼
POST /api/login
   │
   ├── Access Token
   │
   └── Refresh Token
           │
           ▼
       POST /api/refresh
           │
           ▼
       New Access Token
```

For normal authenticated requests:

```text
Client
  │
  │ Authorization: Bearer <JWT>
  ▼
Chirpy API
  │
  ├── Extract token
  ├── Validate JWT
  ├── Identify user
  └── Execute database operation
```

## Example Chirp Flow

```text
POST /api/chirps
       │
       ▼
Validate JWT
       │
       ▼
Parse request body
       │
       ▼
Check length <= 140
       │
       ▼
Filter blocked words
       │
       ▼
Save chirp to PostgreSQL
       │
       ▼
Return created chirp
```

## Server Configuration

The HTTP server is configured with:

- Address: `:8080`
- Read timeout: `10 seconds`
- Write timeout: `10 seconds`
- Maximum header size: `1 MiB`

The application also wraps the main router with request logging middleware. fileciteturn0file0L560-L565

## Notes

- Keep `.env` out of version control.
- Never commit `JWT_SECRET` or `POLKA_KEY`.
- The reset endpoint should remain development-only.
- Use HTTPS when deploying the API publicly.
- Run migrations before starting the application against a new database.
- Regenerate `sqlc` code after changing SQL queries.
