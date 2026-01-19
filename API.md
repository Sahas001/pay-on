# Pay-On API

Base URL: `http://localhost:8080`

Notes
- IDs are UUID strings.
- Timestamps use RFC3339 (UTC).
- Pagination uses `limit` and `offset` query params.
- Amounts are decimal strings (example: `"10.50"`).
- Most endpoints require `Authorization: Bearer <token>`.

## Auth

Register
```
POST /auth/register
{
  "phone_number": "+9779812345678",
  "email": "user@example.com",
  "password": "strong-password"
}
```

Login
```
POST /auth/login
{
  "phone_number": "+9779812345678",
  "password": "strong-password"
}
```

## Wallets

Create wallet
```
POST /wallets
{
  "name": "Alice",
  "phone_number": "+9779812345678",
  "pin": "1234",
  "device_id": "device-001",
  "public_key": "pub-001"
}
```

List wallets
```
GET /wallets?limit=10&offset=0
```

Get wallet by ID
```
GET /wallets/{id}
```

Update wallet
```
PATCH /wallets/{id}
{
  "name": "New Name",
  "phone_number": "+9779812345678",
  "device_id": "device-002"
}
```

Balance
```
GET /wallets/{id}/balance
PATCH /wallets/{id}/balance
{
  "balance": "100.00"
}
POST /wallets/{id}/balance/increment
POST /wallets/{id}/balance/decrement
```

Trusted peers list
```
GET /wallets/{id}/peers/trusted
```

## Transfers

Transfer funds (atomic)
```
POST /transfers
{
  "from_wallet_id": "uuid",
  "to_wallet_id": "uuid",
  "amount": "10.50",
  "pin": "1234",
  "signature": "sig-demo",
  "nonce": 12345,
  "connection_type": "online",
  "metadata": { "note": "test" }
}
```

## Transactions

Create transaction (log only)
```
POST /transactions
{
  "from_wallet_id": "uuid",
  "to_wallet_id": "uuid",
  "amount": "10.50",
  "signature": "sig-demo",
  "nonce": 12345
}
```

List by status
```
GET /transactions/status/{status}?limit=10&offset=0
```

Recent
```
GET /transactions/recent?limit=10
```

Wallet transactions
```
GET /wallets/{id}/transactions?limit=10&offset=0
GET /wallets/{id}/transactions/sent
GET /wallets/{id}/transactions/received
```

## Peers

Upsert peer
```
POST /peers/upsert
{
  "wallet_id": "uuid",
  "peer_wallet_id": "uuid",
  "public_key": "pub-002",
  "connection_type": "online",
  "is_trusted": false
}
```

Trust peer by wallet
```
PATCH /wallets/{id}/peers/{peer_id}/trusted
{
  "is_trusted": true
}
```

## Sync logs

Create sync log
```
POST /sync-logs
{
  "transaction_id": "uuid",
  "wallet_id": "uuid",
  "status": "pending"
}
```

## Audit logs

List audit logs
```
GET /audit-logs?limit=10&offset=0
```

Balance history
```
GET /audit-logs/balance-history/{wallet_id}?limit=10
```

## Stats

```
GET /stats/system
```
