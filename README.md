# CRM Backend

Backend Go untuk CRM Enterprise dengan clean architecture, GORM/PostgreSQL,
Redis, REST, Telegram outbox, CSV/PDF export, dan isolasi multi-tenant.

## Struktur Clean Architecture

```text
internal/
├── domain/          # entity, kontrak domain, error, service helper
├── usecase/         # business flow per fitur
├── repository/      # implementasi persistence, misalnya PostgreSQL
├── infrastructure/  # database, cache, logger, JWT, Telegram
├── delivery/        # HTTP dan websocket adapter
└── utils/           # helper, validator, constant
```

Alur dependency utama:

```text
delivery -> usecase -> domain
repository -> domain
infrastructure -> domain
cmd/api -> wiring semua dependency
```

Alur membaca satu fitur:

```text
delivery/http/handler -> usecase -> domain/repositories -> repository/postgres -> PostgreSQL
```

`cmd/api` hanya bertugas memasangkan dependency, menjalankan migration,
menyalakan HTTP server dan worker background.

## Menjalankan Lokal

```bash
cp .env.example .env
docker compose up --build
```

REST tersedia pada `http://localhost:8080`, dan OpenAPI berada di
`api/swagger.yaml`.

Tanpa Docker:

```bash
go run ./cmd/migrate -direction up
go run ./cmd/api
```

Migration dijalankan dengan `github.com/go-gormigrate/gormigrate/v2`.
Command yang tersedia adalah `up`, `down`, dan `status`.

Lihat `endpoint-golang.md` untuk keputusan kontrak, model data, cache, RBAC,
dan catatan integrasi frontend.
