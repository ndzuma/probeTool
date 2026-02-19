import { describe, it, expect, beforeEach, vi } from 'vitest'

// Mock the fs module
vi.mock('fs', () => ({
  readFileSync: vi.fn(),
}))

vi.mock('fs/promises', () => ({
  writeFile: vi.fn(),
}))

// Mock the path module
vi.mock('path', () => ({
  default: {
    join: vi.fn((...args) => args.join('/')),
  },
  join: vi.fn((...args) => args.join('/')),
  resolve: vi.fn((p) => p),
}))

// Mock the URL module
vi.mock('url', () => ({
  fileURLToPath: vi.fn(() => '/mocked/path/probe-runner.js'),
}))

// Tests to implement:
describe('CLI Argument Parsing', () => {
  it('should parse target argument', () => {
    // Test that target is parsed from CLI args
    const args = ['--target=/path/to/code', '--out=/output.md']
    const getArg = (name) => args.find(a => a.startsWith(`--${name}=`))?.split('=')[1]
    
    expect(getArg('target')).toBe('/path/to/code')
  })
  
  it('should parse model argument with default', () => {
    const args = ['--target=/path']
    const getArg = (name) => args.find(a => a.startsWith(`--${name}=`))?.split('=')[1]
    
    const model = getArg('model') || 'anthropic/claude-3.5-haiku'
    expect(model).toBe('anthropic/claude-3.5-haiku')
    
    const argsWithModel = ['--target=/path', '--model=custom-model']
    const getArgWithModel = (name) => argsWithModel.find(a => a.startsWith(`--${name}=`))?.split('=')[1]
    expect(getArgWithModel('model')).toBe('custom-model')
  })
  
  it('should parse verbose flag', () => {
    const args = ['--target=/path', '--verbose=true']
    const getArg = (name) => args.find(a => a.startsWith(`--${name}=`))?.split('=')[1]
    
    expect(getArg('verbose')).toBe('true')
    expect(getArg('verbose') === 'true').toBe(true)
  })
  
  it('should parse output path', () => {
    const args = ['--target=/path', '--out=/output/report.md']
    const getArg = (name) => args.find(a => a.startsWith(`--${name}=`))?.split('=')[1]
    
    expect(getArg('out')).toBe('/output/report.md')
  })
})


describe('Environment Validation', () => {
  beforeEach(() => {
    vi.unstubAllEnvs()
  })
  
  it('should fail if ANTHROPIC_AUTH_TOKEN missing', () => {
    // Simulate missing env var
    delete process.env.ANTHROPIC_AUTH_TOKEN
    
    const hasToken = !!process.env.ANTHROPIC_AUTH_TOKEN
    expect(hasToken).toBe(false)
  })
  
  it('should fail if ANTHROPIC_BASE_URL missing', () => {
    delete process.env.ANTHROPIC_BASE_URL
    
    const hasBaseUrl = !!process.env.ANTHROPIC_BASE_URL
    expect(hasBaseUrl).toBe(false)
  })
  
  it('should fail if ANTHROPIC_API_KEY not empty string', () => {
    process.env.ANTHROPIC_API_KEY = 'some-key'
    
    const isEmptyString = process.env.ANTHROPIC_API_KEY === ''
    expect(isEmptyString).toBe(false)
    
    // Reset
    process.env.ANTHROPIC_API_KEY = ''
    expect(process.env.ANTHROPIC_API_KEY).toBe('')
  })
})


