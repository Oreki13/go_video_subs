# go_video_subs

REST API sederhana untuk manajemen video berbasis tier subscription, dibangun dengan Go dan Clean Architecture.

## Fitur

- **Autentikasi** — Register & login dengan JWT
- **Manajemen Video** — Akses video dibatasi berdasarkan tier subscription user (gold / silver / bronze)
- **Subscription** — User memiliki satu subscription aktif yang dikaitkan ke tier tertentu
- **Payment Gateway (Mock)** — Simulasi alur pembayaran: inisiasi transaksi → callback → aktivasi subscription
- **Unit Test** — Test usecase layer dengan mock dependency (tanpa database)

## Arsitektur

Project mengikuti **Clean Architecture** dengan pemisahan layer yang jelas:

```
go_video_subs/
├── cmd/                        # Entry point CLI (cobra)
│   ├── root.go
│   └── serve.go                # Bootstrap semua dependency & jalankan server
├── config/                     # Konfigurasi dari environment variable
├── internal/
│   ├── domain/                 # Entity & interface (inti bisnis, tidak bergantung layer lain)
│   │   ├── payment/
│   │   ├── subscription/
│   │   ├── user/
│   │   └── video/
│   ├── repository/             # Implementasi akses database (MariaDB via sqlx)
│   │   ├── payment/
│   │   ├── subscription/
│   │   ├── user/
│   │   └── video/
│   ├── usecase/                # Business logic
│   │   ├── payment/            # InitiatePayment, HandleCallback + unit test
│   │   ├── user/               # Register, Login + unit test
│   │   └── video/              # List video sesuai tier
│   └── delivery/
│       └── http/
│           ├── handler/        # Request/response handling (Fiber)
│           ├── middleware/     # JWT auth middleware
│           └── router/         # Definisi route
├── pkg/
│   ├── database/               # Koneksi MariaDB
│   ├── jwt/                    # Helper JWT
│   └── response/               # Format respons standar
├── scheme.sql                  # Skema database
└── .env.example
```

### Alur Payment Mock

```
POST /api/v1/payments/initiate
        │
        ▼
  Buat PaymentTransaction (status: pending)
  Return: external_payment_id + mock_callback_url
        │
        ▼
POST /api/v1/payments/callback
  { transaction_id, status: "success" | "failed" }
        │
        ├── success → Update status + Upsert Subscription (status: active)
        └── failed  → Update status saja
```

### Aturan Akses Video per Tier

| Tier User | Bisa Akses          |
|-----------|---------------------|
| gold      | gold, silver, bronze |
| silver    | silver, bronze       |
| bronze    | bronze               |
| (none)    | tidak bisa akses     |

## Prerequisites

- Go 1.21+
- MariaDB / MySQL

## Cara Menjalankan

**1. Clone & install dependency**
```bash
git clone <repo-url>
cd go_video_subs
go mod tidy
```

**2. Setup database**
```bash
mysql -u root -p < scheme.sql
```

**3. Konfigurasi environment**
```bash
cp .env.example .env
# Edit .env sesuai konfigurasi lokal
```

```env
APP_ENV=development
APP_PORT=8080

DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_NAME=go_video_subs

JWT_SECRET=secret
JWT_EXPIRY_HOURS=6
```

**4. Jalankan server**
```bash
go run main.go serve

# Override port (opsional)
go run main.go serve --port 9000
```

## Menjalankan Unit Test

```bash
go test ./...

# Dengan coverage
go test ./... -cover
```

## API Endpoints

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| POST | `/api/v1/users/register` | ✗ | Daftar akun baru |
| POST | `/api/v1/auth/login` | ✗ | Login, dapat JWT |
| GET | `/api/v1/videos` | ✓ | List video sesuai tier |
| POST | `/api/v1/payments/initiate` | ✓ | Mulai transaksi pembayaran |
| POST | `/api/v1/payments/callback` | ✗ | Callback mock payment gateway |

> Koleksi Postman tersedia di `video_subs.postman_collection.json`

## Tech Stack

- **Framework**: [Fiber v3](https://gofiber.io/)
- **Database**: MariaDB + [sqlx](https://github.com/jmoiron/sqlx)
- **Auth**: JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt))
- **CLI**: [Cobra](https://github.com/spf13/cobra)
- **Testing**: [testify](https://github.com/stretchr/testify)
