***
name: code-audit
description: Performs comprehensive software audits with structured analysis of security, performance, architecture, and code quality
***

# Code Audit Skill

## Purpose
Guide systematic code audits that produce actionable, prioritized findings.

## Analysis Framework

### 1. Security Assessment
- Authentication/authorization flows
- Input validation and sanitization
- Secret management (hardcoded keys, tokens)
- API security (rate limiting, auth)
- File upload/download handling
- SQL injection vectors
- XSS vulnerabilities
- CSRF protection
- Dependency vulnerabilities

### 2. Performance Analysis
- Database query optimization
- N+1 query patterns
- Inefficient algorithms (O(n²) or worse)
- Memory leaks
- Unnecessary API calls
- Missing indexes
- Caching opportunities
- Bundle size issues

### 3. Architecture Review
- Separation of concerns
- SOLID principles adherence
- Design pattern usage
- Dependency management
- Service boundaries
- State management
- Error handling strategy

### 4. Code Quality
- Code duplication (DRY violations)
- Cyclomatic complexity
- Naming conventions
- Magic numbers/strings
- Dead code
- Test coverage gaps
- Documentation quality

## Tool Usage Guidelines

- **Glob**: Start with `**/*.{js,ts,py,go,java}` to discover structure
- **Read**: Examine suspicious files in detail
- **Grep**: Find patterns like `eval(`, `exec(`, `password`, `TODO`, `FIXME`
- **Bash**: Run `git log --oneline -20` to understand recent changes

## Report Structure Template

```markdown
# [Project Name] Code Audit

## Executive Summary
[2-3 sentences: health, critical count, risk]

## Critical Issues 🔴
### [Issue Title]
- **File**: `path/to/file:line`
- **Severity**: Critical
- **Category**: Security/Performance/Quality
- **Description**: [Clear explanation]
- **Impact**: [What could happen]
- **Fix**: [Specific steps]

## Security Concerns 🟠
[Same format as above]

## Performance Observations 🟡
[Same format as above]

## Code Quality 🔵
[Same format as above]

## Architecture Notes 🟢
[Same format as above]

## Recommendations
1. **[Action]** - [Effort: hours] - [Priority: High/Med/Low]
2. ...
```

## Finding Quality Standards

✅ **Good Finding**:

```text
### SQL Injection in User Search
- **File**: `src/api/users.js:45`
- **Severity**: Critical
- **Category**: Security
- **Description**: Raw user input concatenated into SQL query without sanitization
- **Impact**: Attacker can execute arbitrary SQL, dump entire database, or delete data
- **Fix**: Replace with parameterized query:
  ```javascript
  db.query('SELECT * FROM users WHERE name = ?', [userInput])
  ```
```

❌ **Bad Finding**:

```text
### Code issue
There's a problem in the user file around line 45.
```

## Common Patterns to Check

### Security Red Flags
- `eval(userInput)`
- `exec(userInput)`
- `innerHTML = data`
- String concatenation in SQL
- Hardcoded credentials
- Missing CSRF tokens
- Unvalidated redirects

### Performance Red Flags
- Loops with DB queries
- Missing database indexes
- Synchronous I/O
- Large bundle sizes
- Unoptimized images
- Missing caching headers

### Quality Red Flags
- Functions > 50 lines
- Cyclomatic complexity > 10
- Duplicate code blocks
- Missing error handling
- No input validation
- Magic numbers
```
