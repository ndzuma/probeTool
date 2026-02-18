export function fullAuditPrompt(targetPath) {
  return `
Audit this codebase: ${targetPath}

Use the security-audit skill to guide your analysis.
`.trim();
}
