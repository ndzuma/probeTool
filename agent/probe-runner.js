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
const verbose = getArg('verbose') === 'true'

// Helper function for verbose logging
function verboseLog(msg) {
  if (verbose) {
    console.log(`VERBOSE:${msg}`)
  }
}

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

// Verbose logging
verboseLog(`Agent directory: ${__dirname}`)
verboseLog(`Target directory: ${target}`)
verboseLog(`Looking for skills in: ${path.join(__dirname, '.claude/skills')}`)
verboseLog(`Model: ${model}`)
verboseLog(`Allowed tools: ${options.allowedTools.join(', ')}`)

// Check if skill file exists
const skillPath = path.join(__dirname, '.claude/skills/security-audit/SKILL.md')
const { existsSync } = await import('fs')
if (existsSync(skillPath)) {
  verboseLog(`✓ security-audit skill found at: ${skillPath}`)
} else {
  verboseLog(`✗ security-audit skill NOT found at: ${skillPath}`)
}

// Run audit
let markdownOutput = ''
let currentSection = ''

console.log('PROGRESS:reading_files')

try {
  for await (const message of query({ prompt: fullAuditPrompt(target), options })) {
    
    // ADD: Log message types in verbose mode
    if (verbose && message.type) {
      verboseLog(`Message type: ${message.type}`)
    }
    
    if (message.type === 'assistant') {
      for (const block of message.message.content) {
        if (block.type === 'text') {
          markdownOutput += block.text
          
          // ADD: Show snippet of what model is saying
          if (verbose) {
            const snippet = block.text.substring(0, 100).replace(/\n/g, ' ')
            verboseLog(`Assistant: ${snippet}...`)
          }
          
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
        
        // ADD: Log tool use in verbose mode
        if (block.type === 'tool_use' && verbose) {
          verboseLog(`Tool call: ${block.name}`)
          if (block.name === 'Skill') {
            verboseLog(`  → Using skill: ${block.input?.name || 'unknown'}`)
          }
        }
      }
    } 
    
    // ADD: Log tool results
    else if (message.type === 'tool_result' && verbose) {
      verboseLog(`Tool result: ${message.tool_name}`)
    }
    
    else if (message.type === 'result') {
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
