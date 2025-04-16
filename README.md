# Cake Store API

This is a Cake Store API built with Go using Fiber as the web framework and GORM as the ORM. The application provides functionality for managing cakes, customers, orders, and payments. It uses **MySQL (via Docker) as the primary database** and **SQLite for unit testing**.

## Features

- User authentication and authorization (JWT)
- Customer registration and profile management
- Create, read, update, and delete cakes
- Order management system
- Payment integration with Midtrans
- Uses MySQL as the database (via Docker container)
- Unit testing

## Prerequisites

- [Go](https://go.dev/doc/install) (v1.22+)
- [Docker](https://www.docker.com/get-started)

## Setup & Installation

### 1. Clone the repository

```sh
git clone https://github.com/Rafli-Dewanto/lsp-programmer.git cakestore-be
cd cakestore-be
```

### 2. Create a `.env` file

```sh
touch .env.example .env
```

### 3. Run with Docker

```sh
docker compose up -d --build
```

### 4. Install dependencies

```sh
go mod tidy
```

### 5. Start the application

```sh
go run main.go
```

The API will be available at `http://localhost:8080`.

## API Endpoints

### Public Routes

| Method | Endpoint                | Description              |
| ------ | ----------------------- | ------------------------ |
| POST   | `/register`             | Register a new customer  |
| POST   | `/login`                | Customer login           |
| POST   | `/payment/notification` | Midtrans payment webhook |

### Protected Routes (Requires JWT Authentication)

#### Customer Routes

| Method | Endpoint         | Description                  |
| ------ | ---------------- | ---------------------------- |
| GET    | `/customers/me`  | Get current customer profile |
| PUT    | `/customers/:id` | Update customer profile      |

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
