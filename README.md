# CakeStore Backend

A backend service for CakeStore, built with Go, Fiber, and GORM.  
Supports customer registration, authentication, reservations, orders, and payment integration (Midtrans).

## Features

- Customer registration, login, and profile management
- JWT-based authentication
- Reservation system:
  - Create, update, delete, and view reservations
  - Reservation can be made with or without a table (`table_id` is optional; relation to "tables" is only created if provided)
- Order and payment management
  - Midtrans integration for payment processing
  - Payment status and notification handling
- Admin endpoints for managing customers and reservations

## Project Structure

```
internal/
  config/         # Configuration loading
  constants/      # Project-wide constants
  database/       # Database connection and migration
  delivery/
    http/         # HTTP handlers/controllers
  domain/
    entity/       # GORM models/entities
    model/        # Request/response models (including Midtrans)
  repository/     # Data access layer
  usecase/        # Business logic
test/             # Test suites
```

## Reservation Logic

- When creating a reservation, if `table_id` is provided in the request payload, the reservation will be linked to the specified table and table availability will be checked.
- If `table_id` is omitted or zero, the reservation will not be linked to any table.

## Payment Integration

- Uses Midtrans for payment processing.
- Payment models and notification structs are up-to-date with Midtrans API.
- Handles payment status updates and notifications.

## Running the Project

1. **Clone the repository**
2. **Configure environment variables** (see `.env.example`)
3. **Run database migrations** (auto-migrated on startup)
4. **Start the server**
   ```bash
   go run main.go
   ```

## Running Tests

```bash
go test ./test/...
```

## API Documentation

- See [docs/](docs/) or use the included Postman collection.

## License

MIT
| Method | Endpoint | Description |
| ------ | ---------------- | ---------------------------- |
| GET | `/customers/me` | Get current customer profile |
| PUT | `/customers/:id` | Update customer profile |

#### Cake Routes

| Method | Endpoint     | Description         |
| ------ | ------------ | ------------------- |
| GET    | `/cakes`     | Get all cakes       |
| GET    | `/cakes/:id` | Get a specific cake |
| POST   | `/cakes`     | Create a new cake   |
| PUT    | `/cakes/:id` | Update cake details |
| DELETE | `/cakes/:id` | Delete a cake       |

#### Order Routes

| Method | Endpoint      | Description              |
| ------ | ------------- | ------------------------ |
| POST   | `/orders`     | Create a new order       |
| GET    | `/orders`     | Get all customer orders  |
| GET    | `/orders/:id` | Get specific order by ID |

### Example Requests

#### Register Customer

```sh
curl --location 'http://127.0.0.1:8080/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "name",
    "email": "test4@email.com",
    "password": "password123@Q",
    "address": "sudirman"
}'
```

#### Create Cake (Protected Route)

```sh
curl --location 'http://127.0.0.1:8080/cakes' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjdXN0b21lcl9pZCI6MSwiZW1haWwiOiJ0ZXN0NEBlbWFpbC5jb20iLCJleHAiOjE3NDQ4MTI1Nzd9.c5nWXiG0yf97LnLDnZjP_N7YJ8hQmxHII-RAgetmI3Q' \
--data '{
    "title": "Lemon cheesecake",
    "description": "A cheesecake made of lemon",
    "rating": 7,
    "image": "https://img.taste.com.au/ynYrqkOs/w720-h480-cfill-q8"
}'
```

## Running Tests

Unit tests.

```sh
go test ./...
```

To run tests with detailed output:

```sh
go test -v ./...
```

## License

This project is licensed under the MIT License.
