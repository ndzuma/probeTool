---
name: security-audit
description: Performs comprehensive security and performance audits of codebases, identifying vulnerabilities, unsafe patterns, security weaknesses, and performance bottlenecks. Use this skill when the user requests a security assessment, penetration testing analysis, vulnerability scan, or performance review of their code. Generates detailed, actionable reports with severity classifications and remediation guidance.
---

This skill guides systematic security and performance auditing of software projects. The user provides a codebase directory. Analyze every file for security vulnerabilities, unsafe patterns, credential exposure, exploitable weaknesses, and performance issues.

## Analysis Framework

Before auditing, understand the threat model:

- **Attack Surface**: What does this application expose? (API endpoints, file uploads, user input, network services)
- **Threat Actors**: Who might attack this? (External hackers, malicious insiders, automated bots)
- **Critical Assets**: What must be protected? (User data, API keys, financial info, PII)
- **Performance Profile**: What are the performance-critical paths? (Database queries, API calls, file I/O)

Then conduct systematic analysis across ALL files using these tools:
- **Glob**: Discover all source files (`**/*.{go,js,py,java,ts,jsx,tsx}`)
- **Read**: Examine suspicious files in detail
- **Grep**: Find security-critical patterns (credentials, SQL, eval, exec, shell commands)
- **Bash**: Check git history for leaked secrets (`git log -S "password"`), analyze file sizes

## Security Categories

### 1. **Injection Vulnerabilities**
- SQL Injection (string concatenation in queries)
- Command Injection (unsanitized input to shell)
- Code Injection (eval, exec with user input)
- Path Traversal (file operations with user paths)
- LDAP/XML/NoSQL injection

**Detection patterns**:
```
Grep for: "exec(", "eval(", "system(", "os.system", "subprocess.call"
Check: User input flows to these functions without validation
```

### 2. **Authentication & Authorization**
- Hardcoded credentials (passwords, API keys, tokens)
- Weak password policies
- Missing authentication checks
- Privilege escalation vectors
- Session management flaws
- JWT vulnerabilities (weak signing, no expiry)

**Detection patterns**:
```
Grep for: "password", "api_key", "secret", "token", "Authorization"
Check: Plaintext storage, weak hashing, missing auth middleware
```

### 3. **Cryptography**
- Weak hashing algorithms (MD5, SHA1)
- Insecure encryption (DES, ECB mode)
- Hardcoded crypto keys
- Insufficient key length
- Missing TLS/HTTPS enforcement
- Certificate validation disabled

**Detection patterns**:
```
Grep for: "md5", "sha1", "DES", "ECB", "InsecureSkipVerify"
Check: Crypto library usage, key generation, TLS config
```

### 4. **Data Exposure**
- Sensitive data in logs
- API responses leaking internal info
- Error messages exposing stack traces
- Unencrypted data at rest
- Missing data sanitization
- PII handling violations

**Detection patterns**:
```
Grep for: "console.log", "fmt.Printf", "print(", "logger."
Check: What's being logged? Passwords, tokens, PII?
```

### 5. **Input Validation**
- Missing input sanitization
- Insufficient type checking
- Buffer overflows (C/C++)
- Integer overflow/underflow
- Regex DoS (ReDoS)
- XML external entities (XXE)

**Detection patterns**:
```
Check: Every user input point (HTTP params, form data, file uploads)
Validate: Type checking, length limits, whitelist validation
```

### 6. **Access Control**
- Missing authorization checks
- Insecure direct object references (IDOR)
- Broken access control
- Privilege escalation
- CORS misconfigurations
- Missing rate limiting

**Detection patterns**:
```
Check: API endpoints - who can call them?
Verify: Resource access checks before operations
```

### 7. **Configuration & Deployment**
- Debug mode in production
- Default credentials
- Unnecessary services exposed
- Missing security headers
- Open ports and services
- Dependency vulnerabilities

**Detection patterns**:
```
Grep for: "debug = true", "DEBUG=1", "development"
Check: .env files, docker-compose.yml, config files
```

## Performance Categories

### 8. **Database Performance**
- N+1 query patterns
- Missing database indexes
- Inefficient queries (SELECT *)
- Unbounded result sets (no LIMIT)
- Connection pool exhaustion
- Missing query caching

**Detection patterns**:
```
Grep for: "SELECT \\*", "for.*db.Query", "range.*Query"
Check: Queries inside loops, missing pagination
```

