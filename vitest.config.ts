import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    projects: [
      {
        test: {
          environment: 'node',
          include: ['test/unit/**/*.{test,spec}.ts'],
          name: 'unit',
        },
      },
    ],
  },
})
