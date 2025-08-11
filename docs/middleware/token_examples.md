# Token Examples

Date: 2025-08-08 17:01 local

This guide shows how to call the protected API endpoints using:
- Postman (with a pre-request script that builds the signature)
- JavaScript fetch (Node.js and Browser)

It complements the middleware analysis by giving copy‑pasteable, working examples.

---

### Overview

Protected routes (see metal/kernel/router.go):
- POST /posts — list/filter posts
- GET /posts/{slug} — show a post by slug

Gateway paths differ by environment:
- Local (via Caddy local): http://localhost:8080/posts
- Production (via Caddy prod): https://oullin.io/api/posts

Required headers on every protected request:
- X-Request-ID: A unique ID per request (string)
- X-API-Username: Your account name (case-insensitive lookup)
- X-API-Key: Your public token (pk_...)
- X-API-Timestamp: Unix epoch seconds
- X-API-Nonce: Unique per request within TTL window
- X-API-Signature: hex(HMAC-SHA256(secret, canonical_request))

Canonical request string (exact order):
METHOD + "\n" + PATH + "\n" + SORTED_QUERY + "\n" + username + "\n" + public + "\n" + timestamp + "\n" + nonce + "\n" + sha256_hex(body)

Notes:
- METHOD must be uppercase.
- PATH must be escaped path (e.g., /posts/%7Bslug%7D if encoded).
- SORTED_QUERY: sort keys then values; join as key=value&...
- sha256_hex(body): hex SHA-256 of raw request body ("" for GET/DELETE by convention here).
- Time skew default ±5m, and nonce replay is blocked during TTL.

---

### Postman

Environment variables to set:
- baseUrl: Local: http://localhost:8080; Prod: https://oullin.io/api
- username: Your account name
- publicKey: Your pk_... value (plaintext)
- secretKey: Your sk_... value (plaintext)
- requestId: Optional; if absent, we’ll reuse the nonce

Pre-request Script (copy/paste into your request or collection):

pm.sendRequest = pm.sendRequest; // keep reference

```js
(function() {
  // Postman sandbox: CryptoJS is available via require
  const crypto = require('crypto-js');
  function sha256Hex(str: string): string { return crypto.SHA256(str || '').toString(crypto.enc.Hex); }
  function sortedQuery(u: string): string {
    const url = new URL(u);
    const keys = Array.from(url.searchParams.keys());
    keys.sort();
    const parts: string[] = [];
    for (const k of keys) {
      const vs = url.searchParams.getAll(k).sort();
      for (const v of vs) parts.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
    }
    return parts.join('&');
  }
  
  const method = pm.request.method.toUpperCase();
  const urlStr = pm.environment.get('baseUrl') + pm.request.url.getPathWithQuery();
  const urlObj = new URL(urlStr);
  const path = urlObj.pathname;
  const query = sortedQuery(urlStr);
  const username = pm.environment.get('username');
  const publicKey = pm.environment.get('publicKey');
  const secretKey = pm.environment.get('secretKey');
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const nonce = crypto.lib.WordArray.random(16).toString();
  const body = (method === 'GET' || method === 'DELETE') ? '' : (pm.request.body?.raw || '');
  const bodyHash = sha256Hex(body);
  const canonical = [method, path, query, username, publicKey, timestamp, nonce, bodyHash].join('\n');
  const signature = crypto.HmacSHA256(canonical, secretKey).toString();
  
  pm.request.headers.upsert({ key: 'X-Request-ID', value: pm.environment.get('requestId') || nonce });
  pm.request.headers.upsert({ key: 'X-API-Username', value: username });
  pm.request.headers.upsert({ key: 'X-API-Key', value: publicKey });
  pm.request.headers.upsert({ key: 'X-API-Timestamp', value: timestamp });
  pm.request.headers.upsert({ key: 'X-API-Nonce', value: nonce });
  pm.request.headers.upsert({ key: 'X-API-Signature', value: signature });
})();
```

Example requests:
- Local list posts
  - Method: POST
  - URL: {{baseUrl}}/posts
  - Body: raw JSON {}
- Local show post
  - Method: GET
  - URL: {{baseUrl}}/posts/hello
- Production list posts
  - Method: POST
  - URL: https://oullin.io/api/posts

Tips:
- Ensure your system clock is correct (NTP); skew is typically ±5 minutes.
- Do not reuse the same Nonce within the TTL window.
- Always include X-Request-ID.

---

### JavaScript fetch — Node.js (TypeScript + ESM)

