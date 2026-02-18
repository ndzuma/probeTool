export function fullAuditPrompt(targetPath) {
  return `
Perform a comprehensive security and performance audit of this codebase: ${targetPath}

Use the security-audit skill to guide your analysis.

Requirements:
1. Scan EVERY file in the repository
2. Only report vulnerabilities you actually find in the code
3. Include exact file paths and line numbers
4. Show actual code snippets as proof
5. Follow the report structure defined in the security-audit skill

Begin the audit now.
`.trim()
}