### 9. **Algorithm Efficiency**
- O(n²) or worse algorithms
- Inefficient data structures
- Redundant computations
- Nested loops with DB/API calls
- Unnecessary sorting/filtering

**Detection patterns**:
```
Check: Nested loops, recursive functions without memoization
Look for: Duplicate processing, inefficient string operations
```

### 10. **Memory Management**
- Memory leaks (unclosed resources)
- Excessive memory allocation
- Large object retention
- Missing garbage collection hints
- Buffer size issues

**Detection patterns**:
```
Grep for: "new \\[", "make(", "malloc", "append"
Check: Resource cleanup (defer, finally, close)
```

### 11. **Network & I/O**
- Synchronous blocking operations
- Missing request timeouts
- No connection pooling
- Excessive API calls
- Large payload transfers
- Missing compression

**Detection patterns**:
```
Check: HTTP client configs, file I/O patterns
Look for: Sequential API calls that could be parallel
```

### 12. **Concurrency Issues**
- Race conditions
- Deadlocks
- Missing synchronization
- Unbounded goroutines/threads
- Channel blocking
- Inefficient locking

**Detection patterns**:
```
Grep for: "go func", "goroutine", "thread", "async"
Check: Shared state access, mutex usage, channel operations
```

## Audit Report Structure

Generate a markdown report with this EXACT structure:

```markdown
# Security & Performance Audit Report
## [Project Name]

**Generated**: [Date]
**Repository**: [Repo URL or path]
**Languages**: [Go, JavaScript, Python, etc.]

***

## Executive Summary

[2-3 sentences: overall security posture, performance bottlenecks, critical vulnerability count, risk level]

**Overall Security Risk**: [CRITICAL | HIGH | MEDIUM | LOW]
**Performance Rating**: [POOR | NEEDS IMPROVEMENT | GOOD | EXCELLENT]

***

## Findings Summary

| Issue | Severity | Type | Impact |
|-------|----------|------|--------|
| [Issue title] | Critical/High/Medium/Low | Security/Performance | [Brief impact] |
| ... | ... | ... | ... |

**Statistics**:
- Critical: X
- High: Y
- Medium: Z
- Low: W

***

## 🔴 Critical Security Vulnerabilities

### 1. [Vulnerability Title]
**File**: `path/to/file.ext:line`  
**Severity**: Critical  
**Type**: [Injection/Auth/Crypto/etc]  
**CWE**: [CWE-89 (SQL Injection) if applicable]

**Description**: Clear explanation of the vulnerability

**Vulnerable Code**:
\`\`\`[language]
[actual vulnerable code snippet]
\`\`\`

**Exploitation Scenario**:
[Step-by-step how an attacker exploits this]

**Impact**: [What happens if exploited - data breach, RCE, etc.]

**Remediation**:
\`\`\`[language]
[fixed code example]
\`\`\`

***

## 🟠 High Severity Issues

[Same format as Critical - both security and performance]

***

## 🟡 Medium Severity Issues

[Same format as Critical - both security and performance]

***

## 🔵 Low Severity Issues

[Same format as Critical - both security and performance]

***

## ⚡ Performance Bottlenecks

### Database Performance
- **N+1 Queries**: [File locations where this occurs]
- **Missing Indexes**: [Tables that need indexes]
- **Unbounded Results**: [Queries without LIMIT]

### Algorithm Inefficiency
- **O(n²) Operations**: [Functions with nested loops]
- **Redundant Computations**: [Repeated calculations]

### Memory Issues
- **Potential Leaks**: [Unclosed resources]
- **Large Allocations**: [Excessive memory usage]

### Network & I/O
- **Blocking Operations**: [Synchronous I/O locations]
- **Missing Timeouts**: [Network calls without timeouts]

***

## ✅ Security Strengths

[List what the codebase does RIGHT]
- Proper use of parameterized queries
- Strong password hashing (bcrypt)
- Good input validation in module X
- TLS/HTTPS properly configured

***

## ✅ Performance Strengths

[List what the codebase does RIGHT]
- Efficient caching strategy
- Good use of connection pooling
- Proper indexing on critical tables
- Optimized algorithms in hot paths

***

## Testing Recommendations

### Security Testing
**Automated Tools**:
\`\`\`bash
# Static analysis
[Tool recommendations for detected languages]

# Dependency scanning
[Tool recommendations]

# SAST/DAST tools
[Tool recommendations]
\`\`\`

**Manual Testing Checklist**:
- [ ] Test for SQL injection on all inputs
- [ ] Test authentication bypass techniques
- [ ] Test authorization bypass (IDOR)
- [ ] Test input validation boundaries
- [ ] Verify access controls on all endpoints

### Performance Testing
**Load Testing**:
\`\`\`bash
# Example load test commands
[Tool recommendations: k6, ab, wrk, etc.]
\`\`\`

**Profiling**:
- [ ] Profile CPU usage under load
- [ ] Profile memory allocation patterns
- [ ] Analyze database query performance
- [ ] Check for memory leaks
- [ ] Monitor goroutine/thread counts

***

## Files Analyzed

[MANDATORY: List ALL files you scanned]
\`\`\`
path/to/file1.go ✓
path/to/file2.js ✓
path/to/file3.py ✓
...
\`\`\`

**Total**: [X] files across [Y] directories

***

**Report End**
```

