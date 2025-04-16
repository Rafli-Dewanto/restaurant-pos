# Cake Store API

This is a Cake Store API built with Go using Fiber as the web framework and GORM as the ORM. The application stores cake details such as title, description, rating, and image URL. It uses **MySQL (via Docker) as the primary database** and **SQLite for unit testing**.

## Features
- Create, read, update, and delete cakes.
- Retrieve all cakes sorted by rating and title.
- Uses MySQL as the database (via Docker container).
- Unit testing with SQLite.

## Prerequisites
- [Go](https://go.dev/doc/install) (v1.22+)
- [Docker](https://www.docker.com/get-started)

## Setup & Installation

### 1. Clone the repository
```sh
git clone https://github.com/Rafli-Dewanto/cakestore-be.git
cd cakestore-be
```

### 2. Create a `.env` file
```sh
touch .env
```
Add the following environment variables:
```env
MYSQL_ROOT_PASSWORD=password
MYSQL_DATABASE=cakestore
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_HOST=db
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

| Method | Endpoint       | Description         |
|--------|--------------|--------------------|
| POST   | `/cakes`      | Create a new cake  |
| GET    | `/cakes`      | Get all cakes (sorted by rating & title) |
| GET    | `/cakes/:id`  | Get a specific cake by ID |
| PUT    | `/cakes/:id`  | Update cake details |
| DELETE | `/cakes/:id`  | Delete a cake |

### Example Request: Create Cake
```sh
curl -X POST http://localhost:8080/cakes -H "Content-Type: application/json" -d '{
  "title": "Chocolate Cake",
  "description": "Delicious chocolate cake",
  "rating": 4,
  "image": "https://example.com/chocolate.jpg"
}'
```

## Running Tests

Unit tests use SQLite in-memory database.

```sh
go test ./...
```

To run tests with detailed output:
```sh
go test -v ./...
```

## License
This project is licensed under the MIT License.

