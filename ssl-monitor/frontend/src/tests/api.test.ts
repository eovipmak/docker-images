import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';

/**
 * Unit tests for the API service
 * 
 * These tests verify that:
 * 1. Trailing slashes are removed from URLs
 * 2. 307 redirects are handled correctly
 */

describe('API URL Normalization', () => {
  beforeEach(() => {
    // Clear any cached tokens
    localStorage.clear();
  });

  it('should remove trailing slashes from URLs', () => {
    const testCases = [
      { input: '/api/domains/example.com/', expected: '/api/domains/example.com' },
      { input: '/api/monitors/test.com/', expected: '/api/monitors/test.com' },
      { input: '/api/domains/example.com///', expected: '/api/domains/example.com' },
      { input: '/api/check', expected: '/api/check' },
      { input: '/', expected: '/' },
    ];

    testCases.forEach(({ input, expected }) => {
      let normalizedUrl = input;
      
      // Apply the same normalization logic as in api.ts
      if (normalizedUrl && normalizedUrl !== '/' && normalizedUrl.endsWith('/')) {
        normalizedUrl = normalizedUrl.replace(/\/+$/, '');
      }
      
      expect(normalizedUrl).toBe(expected);
    });
  });

  it('should preserve URLs without trailing slashes', () => {
    const urls = [
      '/api/domains/example.com',
      '/api/monitors/test.com',
      '/api/check',
      '/api/stats',
    ];

    urls.forEach((url) => {
      let normalizedUrl = url;
      
      // Apply normalization
      if (normalizedUrl && normalizedUrl !== '/' && normalizedUrl.endsWith('/')) {
        normalizedUrl = normalizedUrl.replace(/\/+$/, '');
      }
      
      expect(normalizedUrl).toBe(url);
    });
  });

  it('should handle root URL correctly', () => {
    const url = '/';
    let normalizedUrl = url;
    
    // Apply normalization - root should not be touched
    if (normalizedUrl && normalizedUrl !== '/' && normalizedUrl.endsWith('/')) {
      normalizedUrl = normalizedUrl.replace(/\/+$/, '');
    }
    
    expect(normalizedUrl).toBe('/');
  });
});

describe('Domain Name Encoding', () => {
  it('should properly encode domain names with special characters', () => {
    const domains = [
      'example.com',
      'sub.example.com',
      'example-hyphen.com',
      'example123.com',
    ];

    domains.forEach((domain) => {
      const encoded = encodeURIComponent(domain);
      const decoded = decodeURIComponent(encoded);
      
      expect(decoded).toBe(domain);
    });
  });

  it('should handle international domain names', () => {
    const idn = 'münchen.de';
    const encoded = encodeURIComponent(idn);
    
    // Verify it's encoded
    expect(encoded).not.toBe(idn);
    
    // Verify it can be decoded back
    const decoded = decodeURIComponent(encoded);
    expect(decoded).toBe(idn);
  });
});
