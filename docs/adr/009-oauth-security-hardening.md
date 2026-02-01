# ADR-009: OAuth Security Hardening

## Status

Accepted

## Context

A security review of the OAuth implementation identified several vulnerabilities and areas for improvement:

1. **Apple OAuth**: JWT ID tokens were parsed without signature verification (`ParseUnverified`), allowing forged tokens
2. **Facebook OAuth**: Access tokens were passed in URL query parameters, risking exposure in logs and referrer headers
3. **No PKCE**: Authorization code interception attacks were possible without Proof Key for Code Exchange
4. **No OAuth Audit Logging**: OAuth operations (linking/unlinking identities) were not being audited
5. **Short State Expiry**: 10-minute OAuth state expiry was too short for real-world user flows
6. **No HTTP Timeouts**: OAuth HTTP requests used `http.DefaultClient` with no timeout

## Decision

### 1. Apple JWT Signature Verification

**Problem**: `jwt.ParseUnverified()` allows attackers to forge arbitrary JWT claims.

**Solution**: Implement proper JWT verification using Apple's JWKS endpoint:

```go
// Fetch Apple's public keys from JWKS endpoint
resp, err := oauthHTTPClient.Get("https://appleid.apple.com/auth/keys")

// Parse and verify JWT with:
// - Signature verification using Apple's public keys
// - Issuer validation (https://appleid.apple.com)
// - Audience validation (client_id)
// - Expiration check
token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
    // Get key from JWKS by kid
    return publicKey, nil
}, jwt.WithValidMethods([]string{"ES256"}),
   jwt.WithIssuer(appleIssuer),
   jwt.WithAudience(clientID),
   jwt.WithExpirationRequired())
```

**Key Caching**: Apple's public keys are cached for 24 hours with automatic refresh on key miss (key rotation).

### 2. Facebook Token in Authorization Header

**Problem**: Access token in URL query parameter is logged and leaked via referrer headers.

**Solution**: Use `Authorization: Bearer` header instead:

```go
// Before (vulnerable):
reqURL := facebookUserInfoURL + "?fields=...&access_token=" + token

// After (secure):
req.Header.Set("Authorization", "Bearer "+accessToken)
```

### 3. PKCE Implementation (RFC 7636)

**Problem**: Without PKCE, authorization codes can be intercepted and exchanged by attackers.

**Solution**: Implement PKCE for all providers that support it (Google, Facebook):

```go
// Generate PKCE challenge
type PKCEChallenge struct {
    Verifier  string // Stored server-side in Redis
    Challenge string // Sent to authorization server
    Method    string // "S256" (SHA-256)
}

func GeneratePKCE() (*PKCEChallenge, error) {
    verifier := base64.RawURLEncoding.EncodeToString(randomBytes)
    hash := sha256.Sum256([]byte(verifier))
    challenge := base64.RawURLEncoding.EncodeToString(hash[:])
    return &PKCEChallenge{Verifier: verifier, Challenge: challenge, Method: "S256"}, nil
}
```

**Flow**:
1. Generate verifier and challenge
2. Store verifier in `OAuthStateData` (Redis)
3. Include `code_challenge` and `code_challenge_method=S256` in auth URL
4. Pass `code_verifier` in token exchange request

**Provider Support**:
- Google: Full PKCE support
- Facebook: Full PKCE support
- Apple: Not supported with `form_post` response mode

### 4. OAuth Audit Logging

**Problem**: OAuth operations were not being logged for security auditing.

**Solution**: Add audit logging using existing audit infrastructure:

| Event | Action | When |
|-------|--------|------|
| OAuth Registration | `self_registered` | New user created via OAuth |
| Identity Linked | `oauth_linked` | OAuth identity linked to user (explicit or auto-link) |
| Identity Unlinked | `oauth_unlinked` | OAuth identity removed from user |

**Details logged**:
- Provider name (google, facebook, apple)
- User ID
- Auto-link indicator (for automatic email-based linking)

### 5. Extended OAuth State Expiry

**Problem**: 10-minute state expiry is too short for users who:
- Get distracted during OAuth flow
- Have slow network connections
- Need to create accounts on the OAuth provider

**Solution**: Increase `oauthStateExpiry` from 10 minutes to 30 minutes.

### 6. Shared HTTP Client with Timeout

**Problem**: `http.DefaultClient` has no timeout, allowing slow/malicious OAuth providers to hang requests indefinitely.

**Solution**: Create shared HTTP client with 30-second timeout:

```go
var oauthHTTPClient = &http.Client{
    Timeout: 30 * time.Second,
}
```

All OAuth providers now use this client for:
- Token exchange requests
- User info requests
- JWKS fetching (Apple)

## Updated OAuthProvider Interface

```go
type OAuthProvider interface {
    Name() string
    GetAuthURL(state string, pkce *PKCEChallenge) string
    ExchangeCode(ctx context.Context, code, verifier string) (*OAuthTokens, error)
    GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)
    SupportsPKCE() bool
}
```

## Files Modified

| File | Changes |
|------|---------|
| `oauth_apple.go` | JWT signature verification with JWKS caching |
| `oauth_facebook.go` | Token in Authorization header, PKCE support |
| `oauth_google.go` | PKCE support |
| `oauth_provider.go` | Interface updated for PKCE |
| `oauth_http.go` | New file: shared HTTP client, PKCE generation |
| `token_store.go` | Added `PKCEVerifier` to `OAuthStateData` |
| `service_impl.go` | PKCE flow, audit logging, extended state expiry |
| `register.go` | Added audit service dependency |

## Consequences

### Positive

- **Apple OAuth**: Forged JWT tokens are now rejected with signature verification
- **Facebook OAuth**: Access tokens no longer leaked in URLs/logs
- **PKCE**: Authorization code interception attacks are prevented
- **Audit Trail**: All OAuth operations are now logged for security review
- **Reliability**: HTTP timeouts prevent hanging requests
- **User Experience**: Extended state expiry accommodates real-world OAuth flows

### Negative

- **JWKS Dependency**: Apple OAuth now depends on Apple's JWKS endpoint availability (mitigated by 24h cache)
- **Slightly Larger State**: PKCE verifier adds ~44 bytes to Redis state storage
- **Interface Change**: Existing `OAuthProvider` implementations need updating

### Neutral

- **No Apple PKCE**: Apple Sign In doesn't support PKCE with `form_post` response mode, but their JWT signature verification provides equivalent security

## Security Verification

After implementation, verify:
1. **Apple OAuth**: Test with forged JWT → should reject with signature error
2. **Facebook OAuth**: Check server logs → token should NOT appear in URLs
3. **PKCE**: Intercept OAuth callback → should fail without matching verifier
4. **Audit Logging**: Query `audit_logs` table → should see `oauth_linked`/`oauth_unlinked` events

## Related

- ADR-006: Self-Registration and OAuth Authentication (original OAuth design)
- [RFC 7636](https://tools.ietf.org/html/rfc7636): Proof Key for Code Exchange (PKCE)
- [Apple Sign In JWT](https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/verifying_a_user)
