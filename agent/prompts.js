export function fullAuditPrompt(targetPath) {
  return `
First, list what skills you have available.

Then audit this codebase: ${targetPath}

Use the security-audit skill to guide your comprehensive security and performance analysis.
`.trim()
}
