import { describe, expect, it } from 'vitest'
import { dirname, join } from 'node:path'
import { readFileSync, readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { icoToSharp } from '../../server/utils/ico-to-sharp'
import sharp from 'sharp'

const __dirname = dirname(fileURLToPath(import.meta.url))
const fixturesDir = join(__dirname, '../fixtures/ico-to-sharp')
const files = readdirSync(fixturesDir).filter((file) => file.endsWith('.ico'))

describe('icoToSharp', () => {
  describe('image parsing', () => {
    for (const file of files) {
      it(`converts ${file} to sharp-compatible format`, async () => {
        const input = new Uint8Array(readFileSync(join(fixturesDir, file)))
        const output = icoToSharp(input)

        const metadata = await sharp(Buffer.from(output)).metadata()
        expect(['png', 'jpeg', 'gif', 'webp']).toContain(metadata.format)
        expect(metadata.width).toBeGreaterThan(0)
        expect(metadata.width).toBeLessThanOrEqual(1024)
        expect(metadata.height).toBeGreaterThan(0)
        expect(metadata.height).toBeLessThanOrEqual(1024)
        expect(metadata.channels).toBeGreaterThanOrEqual(3)

        const resized = await sharp(Buffer.from(output))
          .resize(32, 32)
          .toBuffer()
        expect(resized.length).toBeGreaterThan(0)
      })
    }
  })

  describe('utility functions', () => {
    it('accepts ArrayBuffer input', async () => {
      const input = readFileSync(join(fixturesDir, 'github.ico'))
      const ab = input.buffer.slice(
        input.byteOffset,
        input.byteOffset + input.byteLength,
      )
      const output = icoToSharp(ab)
      const metadata = await sharp(Buffer.from(output)).metadata()

      expect(['png', 'jpeg', 'gif', 'webp']).toContain(metadata.format)
    })

    it('respects size parameter', async () => {
      const input = new Uint8Array(readFileSync(join(fixturesDir, 'github.ico')))
      const out16 = icoToSharp(input, 16)
      const out32 = icoToSharp(input, 32)
      const meta16 = await sharp(Buffer.from(out16)).metadata()
      const meta32 = await sharp(Buffer.from(out32)).metadata()

      expect(meta16.width).toBe(16)
      expect(meta32.width).toBe(32)
    })

    it('defaults to largest size when size not provided', async () => {
      const input = new Uint8Array(readFileSync(join(fixturesDir, 'github.ico')))
      const output = icoToSharp(input)
      const metadata = await sharp(Buffer.from(output)).metadata()

      expect(metadata.width).toBe(32)
    })

    it('throws on invalid input', () => {
      expect(() => icoToSharp(new Uint8Array([0, 0, 0]))).toThrow(
        /Invalid ICO/,
      )
    })
  })
})