```ts
import { createHash, createHmac, randomBytes } from 'node:crypto';

function sha256Hex(text: string): string {
  return createHash('sha256').update(text || '').digest('hex');
}

function sortedQuery(u: string): string {
  const url = new URL(u);
  const keys = Array.from(url.searchParams.keys()).sort();
  const pairs: string[] = [];
  for (const k of keys) {
    const vs = url.searchParams.getAll(k).sort();
    for (const v of vs) pairs.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
  }
  return pairs.join('&');
}

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

type SignedFetchParams = {
  baseUrl: string;
  path: string;
  method: HttpMethod;
  body?: unknown;
  username: string;
  publicKey: string;
  secretKey: string;
};

export async function signedFetch(params: SignedFetchParams): Promise<any> {
  const { baseUrl, path, method, body, username, publicKey, secretKey } = params;
  const url = new URL(path, baseUrl).toString();
  const u = new URL(url);
  const ts = Math.floor(Date.now() / 1000).toString();
  const nonce = randomBytes(16).toString('hex');
  const payload = body && (method === 'POST' || method === 'PUT') ? JSON.stringify(body) : '';
  const bodyHash = sha256Hex(payload);
  const canonical = [
    method.toUpperCase(),
    u.pathname,
    sortedQuery(u.toString()),
    username,
    publicKey,
    ts,
    nonce,
    bodyHash,
  ].join('\n');
  const signature = createHmac('sha256', secretKey).update(canonical).digest('hex');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Request-ID': nonce,
    'X-API-Username': username,
    'X-API-Key': publicKey,
    'X-API-Timestamp': ts,
    'X-API-Nonce': nonce,
    'X-API-Signature': signature,
  };

  const init: RequestInit = { method, headers, body: payload || undefined };
  const res = await fetch(u.toString(), init);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

// Example usage (Node 18+ has global fetch; run with: node --env-file=.env example.ts after transpile)
(async () => {
  const data = await signedFetch({
    baseUrl: 'http://localhost:8080', // local via Caddy
    path: '/posts',
    method: 'POST',
    body: {},
    username: process.env.API_USER as string,
    publicKey: process.env.API_PUBLIC as string,
    secretKey: process.env.API_SECRET as string,
  });
  console.log('posts:', data);
})();
```
---

### JavaScript fetch — Browser (Web Crypto, TypeScript)

```ts
async function sha256Hex(text: string): Promise<string> {
  const enc = new TextEncoder();
  const buf = await crypto.subtle.digest('SHA-256', enc.encode(text || ''));
  return Array.from(new Uint8Array(buf)).map(b => b.toString(16).padStart(2, '0')).join('');
}

function sortedQuery(u: string): string {
  const url = new URL(u);
  const keys = Array.from(url.searchParams.keys()).sort();
  const pairs: string[] = [];
  for (const k of keys) {
    const vs = url.searchParams.getAll(k).sort();
    for (const v of vs) pairs.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
  }
  return pairs.join('&');
}

async function hmacSha256Hex(secret: string, message: string): Promise<string> {
  const enc = new TextEncoder();
  const key = await crypto.subtle.importKey('raw', enc.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']);
  const sig = await crypto.subtle.sign('HMAC', key, enc.encode(message));
  return Array.from(new Uint8Array(sig)).map(b => b.toString(16).padStart(2, '0')).join('');
}

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

type SignedFetchParams = {
  baseUrl: string;
  path: string;
  method: HttpMethod;
  body?: unknown;
  username: string;
  publicKey: string;
  secretKey: string;
};

async function signedFetch(params: SignedFetchParams): Promise<any> {
  const { baseUrl, path, method, body, username, publicKey, secretKey } = params;
  const url = new URL(path, baseUrl).toString();
  const u = new URL(url);
  const ts = Math.floor(Date.now() / 1000).toString();
  const nonce = crypto.getRandomValues(new Uint8Array(16)).reduce((s, b) => s + b.toString(16).padStart(2, '0'), '');
  const payload = body && (method === 'POST' || method === 'PUT') ? JSON.stringify(body) : '';
  const bodyHash = await sha256Hex(payload);
  const canonical = [method.toUpperCase(), u.pathname, sortedQuery(u.toString()), username, publicKey, ts, nonce, bodyHash].join('\n');
  const signature = await hmacSha256Hex(secretKey, canonical);
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Request-ID': nonce,
    'X-API-Username': username,
    'X-API-Key': publicKey,
    'X-API-Timestamp': ts,
    'X-API-Nonce': nonce,
    'X-API-Signature': signature,
  };
  const res = await fetch(u.toString(), { method, headers, body: payload || undefined, mode: 'cors' });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

// Local example (Vite dev server at http://localhost:5173)
(async () => {
  const data = await signedFetch({
    baseUrl: 'http://localhost:8080',
    path: '/posts',
    method: 'POST',
    body: {},
    username: 'your-account',
    publicKey: 'pk_...',
    secretKey: 'sk_...',
  });
  console.log('posts', data);
})();
```

### Notes:
- In production, prefix routes with /api (e.g., https://oullin.io/api/posts).
- Do not reuse Nonces within TTL; replay requests are rejected.
- X-Request-ID is required for tracing.
- Your canonicalization must match the server (method upper-casing, path escaping, SortedQuery, body hash).
