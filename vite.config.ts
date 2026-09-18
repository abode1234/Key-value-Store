import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts'],
    // Force Vitest to ignore compiled output and node_modules
    exclude: ['dist/**', 'node_modules/**'], 
  },
})
