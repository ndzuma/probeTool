export function fullAuditPrompt(targetPath) {
  return `
Perform a comprehensive software audit of this repository: ${targetPath}

**Use the code-audit skill** to structure your analysis systematically.

## Your Task

1. **Discover Structure**: Use Glob to understand the codebase layout
2. **Analyze Security**: Find auth issues, injection risks, secret leaks
3. **Review Performance**: Identify bottlenecks, N+1 queries, inefficiencies
4. **Assess Architecture**: Evaluate design patterns, separation of concerns
5. **Check Code Quality**: Find duplication, complexity, anti-patterns

## Output Requirements

Generate a **detailed markdown report** with these sections:

### Executive Summary
2-3 sentences: overall health, critical issue count, risk level

### Critical Issues 🔴
Issues requiring immediate attention with:
- File path and line number
- Severity justification
- Specific fix steps

### Security Concerns 🟠
Vulnerabilities with exploitation scenarios

### Performance Observations 🟡
Bottlenecks with optimization approaches

### Code Quality 🔵
Maintainability issues with refactoring suggestions

### Architecture Notes 🟢
Design improvements and pattern recommendations

### Recommendations
Prioritized action items (1-10) with effort estimates

## Format Each Finding As:
- **File**: \`path/to/file:line\`
- **Severity**: Critical/High/Medium/Low
- **Description**: Clear explanation
- **Impact**: What could happen
- **Fix**: Specific remediation steps

Begin your audit now.
`.trim()
}