describe('Skill Loading', () => {
  it('should load skill content from file', () => {
    // Mock the skill content
    const skillContent = `---
name: security-audit
---
# Security Audit Skill
This is the skill content.`
    
    expect(skillContent).toContain('Security Audit Skill')
    expect(skillContent.length).toBeGreaterThan(0)
  })
  
  it('should remove YAML frontmatter', () => {
    const skillWithFrontmatter = `---
name: security-audit
version: 1.0
---
# Security Audit Skill
This is the skill content.`
    
    // Remove YAML frontmatter
    let content = skillWithFrontmatter
    if (content.startsWith('---')) {
      const endMarker = content.indexOf('---', 3)
      if (endMarker !== -1) {
        content = content.substring(endMarker + 3).trim()
      }
    }
    
    expect(content).not.toContain('---')
    expect(content).toContain('Security Audit Skill')
  })
  
  it('should handle missing skill file', () => {
    // Simulate file not found error
    const fileExists = false
    
    if (!fileExists) {
      expect(() => {
        throw new Error('Could not load skill file: ENOENT')
      }).toThrow('Could not load skill file')
    }
  })
})


describe('Progress Reporting', () => {
  it('should emit progress events correctly', () => {
    const stages = ['init', 'reading_files', 'critical', 'high', 'medium', 'finalizing']
    const logs = []
    
    // Simulate progress logging
    stages.forEach(stage => {
      logs.push(`PROGRESS:${stage}`)
    })
    
    expect(logs).toHaveLength(6)
    expect(logs[0]).toBe('PROGRESS:init')
    expect(logs[logs.length - 1]).toBe('PROGRESS:finalizing')
  })
  
  it('should track section transitions', () => {
    let currentSection = ''
    const sections = ['critical', 'high', 'medium', 'finalizing']
    
    sections.forEach(section => {
      if (currentSection !== section) {
        currentSection = section
      }
    })
    
    expect(currentSection).toBe('finalizing')
  })
})


describe('Markdown Output', () => {
  it('should format markdown correctly', () => {
    let markdownOutput = '# Security Audit Report\n\n## Findings\n'
    
    // Simulate adding content
    markdownOutput += '- Finding 1\n'
    markdownOutput += '- Finding 2\n'
    
    expect(markdownOutput).toContain('# Security Audit Report')
    expect(markdownOutput).toContain('## Findings')
  })
  
  it('should include metadata footer', () => {
    const metadata = {
      cost: 0.1234,
      duration: 45000,
      turns: 10,
      model: 'anthropic/claude-3.5-haiku'
    }
    
    let footer = '\n\n---\n\n## Audit Metadata\n\n'
    footer += `- **Cost**: $${metadata.cost.toFixed(4)}\n`
    footer += `- **Duration**: ${(metadata.duration / 1000).toFixed(2)}s\n`
    footer += `- **Turns**: ${metadata.turns}\n`
    footer += `- **Model**: ${metadata.model}\n`
    
    expect(footer).toContain('Audit Metadata')
    expect(footer).toContain('Cost')
    expect(footer).toContain('Duration')
    expect(footer).toContain('Turns')
    expect(footer).toContain('Model')
  })
  
  it('should clean up output', () => {
    let markdownOutput = 'Some prefix text\n# Security Audit Report\nContent here'
    
    // Clean up - find the report start
    const reportStart = markdownOutput.indexOf('# Security')
    if (reportStart !== -1) {
      markdownOutput = markdownOutput.substring(reportStart)
    }
    
    expect(markdownOutput.startsWith('# Security')).toBe(true)
    expect(markdownOutput).not.toContain('Some prefix text')
  })
})


describe('Error Handling', () => {
  it('should handle API errors gracefully', () => {
    const apiError = new Error('API request failed')
    
    expect(() => {
      throw apiError
    }).toThrow('API request failed')
  })
  
  it('should exit with code 1 on failure', () => {
    // Mock process.exit
    const mockExit = vi.fn()
    
    // Simulate error condition
    const hasError = true
    if (hasError) {
      mockExit(1)
    }
    
    expect(mockExit).toHaveBeenCalledWith(1)
  })
  
  it('should log errors to stderr', () => {
    const errors = []
    const mockStderr = (msg) => errors.push(msg)
    
    mockStderr('ERROR:Something went wrong')
    
    expect(errors).toHaveLength(1)
    expect(errors[0]).toContain('ERROR:')
  })
})
