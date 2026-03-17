# Shipment Tracking gRPC Microservice

A shipment tracking microservice built with Go, gRPC, and PostgreSQL following **Clean Architecture** and **Domain-Driven Design** (DDD) principles.

## Architecture

The project follows a strict layered architecture where dependencies only point inward:

```
Infrastructure → Application → Domain
```

### Domain Layer (`internal/domain/`)

Pure business logic with zero external dependencies.

| File | Description |
|---|---|
| `entity/shipment.go` | **Shipment** aggregate root — factory method with validation, status transition logic via `AddStatusEvent()` |
| `entity/status_event.go` | **StatusEvent** entity — records each status change with location, notes, and timestamp |
| `entity/log.go` | **Log** entity — audit trail recording action, payload, and timestamp |
| `valueobject/status.go` | **Status** value object — defines valid statuses (`pending`, `picked_up`, `in_transit`, `delivered`, `cancelled`), allowed transitions, and terminal state checks |
| `valueobject/money.go` | **Money** value object — stores monetary values as cents (int64) to avoid floating-point issues |
| `repository/shipment_repository.go` | Repository interfaces (ports) — `ShipmentRepository`, `StatusEventRepository`, `LogRepository` |
| `service/shipment_service.go` | Domain service — orchestrates `CreateShipment`, `GetShipment`, `UpdateStatus`, `GetEventHistory` with audit logging |

### Application Layer (`internal/application/`)

Thin orchestration layer between gRPC and domain.

| File | Description |
|---|---|
| `dto/request.go` | Input DTOs — `CreateShipmentRequest`, `UpdateStatusRequest` with Money conversion helpers |
| `dto/response.go` | Output converters — `ShipmentToProto()`, `StatusEventToProto()` mapping domain entities to protobuf messages |
| `usecase/shipment_usecase.go` | Use cases — parses raw string inputs (UUIDs, status strings) into domain types and delegates to domain service |

### Infrastructure Layer (`internal/infrastructure/`)

External concerns: database, gRPC, configuration.

| File | Description |
|---|---|
| `config/config.go` | Configuration loader — reads `config.yml` for Postgres mode, falls back to JSON file storage if config is missing |
| `grpc/server.go` | gRPC server — `Start()` and `Stop()` (graceful shutdown) |
| `grpc/handler/shipment_handler.go` | gRPC handler — implements all 4 RPCs, maps domain errors to gRPC status codes |
| `persistence/postgres/connection.go` | Postgres connection pool with health check |
| `persistence/postgres/shipment_repository.go` | Postgres `ShipmentRepository` implementation |
| `persistence/postgres/status_event_repository.go` | Postgres `StatusEventRepository` implementation |
| `persistence/postgres/log_repository.go` | Postgres `LogRepository` implementation (payload stored as JSON) |
| `persistence/postgres/migrations/*.sql` | Goose migrations for `shipments`, `status_events`, `logs` tables |
| `persistence/jsonfile/store.go` | JSON file storage engine — thread-safe read/write with per-entity directories |
| `persistence/jsonfile/shipment_repository.go` | JSON file `ShipmentRepository` adapter |
| `persistence/jsonfile/status_event_repository.go` | JSON file `StatusEventRepository` adapter |
| `persistence/jsonfile/log_repository.go` | JSON file `LogRepository` adapter |

### Proto (`proto/shipment/`)

| File | Description |
|---|---|
| `shipment.proto` | gRPC service contract — 4 RPCs: `CreateShipment`, `GetShipment`, `UpdateStatus`, `GetEventHistory` |
| `shipment.pb.go` | Generated protobuf message types |
| `shipment_grpc.pb.go` | Generated gRPC server/client interfaces |

### Entrypoint (`cmd/server/`)

| File | Description |
|---|---|
| `main.go` | Application entrypoint — loads config, creates repositories (Postgres or JSON), wires dependency injection, starts gRPC server, handles graceful shutdown |

## Status Transitions

```
pending → picked_up → in_transit → delivered
  ↓           ↓            ↓
cancelled  cancelled   cancelled
```

- `delivered` and `cancelled` are **terminal states** — no further transitions allowed.
- Skipping states (e.g. `pending` → `in_transit`) is not allowed.

## gRPC API

| RPC | Description |
|---|---|
| `CreateShipment` | Create a new shipment with driver details, origin/destination, and monetary values |
| `GetShipment` | Retrieve a shipment by UUID |
| `UpdateStatus` | Transition shipment to a new status with location and notes |
| `GetEventHistory` | Get the full status event timeline for a shipment |

### Error Mapping

