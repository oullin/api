# Caddy

### Debugging Headers
```html
header_down X-Debug-Username {http.request.header.X-API-Username}
header_down X-Debug-Key {http.request.header.X-API-Key}
header_down X-Debug-Signature {http.request.header.X-API-Signature}
```
