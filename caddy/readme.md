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
values (for example: `</assets/index-CQed9K_I.css>; rel=preload; as=style, </assets/index-sqsQjSaJ.js>; rel=modulepreload; as=script`).
When this variable is defined Caddy will emit an HTTP `103 Early Hints`
response for eligible GET and HEAD requests before proxying traffic, allowing
clients to start fetching the referenced assets while the upstream response is
prepared. The local and production Caddyfiles attach a `handle_response`
interceptor to `reverse_proxy` so the proxy can send the Early Hints response
and then `copy_response` from the upstream service without interrupting the
request flow.
