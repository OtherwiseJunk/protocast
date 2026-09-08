// atproto login is backend-mediated: the Go backend holds the OAuth confidential
// client (via indigo's atproto/auth/oauth) and does handle resolution, PDS
// discovery, PAR, token exchange, DPoP and refresh. This package's job is just
// the client side of that: kick off the browser-based login flow against the
// backend, store the resulting session, and attach it to API requests.

export {};
