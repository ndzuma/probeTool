// Mock the Claude SDK for testing
import { vi } from 'vitest'

export const query = vi.fn()

// Helper to mock successful responses
export function mockQuerySuccess(responses) {
  let callIndex = 0
  query.mockImplementation(() => {
    return {
      [Symbol.asyncIterator]: async function* () {
        for (const response of responses) {
          yield response
        }
      }
    }
  })
}

// Helper to mock error responses
export function mockQueryError(error) {
  query.mockImplementation(() => {
    throw error
  })
}

// Helper to create a mock message
export function createMockMessage(type, content) {
  return {
    type,
    ...(type === 'assistant' && { message: { content } }),
    ...(type === 'result' && { 
      subtype: 'success',
      total_cost_usd: 0.1234,
      duration_ms: 45000,
      num_turns: 10,
      ...content 
    }),
  }
}
