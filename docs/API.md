# API Documentation

The API uses a token-based authentication mechanism. Most endpoints require a valid signature/token.

## Authentication

### Generate Signature
**Public Endpoint**
Generates a signature that can be used for authenticated requests.

- **URL**: `POST /generate-signature`
- **Body**:
  ```json
  {
    "username": "string",
    "timestamp": 1234567890
  }
  ```
- **Response**:
  ```json
  {
    "signature": "string",
    "max_tries": 10,
    "cadence": {
      "received_at": "timestamp",
      "created_at": "timestamp",
      "expires_at": "timestamp"
    }
  }
  ```

## Posts & Categories

### List Posts
**Auth Required**
Retrieves a paginated list of posts.

- **URL**: `POST /posts`
- **Body** (Filters - all optional):
  ```json
  {
    "title": "string",
    "author": "string",
    "category": "string",
    "tag": "string",
    "text": "string"
  }
  ```
- **Response**: List of posts objects with pagination metadata.

### Get Post
**Auth Required**
Retrieves a single post by its slug.

- **URL**: `GET /posts/{slug}`
- **Response**: Post object.

### List Categories
**Auth Required**
Retrieves all categories.

- **URL**: `GET /categories`
- **Response**: List of category objects.

## Static Data
**Auth Required**

These endpoints return static JSON data used to populate various sections of the website.

- `GET /profile`
- `GET /experience`
- `GET /projects`
- `GET /social`
- `GET /talks`
- `GET /education`
- `GET /recommendations`

## System & Monitoring

### Health Check
**Auth Required**
Checks if the API is responsive.

- **URL**: `GET /ping`

### Database Health Check
**Auth Required**
Checks if the database connection is active.

- **URL**: `GET /ping-db`

### Metrics
**Internal Access Only**
Prometheus metrics endpoint.

- **URL**: `GET /metrics`
