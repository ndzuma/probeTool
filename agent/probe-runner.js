import { query } from '@anthropic-ai/claude-agent-sdk'
import { writeFile } from 'fs/promises'
import { resolve } from 'path'

// Parse CLI args
const args = process.argv.slice(2)
const getArg = (name) => args.find(a => a.startsWith(`--${name}=`))?.split('=')[1]

const target = getArg('target') || process.cwd()
const outPath = getArg('out')
const model = getArg('model') || 'anthropic/claude-3.5-haiku'

// CRITICAL: OpenRouter requires these exact env vars
// From docs: https://openrouter.ai/docs/guides/community/anthropic-agent-sdk
if (!process.env.ANTHROPIC_AUTH_TOKEN) {
  console.error('ERROR: ANTHROPIC_AUTH_TOKEN not set (should be OpenRouter key)')
  process.exit(1)
}

if (!process.env.ANTHROPIC_BASE_URL) {
  console.error('ERROR: ANTHROPIC_BASE_URL not set (should be https://openrouter.ai/api)')
  process.exit(1)
}

if (process.env.ANTHROPIC_API_KEY !== '') {
  console.error('ERROR: ANTHROPIC_API_KEY must be explicitly empty string')
  process.exit(1)
}

// Import prompt
import { fullAuditPrompt } from './prompts.js'

// Progress streaming
console.log(`PROGRESS:init:openrouter:${model}`)

// Claude Agent SDK options (per TypeScript docs)
const options = {
  model: model,
  allowedTools: ['Read', 'Glob', 'Grep', 'Bash'],
  permissionMode: 'acceptEdits',
  
  // Use Claude Code preset + append custom instructions
  systemPrompt: {
    type: 'preset',
    preset: 'claude_code',
    append: `
You are a senior software engineer performing comprehensive code audits.
Use the code-audit skill to structure your analysis.
Focus on security, performance, architecture, and code quality.
    `.trim()
  },
  
  // Load .claude/skills/ from project
  settingSources: ['project'],
  
  cwd: target
}

// Run audit
let fullOutput = ''
let currentSection = ''

console.log('PROGRESS:reading_files')

try {
  for await (const message of query({ prompt: fullAuditPrompt(target), options })) {
    // Process different message types
    if (message.type === 'assistant') {
      for (const block of message.message.content) {
        if (block.type === 'text') {
          fullOutput += block.text
          
          // Smart progress detection
          const text = block.text.toLowerCase()
          if (text.includes('security') || text.includes('vulnerab')) {
            if (currentSection !== 'security') {
              console.log('PROGRESS:security')
              currentSection = 'security'
            }
          } else if (text.includes('performance') || text.includes('optim')) {
            if (currentSection !== 'performance') {
              console.log('PROGRESS:performance')
              currentSection = 'performance'
            }
          } else if (text.includes('architect') || text.includes('design')) {
            if (currentSection !== 'architecture') {
              console.log('PROGRESS:architecture')
              currentSection = 'architecture'
            }
          } else if (text.includes('quality') || text.includes('refactor')) {
            if (currentSection !== 'quality') {
              console.log('PROGRESS:quality')
              currentSection = 'quality'
            }
          }
        }
      }
    } else if (message.type === 'result') {
      // Final result
      if (message.subtype === 'success') {
        console.log('PROGRESS:finalizing')
        
        // Add metadata to output
        fullOutput += `\n\n---\n\n## Audit Metadata\n\n`
        fullOutput += `- **Cost**: $${message.total_cost_usd.toFixed(4)}\n`
        fullOutput += `- **Duration**: ${(message.duration_ms / 1000).toFixed(2)}s\n`
        fullOutput += `- **Turns**: ${message.num_turns}\n`
        fullOutput += `- **Model**: ${model}\n`
      } else {
        // Error cases
        console.error(`ERROR:Audit failed: ${message.subtype}`)
        if (message.errors) {
          message.errors.forEach(err => console.error(`ERROR:${err}`))
        }
        process.exit(1)
      }
    }
  }
  
  // Write markdown
  await writeFile(outPath, fullOutput, 'utf-8')
  console.log(`SUCCESS:${outPath}`)
  
} catch (error) {
  console.error(`ERROR:${error.message}`)
  process.exit(1)
}
