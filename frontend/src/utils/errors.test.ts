import { describe, expect, it } from 'vitest'
import { getErrorMessage, isNetworkError } from './errors'

describe('getErrorMessage', () => {
  it('extracts backend messages', () => {
    expect(getErrorMessage({ code: 40100, message: 'invalid credentials', status: 401, isNetwork: false })).toBe(
      'invalid credentials',
    )
  })

  it('falls back when the message is missing or empty', () => {
    expect(getErrorMessage({})).toBe('The request failed. Please try again.')
    expect(getErrorMessage({ message: '' })).toBe('The request failed. Please try again.')
    expect(getErrorMessage('boom')).toBe('The request failed. Please try again.')
    expect(getErrorMessage(null)).toBe('The request failed. Please try again.')
    expect(getErrorMessage(undefined, 'custom')).toBe('custom')
  })

  it('accepts a custom fallback', () => {
    expect(getErrorMessage({ message: 42 }, 'fallback')).toBe('fallback')
  })
})

describe('isNetworkError', () => {
  it('detects network-level failures', () => {
    expect(isNetworkError({ status: 0, code: 0, message: 'x', isNetwork: true })).toBe(true)
    expect(isNetworkError({ status: 401, code: 40100, message: 'x', isNetwork: false })).toBe(false)
    expect(isNetworkError(null)).toBe(false)
  })
})