| Domain Error | gRPC Code |
|---|---|
| `ErrShipmentNotFound` | `NOT_FOUND` |
| `ErrDuplicateReference` | `ALREADY_EXISTS` |
| `ErrInvalidTransition` | `FAILED_PRECONDITION` |
| `ErrShipmentTerminated` | `FAILED_PRECONDITION` |

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 16+ (or use Docker)
- [goose](https://github.com/pressly/goose) (for migrations)
- [grpcurl](https://github.com/fullstorydev/grpcurl) (for manual testing)

### Local Development

**1. Configure the database**

Create `config.yml` in the project root:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: shipment
  sslmode: disable

grpc:
  port: 50051
```

> If `config.yml` is missing, the app automatically falls back to JSON file storage in `storage/tmp/`.

**2. Run migrations**

```bash
goose -dir internal/infrastructure/persistence/postgres/migrations \
  postgres "host=localhost port=5432 user=postgres password=postgres dbname=shipment sslmode=disable" up
```

**3. Start the server**

```bash
go run cmd/server/main.go
```

**4. Test with grpcurl**

```bash
# Create a shipment
grpcurl -plaintext -import-path proto/shipment -proto shipment.proto \
  -d '{
    "reference_number": "REF-001",
    "origin": "Almaty",
    "destination": "Astana",
    "driver_name": "John Doe",
    "driver_phone": "+77012345678",
    "unit_number": "TRK-001",
    "shipment_amount": 1000.50,
    "shipment_currency": "KZT",
    "driver_revenue": 800.00,
    "driver_revenue_currency": "KZT"
  }' localhost:50051 shipment.ShipmentService/CreateShipment

# Get a shipment
grpcurl -plaintext -import-path proto/shipment -proto shipment.proto \
  -d '{"id": "<SHIPMENT_UUID>"}' \
  localhost:50051 shipment.ShipmentService/GetShipment

# Update status
grpcurl -plaintext -import-path proto/shipment -proto shipment.proto \
  -d '{
    "shipment_id": "<SHIPMENT_UUID>",
    "status": "picked_up",
    "location": "Warehouse A",
    "notes": "Driver arrived"
  }' localhost:50051 shipment.ShipmentService/UpdateStatus

# Get event history
grpcurl -plaintext -import-path proto/shipment -proto shipment.proto \
  -d '{"shipment_id": "<SHIPMENT_UUID>"}' \
  localhost:50051 shipment.ShipmentService/GetEventHistory
```

## Docker Deployment

**Start everything** (Postgres → migrations → gRPC server):

```bash
docker compose up --build
```

**Stop:**

```bash
docker compose down
```

**Stop and wipe database:**

```bash
docker compose down -v
```

The `docker-compose.yml` defines 3 services:

| Service | Purpose |
|---|---|
| `postgres` | PostgreSQL 16 with health checks |
| `migrate` | One-shot goose migration runner (waits for Postgres to be healthy) |
| `app` | The gRPC server (waits for migrations to complete) |

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./internal/domain/...
go test -v ./internal/application/...
```

### Test Coverage

- **Value Objects** — status validity, allowed transitions, terminal state checks, status parsing
- **Entities** — shipment factory validation, `AddStatusEvent` transition logic and timestamp mutation
- **Domain Service** — create shipment flow (persistence + events + audit logs), status update flows (valid/invalid/terminal), duplicate reference detection
- **Use Cases** — DTO-to-domain conversion, money conversion to cents, delegation to domain service

## Project Structure

```
shipment/
├── cmd/server/main.go                 # Entrypoint
├── proto/shipment/                    # Protobuf definitions + generated code
├── internal/
│   ├── domain/                        # Pure business logic (no dependencies)
│   │   ├── entity/                    # Aggregate roots and entities
│   │   ├── valueobject/               # Value objects (Status, Money)
│   │   ├── repository/               # Repository interfaces (ports)
│   │   └── service/                   # Domain service
│   ├── application/                   # Orchestration layer
│   │   ├── dto/                       # Request/response mapping
│   │   └── usecase/                   # Use cases
│   └── infrastructure/                # External concerns
│       ├── config/                    # Configuration loading
│       ├── grpc/                      # gRPC server and handlers
│       └── persistence/              # Repository implementations
│           ├── postgres/              # PostgreSQL (+ migrations)
│           └── jsonfile/              # JSON file fallback
├── Dockerfile                         # Multi-stage Docker build
├── docker-compose.yml                 # Full stack deployment
├── config.yml                         # Local Postgres config
└── config.docker.yml                  # Docker Postgres config
```
