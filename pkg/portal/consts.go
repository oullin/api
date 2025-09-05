package portal

const DatesLayout = "2006-01-02 15:04:05"
const MaxSignaturesTries = 10

// ---- Middleware / HTTP

const TokenHeader = "X-API-Key"
const UsernameHeader = "X-API-Username"
const SignatureHeader = "X-API-Signature"
const TimestampHeader = "X-API-Timestamp"
const NonceHeader = "X-API-Nonce"
const RequestIDHeader = "X-Request-ID"
const IntendedOriginHeader = "X-API-Intended-Origin"

// ---- Middleware / Context

type contextKey string

const AuthAccountNameKey contextKey = "auth.account_name"
const RequestIdKey contextKey = "request.id"
