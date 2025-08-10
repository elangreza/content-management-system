# Content Management System

This is a Content Management System (CMS) built with Go, designed to manage articles, tags, users, and roles.

## Features

- User authentication (JWT-based)
- Role-based access control (RBAC) using bitwise operator for simplifying the logic
- Article and tag management
- basic User profile
- RESTful API endpoints
- Database migrations using golang-migrate
- Docker support for easy deployment

## Project Structure

- `cmd/server/` - Main application entry point and configuration
- `internal/entity/` - Core domain entities (Article, User, Tag, etc.)
- `internal/params/` - Request/response parameter definitions
- `internal/postgresql/` - Database access and SQL queries
- `internal/rest/` - HTTP handlers and middleware
- `internal/service/` - Business logic and services
- `internal/error/` - Custom error types
- `internal/constanta/` - Constants used throughout the project
- `internal/sharevar/` - Shared variables (e.g., user roles)
- `migrations/` - SQL migration scripts
- `docker-compose.yaml` - Docker Compose configuration
- `Makefile` - Common development commands

## Getting Started

### Prerequisites

- go
- Docker & Docker Compose
- PostgreSQL

### Setup

1. Simply run this command:
   ```bash
   make up
   ```
2. let's explore the app using docs [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
