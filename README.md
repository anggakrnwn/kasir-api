# Kasir API - Go POS System

Sebuah API Point of Sale (POS) sederhana yang dibangun dengan Go untuk mengelola produk dan transaksi penjualan.

## 📋 Endpoints & Responses

### 🔧 Health Check
**GET** `/health`

**Response:**
```json
{
  "status": "OK",
  "message": "API Running"
}
```

### 🛒 Produk

**GET** `/api/product`
Get semua produk dengan optional filtering
- Query params: `?name=` (filter by name)
- Response: `200 OK`
```json
[
  {
    "id": 1,
    "name": "Indomie Goreng",
    "price": 3000,
    "stock": 50
  }
]
```

**POST** `/api/product`
Create produk baru
- Request body:
```json
{
  "name": "Mie Sedap",
  "price": 3500,
  "stock": 25
}
```
- Response: `201 Created`
```json
{
  "id": 6,
  "name": "Mie Sedap",
  "price": 3500,
  "stock": 25
}
```

**GET** `/api/product/{id}`
Get produk by ID
- Response: `200 OK`
```json
{
  "id": 1,
  "name": "Indomie Goreng",
  "price": 3000,
  "stock": 50
}
```

**PUT** `/api/product/{id}`
Update produk
- Request body:
```json
{
  "name": "Indomie Goreng Updated",
  "price": 3500,
  "stock": 40
}
```
- Response: `200 OK`
```json
{
  "id": 1,
  "name": "Indomie Goreng Updated",
  "price": 3500,
  "stock": 40
}
```

**DELETE** `/api/product/{id}`
Delete produk
- Response: `200 OK`
```json
{
  "message": "Product deleted successfully"
}
```

### 💰 Transaksi

**POST** `/api/checkout`
Process checkout transaction
- Request body:
```json
{
  "items": [
    {"product_id": 1, "quantity": 2},
    {"product_id": 2, "quantity": 3},
    {"product_id": 3, "quantity": 1}
  ]
}
```
- Response: `200 OK`
```json
{
  "id": 1,
  "total_amount": 14000,
  "created_at": "2024-01-20T10:30:00Z",
  "details": [
    {
      "id": 1,
      "transaction_id": 1,
      "product_id": 1,
      "product_name": "Indomie Goreng",
      "quantity": 2,
      "subtotal": 6000
    },
    {
      "id": 2,
      "transaction_id": 1,
      "product_id": 2,
      "product_name": "Aqua 600ml",
      "quantity": 3,
      "subtotal": 9000
    }
  ]
}
```

### 📊 Laporan

**GET** `/api/report/hari-ini`
Get today's sales summary
- Response: `200 OK`
```json
{
  "total_revenue": 45000,
  "total_transaksi": 5,
  "produk_terlaris": {
    "nama": "Indomie Goreng",
    "qty_terjual": 12
  }
}
```

**GET** `/api/report?start_date=2024-01-01&end_date=2024-01-31`
Get sales report by date range
- Query params:
  - `start_date` (required): YYYY-MM-DD format
  - `end_date` (required): YYYY-MM-DD format
- Response: `200 OK`
```json
{
  "total_revenue": 150000,
  "total_transaksi": 15,
  "produk_terlaris": {
    "nama": "Aqua 600ml",
    "qty_terjual": 45
  }
}
```

## 🚀 Quick Start

1. Setup PostgreSQL database
2. Configure `.env` file:
```env
PORT=8080
DB_CONN=postgres://user:pass@localhost:5432/kasir_db
```
3. Run the server:
```bash
go run main.go
```

## ⚠️ Error Responses

**400 Bad Request:**
```json
{
  "error": "Invalid request body"
}
```

**404 Not Found:**
```json
{
  "error": "Product not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Insufficient stock for product 'Indomie Goreng'. Available: 5, Requested: 10"
}
```

**405 Method Not Allowed:**
```json
{
  "error": "Method not allowed"
}
```

---

## Contributing

This project is developed as part of CodeWithUmam. While contributions are welcome, please note this is primarily a learning project.

## License

This project is created for educational purposes as part of CodeWithUmam.

---

**Built with ❤️ using Go**