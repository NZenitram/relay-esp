# Relay ESP API

Relay ESP API is a Go-based backend service that manages Email Service Provider (ESP) integrations, user authentication, and event tracking for email campaigns.

## Features

- User authentication and management
- ESP integration management
- Event tracking and statistics
- RESTful API design
- JWT-based authentication
- Database integration (PostgreSQL)

## API Endpoints

### Public Routes

- `GET /health`: Health check endpoint
- `POST /login`: User login
- `POST /request-password-reset`: Request a password reset
- `POST /reset-password`: Reset user password

### Protected Routes (require JWT authentication)

#### User Management
- `GET /api/v1/users`: Get all users
- `GET /api/v1/users/{id}`: Get a specific user
- `PUT /api/v1/users/{id}`: Update a user
- `DELETE /api/v1/users/{id}`: Delete a user

#### Event Management
- `GET /api/v1/events`: Get all events
- `GET /api/v1/events/types`: Get available event types
- `GET /api/v1/events/{type}`: Get events by type
- `GET /api/v1/events/{provider}/{event}`: Get provider event stats by type

#### ESP Management
- `GET /api/v1/esps`: Get all ESPs
- `POST /api/v1/esps`: Create a new ESP
- `PUT /api/v1/esps/{id}`: Update an ESP
- `DELETE /api/v1/esps/{id}`: Delete an ESP
- `GET /api/v1/esps/{provider}/event-stats`: Get provider event stats

#### User Event Statistics
- `GET /api/v1/event-stats`: Get user event statistics

## Setup and Installation

1. Clone the repository
2. Install dependencies:
