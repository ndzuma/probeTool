import { query } from '@anthropic-ai/claude-agent-sdk'
import { writeFile } from 'fs/promises'
import { resolve } from 'path'
import { fileURLToPath } from 'url'
import { dirname } from 'path'
import path from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// Parse CLI args
const args = process.argv.slice(2)
const getArg = (name) => args.find(a => a.startsWith(`--${name}=`))?.split('=')[1]

const target = getArg('target') || process.cwd()
const outPath = getArg('out')
const model = getArg('model') || 'anthropic/claude-3.5-haiku'

// Verify required env vars for OpenRouter
if (!process.env.ANTHROPIC_AUTH_TOKEN) {
  console.error('ERROR: ANTHROPIC_AUTH_TOKEN not set')
  process.exit(1)
}

if (!process.env.ANTHROPIC_BASE_URL) {
  console.error('ERROR: ANTHROPIC_BASE_URL not set')
  process.exit(1)
}

if (process.env.ANTHROPIC_API_KEY !== '') {
  console.error('ERROR: ANTHROPIC_API_KEY must be explicitly empty string')
  process.exit(1)
}

// Import prompt
import { fullAuditPrompt } from './prompts.js'

console.log(`PROGRESS:init:openrouter:${model}`)

// FIX: Correct skill loading per SDK docs
const options = {
  model: model,
  
  // Add 'Skill' to allowed tools
  allowedTools: ['Skill', 'Read', 'Glob', 'Grep', 'Bash'],
  
  permissionMode: 'acceptEdits',
  
  systemPrompt: {
    type: 'preset',
    preset: 'claude_code',
    append: `
You are a senior security engineer conducting a comprehensive security audit.
Use the security-audit skill to structure your analysis.
    `.trim()
  },
  
  // Load skills from 'project' (looks for .claude/skills/ relative to cwd)
  settingSources: ['project'],
  
  // Set cwd to agent directory (where .claude/skills/ is located)
  cwd: __dirname,  // This is ~/.probe/agent/ which contains .claude/
  
  // But tools should operate on target repo
  workingDirectory: target  // Tools like Read, Grep work here
}

// Run audit
let markdownOutput = ''
let currentSection = ''

console.log('PROGRESS:reading_files')

try {
  for await (const message of query({ prompt: fullAuditPrompt(target), options })) {
    if (message.type === 'assistant') {
      for (const block of message.message.content) {
        if (block.type === 'text') {
          markdownOutput += block.text
          
          const text = block.text.toLowerCase()
          if (text.includes('critical vulnerabilities') || text.includes('🔴')) {
            if (currentSection !== 'critical') {
              console.log('PROGRESS:critical')
              currentSection = 'critical'
            }
          } else if (text.includes('high severity') || text.includes('🟠')) {
            if (currentSection !== 'high') {
              console.log('PROGRESS:high')
              currentSection = 'high'
            }
          } else if (text.includes('medium severity') || text.includes('🟡')) {
            if (currentSection !== 'medium') {
              console.log('PROGRESS:medium')
              currentSection = 'medium'
            }
          } else if (text.includes('files analyzed')) {
            if (currentSection !== 'finalizing') {
              console.log('PROGRESS:finalizing')
              currentSection = 'finalizing'
            }
          }
        }
      }
    } else if (message.type === 'result') {
      if (message.subtype === 'success') {
        console.log('PROGRESS:finalizing')
        
        // Clean up markdown
        const reportStart = markdownOutput.indexOf('# Security')
        if (reportStart !== -1) {
          markdownOutput = markdownOutput.substring(reportStart)
        }
        
        // Add metadata footer
        markdownOutput += `\n\n---\n\n## Audit Metadata\n\n`
        markdownOutput += `- **Cost**: $${message.total_cost_usd.toFixed(4)}\n`
        markdownOutput += `- **Duration**: ${(message.duration_ms / 1000).toFixed(2)}s\n`
        markdownOutput += `- **Turns**: ${message.num_turns}\n`
        markdownOutput += `- **Model**: ${model}\n`
        markdownOutput += `- **Auditor**: Claude Security Engineer\n`
      } else {
        console.error(`ERROR:Audit failed: ${message.subtype}`)
        if (message.errors) {
          message.errors.forEach(err => console.error(`ERROR:${err}`))
        }
        process.exit(1)
      }
    }
  }
  
  await writeFile(outPath, markdownOutput, 'utf-8')
  console.log(`SUCCESS:${outPath}`)
  
} catch (error) {
  console.error(`ERROR:${error.message}`)
  process.exit(1)
}
