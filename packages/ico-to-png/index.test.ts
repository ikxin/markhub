import { beforeAll, describe, expect, it } from 'vitest'
import { dirname, join } from 'node:path'
import { mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import icoToPng from './src/index.ts'
import sharp from 'sharp'

const __dirname = dirname(fileURLToPath(import.meta.url))
const fixturesDir = join(__dirname, 'resource')
const outputDir = join(__dirname, 'output')
const files = readdirSync(fixturesDir).filter((f) => f.endsWith('.ico'))

describe('icoToPng', () => {
  beforeAll(() => {
    mkdirSync(outputDir, { recursive: true })
  })

  for (const file of files) {
    it(`converts ${file} to valid PNG (sharp-compatible)`, async () => {
      const ico = new Uint8Array(readFileSync(join(fixturesDir, file)))
      const png = icoToPng(ico)

      writeFileSync(join(outputDir, file.replace('.ico', '.png')), png)

      const metadata = await sharp(Buffer.from(png)).metadata()
      expect(metadata.format).toBe('png')
      expect(metadata.width).toBeGreaterThan(0)
      expect(metadata.width).toBeLessThanOrEqual(1024)
      expect(metadata.height).toBeGreaterThan(0)
      expect(metadata.height).toBeLessThanOrEqual(1024)
      expect(metadata.channels).toBeGreaterThanOrEqual(3)

      const resized = await sharp(Buffer.from(png)).resize(32, 32).toBuffer()
      expect(resized.length).toBeGreaterThan(0)
    })
  }

  it('accepts ArrayBuffer input', async () => {
    const ico = readFileSync(join(fixturesDir, 'github.ico'))
    const ab = ico.buffer.slice(ico.byteOffset, ico.byteOffset + ico.byteLength)
    const png = icoToPng(ab)
    const metadata = await sharp(Buffer.from(png)).metadata()
    expect(metadata.format).toBe('png')
  })

  it('respects size parameter', async () => {
    const ico = new Uint8Array(readFileSync(join(fixturesDir, 'github.ico')))
    const png16 = icoToPng(ico, 16)
    const png32 = icoToPng(ico, 32)
    const meta16 = await sharp(Buffer.from(png16)).metadata()
    const meta32 = await sharp(Buffer.from(png32)).metadata()
    expect(meta16.width).toBe(16)
    expect(meta32.width).toBe(32)
  })

  it('throws on invalid input', () => {
    expect(() => icoToPng(new Uint8Array([0, 0, 0]))).toThrow(/Invalid ICO/)
  })
})
