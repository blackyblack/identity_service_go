# identity_service_go

Identity service in Go language - HTTP REST API skeleton with basic endpoints.

## Installation

```bash
go mod download
```

## Running the Service

```bash
go run src/main.go
```

The service will start on port 8080.

## API Endpoints

### POST /vouch

Accepts a JSON body with the following fields:
- `from` (string, required) - Source user
- `signature` (string, required) - Cryptographic signature
- `nonce` (string, required) - Unique nonce for the request
- `to` (string, required) - Target user

Example request:
```bash
curl -X POST http://localhost:8080/vouch \
  -H "Content-Type: application/json" \
  -d '{
    "from": "user1",
    "signature": "sig123",
    "nonce": "nonce456",
    "to": "user2"
  }'
```

Example response:
```json
{
  "success": true,
  "message": "Vouch accepted"
}
```

### GET /idt/:user

Retrieves user identity information.

Example request:
```bash
curl http://localhost:8080/idt/testuser
```

Example response:
```json
{
  "user": "testuser"
}
```
