# URL Management System

A lightweight Go service that manages URL redirects with caching and proxy support.

## Features
- RESTful API for CRUD operations on redirects
- Redis caching for fast look‑ups
- MongoDB persistence
- Optional proxying of destination URLs
- Swagger UI for interactive API docs
- Health‑check endpoint
- Environment‑based configuration

## Architecture
```
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  Controller   │──▶│   Service     │──▶│  Repository   │
└───────────────┘   └───────────────┘   └───────────────┘
      ▲                 ▲                 ▲
      │                 │                 │
┌─────┴─────┐   ┌─────┴─────┐   ┌─────┴─────┐
│  HTTP     │   │  MongoDB   │   │  Redis    │
└───────────┘   └───────────┘   └───────────┘
```

## Getting Started
```bash
# Clone
git clone https://github.com/fernandoglatz/url-management.git
cd url-management

# Install dependencies
go mod tidy

# Configure
cp conf/application.yml.example conf/application.yml
# edit the file as needed

# Run locally
docker-compose up -d
go run main.go
```

## API Endpoints
- `GET /redirect` – list all redirects
- `GET /redirect/{id}` – get a redirect
- `PUT /redirect` – create a redirect
- `PUT /redirect/{id}` – update a redirect
- `DELETE /redirect/{id}` – delete a redirect
- `GET /` – execute a redirect (query param `to`)
- `GET /health` – health check
- `GET /swagger-ui/*any` – Swagger UI

## Configuration
See `conf/application.yml` for server, MongoDB, Redis, and logging settings.

## Testing
```bash
go test ./...
```

## Docker
```bash
docker build -t url-management .
docker run -p 8080:8080 url-management
```

## License
MIT © 2025

# Detailed Description

This section provides a more in-depth look at the components and workings of the URL Management System.

## Features

### RESTful API
The system exposes a RESTful API that allows clients to perform CRUD (Create, Read, Update, Delete) operations on URL redirects. This API is the primary interface for interacting with the URL management system.

### Redis Caching
Redis is used as a caching layer to provide fast look‑ups for redirects. This helps in reducing the load on the MongoDB database and speeds up the response time for clients.

### MongoDB Persistence
MongoDB is used as the primary data store for persisting URL redirects. It provides a flexible schema design that can evolve with the application.

### Proxying of Destination URLs
The system can optionally proxy requests to the destination URLs. This means that the client can be served directly from the URL Management System without having to make an additional request to the destination server.

### Swagger UI
Swagger UI is integrated into the system to provide interactive API documentation. It allows developers to explore and test the API endpoints directly from the browser.

### Health‑check Endpoint
A health-check endpoint is available for monitoring the status of the service. It can be used to check if the service is up and running.

### Environment‑based Configuration
The system supports environment‑based configuration, allowing different settings for development, testing, and production environments.

## Architecture

The architecture of the URL Management System is designed to be simple and lightweight, yet powerful enough to handle the required functionality.

### Controller
The controller is responsible for handling incoming HTTP requests and returning the appropriate HTTP responses. It acts as the entry point for the RESTful API.

### Service
The service contains the business logic of the application. It processes the data received from the controller and interacts with the repository to perform the necessary operations.

### Repository
The repository is responsible for data persistence. It interacts with MongoDB to store and retrieve URL redirects.

### HTTP
The HTTP component handles the incoming HTTP requests and outgoing HTTP responses. It is used by the controller to communicate with the clients.

### MongoDB
MongoDB is used as the primary data store for persisting URL redirects. It provides a flexible schema design that can evolve with the application.

### Redis
Redis is used as a caching layer to provide fast look‑ups for redirects. It helps in reducing the load on the MongoDB database and speeds up the response time for clients.

## Getting Started

To get started with the URL Management System, follow these steps:

1. Clone the repository:
    ```bash
    git clone https://github.com/fernandoglatz/url-management.git
    cd url-management
    ```

2. Install the dependencies:
    ```bash
    go mod tidy
    ```

3. Configure the application:
    ```bash
    cp conf/application.yml.example conf/application.yml
    ```
    Edit the `conf/application.yml` file as needed to configure the server, MongoDB, Redis, and logging settings.

4. Run the application locally:
    ```bash
    docker-compose up -d
    go run main.go
    ```

The application should now be running locally and you can access the API endpoints as described in the API Endpoints section.

## API Endpoints

The following API endpoints are available in the URL Management System:

- `GET /redirect` – List all redirects
- `GET /redirect/{id}` – Get a specific redirect by ID
- `PUT /redirect` – Create a new redirect
- `PUT /redirect/{id}` – Update an existing redirect by ID
- `DELETE /redirect/{id}` – Delete a redirect by ID
- `GET /` – Execute a redirect (requires query param `to`)
- `GET /health` – Health check
- `GET /swagger-ui/*any` – Swagger UI for interactive API documentation

## Configuration

The configuration for the URL Management System is done through the `conf/application.yml` file. This file contains settings for the server, MongoDB, Redis, and logging. There is an example configuration file `conf/application.yml.example` that can be copied to `conf/application.yml` and edited as needed.

## Testing

To run the tests for the URL Management System, use the following command:
```bash
go test ./...
```
This will execute all the tests in the project and report any failures.

## Docker

The URL Management System can be built and run as a Docker container. To build the Docker image, use the following command:
```bash
docker build -t url-management .
```
This will build the Docker image with the tag `url-management`.

To run the Docker container, use the following command:
```bash
docker run -p 8080:8080 url-management
```
This will run the Docker container and map port 8080 from the container to port 8080 on the host, allowing access to the API endpoints.

## License

The URL Management System is licensed under the MIT License. See the LICENSE file for more information.
