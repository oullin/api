# 🔐 Authentication

This API uses JSON Web Tokens (JWT) for stateless authentication.

## Example

1. Obtain a token from the authentication endpoint:

```bash
curl -X POST http://localhost:8080/login \
  -d '{"username":"alice","password":"secret"}'
# => {"token":"<JWT_TOKEN>"}
```

2. Use the token to call a protected API endpoint:

```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/protected
```

The JWT middleware validates the token and exposes its claims to handlers via the request context.