## Quality Standards

### ✅ Good Security Finding:
```markdown
### SQL Injection in User Search
**File**: `api/users.go:45`
**Severity**: Critical
**Type**: Injection - SQL
**CWE**: CWE-89

**Description**: The `SearchUsers()` function concatenates unsanitized user input directly into a SQL query string.

**Vulnerable Code**:
\`\`\`go
query := "SELECT * FROM users WHERE name = '" + userInput + "'"
db.Query(query)
\`\`\`

**Exploitation Scenario**:
1. Attacker sends: `'; DROP TABLE users; --`
2. Query becomes: `SELECT * FROM users WHERE name = ''; DROP TABLE users; --'`
3. Database executes DROP TABLE
4. All user data is deleted

**Impact**: Complete database compromise, data loss, denial of service

**Remediation**:
\`\`\`go
query := "SELECT * FROM users WHERE name = ?"
db.Query(query, userInput)
\`\`\`
```

### ✅ Good Performance Finding:
```markdown
### N+1 Query Pattern in User List
**File**: `api/posts.go:23-28`
**Severity**: High
**Type**: Performance - Database

**Description**: The `ListPosts()` function queries the database once per user in a loop, causing N+1 query problem.

**Inefficient Code**:
\`\`\`go
for _, post := range posts {
    user, _ := db.Query("SELECT * FROM users WHERE id = ?", post.UserID)
    // N queries in loop
}
\`\`\`

**Impact**: With 1000 posts, executes 1001 database queries. Response time scales linearly with post count. At 10k posts, 10+ second response time.

**Remediation**:
\`\`\`go
// Single query with JOIN
query := `SELECT posts.*, users.* FROM posts 
          LEFT JOIN users ON posts.user_id = users.id`
rows, _ := db.Query(query)
\`\`\`

**Performance Gain**: Reduces 1001 queries to 1. Response time: 10s → 100ms (100x improvement)
```

### ❌ Bad Finding:
```markdown
### Security Issue
There's a problem in the database code. Fix it.
```

## Critical Rules

1. **NEVER HALLUCINATE**: Only report vulnerabilities you actually found in the code
2. **SCAN EVERYTHING**: Use Glob to find ALL files, Read each one
3. **BE SPECIFIC**: Include exact file paths and line numbers
4. **SHOW PROOF**: Include actual vulnerable code snippets
5. **NO GUESSING**: If you can't confirm a vulnerability, don't report it
6. **NO FEATURE REQUESTS**: Only report security/performance issues in current code
7. **NO TIME ESTIMATES**: Don't estimate fix time - just explain the fix
8. **QUANTIFY PERFORMANCE**: Show impact (e.g., "10s → 100ms", "1000 queries → 1")

## When No Issues Found

If the codebase is genuinely secure and performant:

```markdown
## Security Assessment: PASS ✅
## Performance Assessment: EXCELLENT ✅

After comprehensive analysis of [X] files across [Y] directories, **no exploitable vulnerabilities or significant performance bottlenecks were identified**.

### Security Strengths:
- Proper input validation throughout
- Secure authentication implementation
- Strong cryptographic practices
- Good separation of concerns
- Principle of least privilege applied

### Performance Strengths:
- Efficient database queries with proper indexing
- Good use of caching strategies
- Optimized algorithms
- Proper connection pooling
- No memory leaks detected

### Recommendations for Continued Excellence:
- Maintain dependency updates
- Regular security audits
- Implement automated security testing
- Monitor for new CVEs in dependencies
- Continue load testing under production-like conditions
```

**DO NOT** invent vulnerabilities or performance issues to fill a report. Absence of findings is a valid audit result.
