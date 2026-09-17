# Run locally

This project runs a Vue frontend, a Go API, and a local PostgreSQL database. The Go API also calls OneStepGPS and uses Cloudflare R2 for device icons.

## Prerequisites

- Go 1.27.1 or newer
- Node.js 22.18 or newer in the 22.x line, or 24.12 or newer, and npm
- Docker with Docker Compose
- The [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- OneStepGPS API credentials, a Google Maps JavaScript API key, and Cloudflare R2 credentials

## 1. Start PostgreSQL

From the repository root:

```sh
docker compose up -d db
docker compose exec db pg_isready -U app -d app
```

The database is available at `localhost:54320`. Wait until `pg_isready` reports that it is accepting connections before running the migration.

## 2. Configure and run the API

In a terminal:

```sh
cd backend
test -f .env || cp .env.example .env
```

Edit `backend/.env`. Keep the example `DATABASE_URL` for the Docker database. Set `ONE_STEP_GPS_DEVICE_URL` to `https://track.onestepgps.com/v3/api/public/device` and set `ONE_STEP_GPS_DEVICE_KEY` to your API key. Fill in all five `R2_*` variables; the server requires them at startup, even if you do not upload an icon. Leave `CORS_ALLOWED_ORIGIN` as `http://localhost:5173` and `PORT` as `8080`.

Then export the variables, apply the migration, and start the API:

```sh
set -a
source .env
set +a
migrate -path migrations -database "$DATABASE_URL" up
go run .
```

Go does not load `.env` automatically. Keep this terminal running. `http://localhost:8080/health` should return `ok`.

## 3. Configure and run the frontend

In a second terminal, from the repository root:

```sh
cd frontend
test -f .env || cp .env.example .env
```

Edit `frontend/.env` and set `VITE_GOOGLE_MAPS_API_KEY` to your browser key. Set `VITE_GOOGLE_MAPS_MAP_ID` to your map ID if you have one; the example uses `DEMO_MAP_ID`. Keep `VITE_API_BASE_URL=http://localhost:8080`.

```sh
npm ci
npm run dev
```

Open `http://localhost:5173`. Vite loads `frontend/.env` automatically; restart the dev server after changing it. Both `.env` files are ignored by Git.

To stop the database later, run `docker compose down` from the repository root. This keeps the PostgreSQL data volume for the next run.
