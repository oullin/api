# 🔐 Authentication

This API uses JSON Web Tokens (JWT) for stateless authentication. Tokens are signed
with the account's secret stored in the `api_keys` table and include the account name
in their claims.

## Example

1. Obtain a token from the authentication endpoint:

```bash
curl -X POST http://localhost:8080/auth/token \
  -d '{"account_name":"alice","secret_key":"secret"}'
# => {"token":"<JWT_TOKEN>"}
```

2. Use the token to call a protected API endpoint:

```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/protected
```

The JWT middleware validates the token using the matching secret from `api_keys` and exposes its
claims to handlers via the request context.
