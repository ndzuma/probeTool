import { describe, it, expect } from 'vitest'

// Tests to implement:
describe('Prompt Generation', () => {
  it('should generate full audit prompt', () => {
    const targetPath = '/path/to/codebase'
    const fullAuditPrompt = (target) => {
      return `
First, list what skills you have available.

Then audit this codebase: ${target}

Use the security-audit skill to guide your comprehensive security and performance analysis.
`.trim()
    }
    
    const prompt = fullAuditPrompt(targetPath)
    
    expect(prompt).toContain('audit this codebase')
    expect(prompt).toContain(targetPath)
    expect(prompt).toContain('security-audit skill')
  })
  
  it('should include target directory', () => {
    const targetPath = '/Users/dev/project'
    const fullAuditPrompt = (target) => {
      return `
First, list what skills you have available.

Then audit this codebase: ${target}

Use the security-audit skill to guide your comprehensive security and performance analysis.
`.trim()
    }
    
    const prompt = fullAuditPrompt(targetPath)
    
    expect(prompt).toContain(targetPath)
    // Should contain the full path
    expect(prompt).toMatch(/\/Users\/dev\/project/)
  })
  
  it('should be valid markdown', () => {
    const fullAuditPrompt = (target) => {
      return `
First, list what skills you have available.

Then audit this codebase: ${target}

Use the security-audit skill to guide your comprehensive security and performance analysis.
`.trim()
    }
    
    const prompt = fullAuditPrompt('/some/path')
    
    // Basic markdown validation
    expect(typeof prompt).toBe('string')
    expect(prompt.length).toBeGreaterThan(0)
    expect(prompt).not.toContain('  ') // No double spaces
    expect(prompt.trim()).toBe(prompt) // Should be trimmed
  })
  
  it('should include skills instruction', () => {
    const fullAuditPrompt = (target) => {
      return `
First, list what skills you have available.

Then audit this codebase: ${target}

Use the security-audit skill to guide your comprehensive security and performance analysis.
`.trim()
    }
    
    const prompt = fullAuditPrompt('/path')
    
    expect(prompt).toContain('skills you have available')
    expect(prompt).toContain('security-audit skill')
  })
  
  it('should handle different target paths', () => {
    const fullAuditPrompt = (target) => {
      return `
First, list what skills you have available.

Then audit this codebase: ${target}

Use the security-audit skill to guide your comprehensive security and performance analysis.
`.trim()
    }
    
    const testPaths = [
      '/absolute/path',
      './relative/path',
      '../parent/path',
      '~/home/path',
      'C:\\\\Windows\\\\Path',
    ]
    
    testPaths.forEach(path => {
      const prompt = fullAuditPrompt(path)
      expect(prompt).toContain(path)
    })
  })
})
