import { defineConfig } from 'tsdown'

export default defineConfig({
  entry: 'src/index.ts',
  dts: {
    tsgo: true,
  },
  exports: true,
  format: ['cjs', 'esm', 'iife', 'umd'],
  platform: 'neutral',
  clean: true,
})