# Scalable URL Shortener (Microservices)

## Services
- `api-gateway` (public entry):
  - `POST /api/v1/urls`
  - `GET /{code}`
- `url-service`:
  - `POST /api/v1/urls`
- `redirect-service`:
  - `GET /{code}`
- `db-service` (internal data APIs):
  - `POST /internal/urls`
  - `GET /internal/urls/by-long?longUrl=...`
  - `GET /internal/urls/{code}`
  - `POST /internal/urls/{code}/visit`
  - `GET /internal/stats/{code}`

## Local run (without Docker)
One-command start:

```bash
cd /home/user/Desktop/GO-Project
./run-local.sh
```

Then open:

```bash
http://localhost:8080/app
```

Manual start in separate terminals:

```bash
cd db-service && go run ./main.go
cd url-service && go run ./main.go
cd redirect-service && go run ./main.go
cd api-gateway && API_KEY=local-dev-key go run ./main.go
```

## Docker Compose

```bash
docker compose up --build
```

## Example usage
Open web app:

```bash
http://localhost:8080/app
```

Use the form to generate short URLs from the browser.

API usage:

Create short URL:

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: local-dev-key' \
  -d '{"longUrl":"https://example.com/docs/architecture","ttlDays":30}'
```

Follow redirect:

```bash
curl -i http://localhost:8080/<code>
```

Read stats from DB service:

```bash
curl http://localhost:8083/internal/stats/<code>
```
