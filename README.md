# secret-santa-go-api

A Secret Santa API as an exercise of Go, challenge proposed by @LukeberryPi.

## Overview

This API allows users to create groups for Secret Santa, add participants to the groups, and run a draw to assign secret friends to each participant.

## Features

- Create users
- Create groups
- Add participants to groups
- Run a draw to assign secret friends
- Retrieve user and group information

## Setup
### Prerequisites

- Go 1.16 or higher
- Docker (optional, for running the database)

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/akctba/secret-santa-go-api.git
    cd secret-santa-go-api
    ```

2. Install dependencies:
    ```sh
    go mod download
    ```

### Running the API

1. Start the database (if using Docker):
    ```sh
    docker-compose up -d
    ```

2. Run the API:
    ```sh
    go run main.go
    ```

### Environment Variables

- `APP_ENV`: Runtime environment (`LOCAL`, `DEV`, `PROD`). If not set, defaults to `PROD`.
- `JWT_SECRET`: Required signing secret for bearer tokens in `DEV` and `PROD` (minimum 32 characters). In `LOCAL`, a development fallback secret is allowed when this variable is not set.
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed web origins for CORS (for example: `http://localhost:3000,https://app.example.com`).
    If this is not set, cross-origin browser requests are disabled.

3. The API will be available at `http://localhost:8080`.

### Running Tests

1. Run the tests:
    ```sh
    go test ./...
    ```