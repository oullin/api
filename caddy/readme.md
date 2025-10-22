# Debugging

### Headers
```text
header_down X-Debug-Username {http.request.header.X-API-Username}
header_down X-Debug-Key {http.request.header.X-API-Key}
header_down X-Debug-Signature {http.request.header.X-API-Signature}
```

### Early Hints

Set the `EARLY_HINTS_LINKS` environment variable to a comma-separated list of
[`Link` header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Link)
values (for example: `</assets/app.css>; rel=preload; as=style, </assets/app.js>; rel=modulepreload`).
When this variable is defined Caddy will emit an HTTP `103 Early Hints`
response for eligible GET and HEAD requests before proxying traffic, allowing
clients to start fetching the referenced assets while the upstream response is
prepared.
