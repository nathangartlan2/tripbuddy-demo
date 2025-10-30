# API Security & Production Readiness Checklist

This document tracks security and production requirements for the TripBuddy API before public deployment.

## Critical Security Issues (URGENT)

### ❌ SQL Injection Vulnerabilities
**Priority: CRITICAL - Fix Immediately**

**Location:** `PostGresParksRepository.cs:56` in `CreateParkAsync`

**Issue:** Using string interpolation for SQL queries:
```csharp
VALUES ('{park.Name}', '{parkCode}', ...)
```

**Risk:** An attacker could send malicious input like:
```json
{"name": "'; DROP TABLE parks; --"}
```

**Fix Required:**
- Switch ALL queries to use parameterized queries with `@parameters`
- Review `CreateParkAsync`, `SearchGeographic`, and any other methods using string interpolation
- The `DeleteParkAsync` method already uses parameters correctly - follow that pattern

---

## Essential Requirements

### 1. ✅ HTTPS Configuration
**Status:** Not Implemented

**Requirements:**
- Configure Kestrel to use HTTPS in production
- Obtain SSL/TLS certificate (Let's Encrypt for free)
- Redirect HTTP to HTTPS
- Set HSTS headers

**Implementation:**
- For Fly.io: Handled automatically by platform
- For local dev: Use `dotnet dev-certs https --trust`

---

### 2. ❌ Configuration Management
**Status:** Partially Implemented

**Current Issues:**
- Database connection string has hardcoded fallback
- No secrets management

**Requirements:**
- Move all sensitive data to environment variables
- Use User Secrets for local development
- Use secure secret management in production (Fly.io Secrets, Azure Key Vault, etc.)
- No hardcoded passwords or connection strings

**Files to Update:**
- `Program.cs` - Remove hardcoded connection string fallback
- Create proper `appsettings.Development.json` and `appsettings.Production.json`

---

### 3. ❌ Authentication & Authorization
**Status:** Not Implemented

**Requirements:**
- Implement Basic Authentication middleware (minimum)
- Protect write endpoints: POST, PUT, DELETE
- Keep read endpoints (GET) public or add separate auth tier
- Store credentials securely (hashed, not plain text)

**Endpoints to Protect:**
- ✅ `GET /parks` - Public (read-only)
- ✅ `GET /park/{id}` - Public (read-only)
- ✅ `GET /park/search` - Public (read-only)
- 🔒 `POST /park` - Requires authentication
- 🔒 `PUT /park/{parkCode}` - Requires authentication
- 🔒 `DELETE /park/{id}` - Requires authentication

**Implementation Options:**
- Basic Auth (simple, good for internal/service-to-service)
- API Keys (better for rate limiting per client)
- JWT tokens (more complex, better for user-facing apps)

---

### 4. ❌ Rate Limiting
**Status:** Not Implemented

**Requirements:**
- Prevent abuse and DDoS attacks
- Limit requests per IP address
- Different limits for authenticated vs anonymous users
- Return 429 (Too Many Requests) when exceeded

**Recommended Limits:**
- Anonymous: 100 requests per 15 minutes
- Authenticated: 1000 requests per 15 minutes
- Write operations: Lower limits (e.g., 50 per 15 min)

**Implementation:**
```csharp
builder.Services.AddRateLimiter(options => { ... });
app.UseRateLimiter();
```

---

### 5. ❌ CORS Policy
**Status:** Not Implemented

**Current Risk:** API is either completely locked down or completely open

**Requirements:**
- Configure explicit allowed origins
- Don't use `AllowAnyOrigin()` in production
- Set allowed HTTP methods
- Configure allowed headers

**Implementation:**
```csharp
builder.Services.AddCors(options =>
{
    options.AddPolicy("ProductionPolicy", policy =>
    {
        policy.WithOrigins("https://yourdomain.com")
              .AllowAnyMethod()
              .AllowAnyHeader();
    });
});
```

---

### 6. ❌ Input Validation & Sanitization
**Status:** Minimal

**Requirements:**
- Validate all input fields (lengths, formats, ranges)
- Use Data Annotations on models
- Sanitize inputs to prevent XSS
- Validate latitude/longitude ranges
- Validate parkCode format
- Max length limits on strings

**Example:**
```csharp
[Required]
[StringLength(200, MinimumLength = 1)]
public string Name { get; set; }

[Range(-90, 90)]
public float Latitude { get; set; }
```

---

### 7. ❌ Error Handling & Logging
**Status:** Minimal

**Current Issues:**
- Detailed errors may expose internal structure
- No centralized error handling
- Limited logging

**Requirements:**
- Global exception handler middleware
- Return generic errors to clients
- Log detailed errors server-side only
- Don't expose stack traces in production
- Log all authentication failures
- Log unusual patterns (rate limit hits, etc.)

**Implementation:**
- Add `app.UseExceptionHandler()` middleware
- Integrate Serilog or similar structured logging
- Configure different log levels for dev/prod

---

### 8. ❌ Logging & Monitoring
**Status:** Basic console logging only

**Requirements:**
- Structured logging (Serilog recommended)
- Log aggregation service (seq, Application Insights, etc.)
- Track key metrics:
  - Failed authentication attempts
  - Rate limit violations
  - Error rates
  - Response times
  - Database connection issues

**Implementation:**
```csharp
builder.Host.UseSerilog((context, config) =>
{
    config.ReadFrom.Configuration(context.Configuration);
});
```

---

### 9. ❌ Health Check Endpoint
**Status:** Not Implemented

**Requirements:**
- Endpoint for monitoring services
- Check database connectivity
- Check external dependencies
- Return appropriate status codes

**Implementation:**
```csharp
builder.Services.AddHealthChecks()
    .AddNpgSql(connectionString);

app.MapHealthChecks("/health");
```

---

## Nice to Have (Future Enhancements)

### 10. API Versioning
- Allow breaking changes without disrupting existing clients
- `/v1/parks`, `/v2/parks` pattern
- Or use headers: `Accept: application/vnd.api.v1+json`

### 11. Request Size Limits
- Prevent large payload attacks
- Configure `MaxRequestBodySize`
- Limit JSON payload sizes

### 12. Database Connection Pooling
- Already handled by Npgsql by default
- Verify settings for production load
- Configure min/max pool sizes

### 13. Response Caching
- Cache frequently accessed data (park lists)
- Use `ResponseCaching` middleware
- Set appropriate cache headers
- Consider Redis for distributed caching

### 14. API Documentation
- ✅ Swagger UI already implemented
- Add XML comments to endpoints for better docs
- Add example requests/responses
- Document authentication requirements

### 15. Compression
- Enable response compression for better performance
- Gzip or Brotli compression
- Reduces bandwidth usage

---

## Implementation Priority

### Phase 1: Critical Security (Do First)
1. ✅ Fix SQL injection vulnerabilities
2. ✅ Add authentication/authorization
3. ✅ Move secrets to configuration
4. ✅ Add basic input validation

### Phase 2: Production Essentials
5. ✅ Configure HTTPS
6. ✅ Add CORS policy
7. ✅ Implement rate limiting
8. ✅ Add error handling middleware
9. ✅ Set up logging

### Phase 3: Monitoring & Reliability
10. ✅ Add health checks
11. ✅ Configure monitoring/alerting
12. ✅ Load testing

### Phase 4: Optimization (After Launch)
13. ✅ Add caching
14. ✅ API versioning
15. ✅ Performance tuning

---

## Testing Checklist

Before going live, test:
- [ ] SQL injection attempts are blocked
- [ ] Unauthenticated users cannot POST/PUT/DELETE
- [ ] Rate limiting triggers correctly
- [ ] CORS blocks unauthorized origins
- [ ] HTTPS redirects work
- [ ] Error messages don't leak sensitive info
- [ ] Health endpoint responds correctly
- [ ] Large payloads are rejected
- [ ] Invalid input is rejected with helpful messages

---

## Deployment Checklist

- [ ] All secrets in environment variables (not code)
- [ ] HTTPS enforced
- [ ] Database connection string secured
- [ ] Authentication enabled
- [ ] Rate limiting configured
- [ ] CORS policy set
- [ ] Logging configured and working
- [ ] Health checks responding
- [ ] Error handling tested
- [ ] Backup strategy in place
- [ ] Monitoring/alerts configured

---

## Notes

- This API will be deployed to Fly.io
- PostgreSQL database also on Fly.io
- Go scraper will authenticate to POST parks
- Public can read parks but not modify

---

*Last Updated: 2025-10-30*
