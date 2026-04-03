import { before, describe, it } from 'node:test'
import { dirname, join } from 'node:path'
import { mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import assert from 'node:assert/strict'
import { fileURLToPath } from 'node:url'
import icoToPng from './src/index.ts'
import sharp from 'sharp'

const __dirname = dirname(fileURLToPath(import.meta.url))
const fixturesDir = join(__dirname, 'test')
const outputDir = join(__dirname, 'output')
const files = readdirSync(fixturesDir).filter((f) => f.endsWith('.ico'))

describe('icoToPng', () => {
  before(() => {
    mkdirSync(outputDir, { recursive: true })
  })

  for (const file of files) {
    it(`converts ${file} to valid PNG (sharp-compatible)`, async () => {
      const ico = new Uint8Array(readFileSync(join(fixturesDir, file)))
      const png = icoToPng(ico)

      // Write output for manual inspection
      writeFileSync(join(outputDir, file.replace('.ico', '.png')), png)

      // Primary check: sharp must be able to parse and process the output
      const metadata = await sharp(Buffer.from(png)).metadata()
      assert.strictEqual(
        metadata.format,
        'png',
        'sharp should recognise output as PNG',
      )
      assert.ok(
        metadata.width != null && metadata.width > 0 && metadata.width <= 1024,
        `sharp width ${metadata.width} should be 1-1024`,
      )
      assert.ok(
        metadata.height != null &&
          metadata.height > 0 &&
          metadata.height <= 1024,
        `sharp height ${metadata.height} should be 1-1024`,
      )
      assert.ok(
        metadata.channels != null && metadata.channels >= 3,
        `sharp channels ${metadata.channels} should be at least 3`,
      )

      // Sanity check: sharp can produce valid output from our PNG
      const resized = await sharp(Buffer.from(png)).resize(32, 32).toBuffer()
      assert.ok(
        resized.length > 0,
        'sharp should produce non-empty resized output',
      )
    })
  }

  it('accepts ArrayBuffer input', async () => {
    const ico = readFileSync(join(fixturesDir, 'github.ico'))
    const ab = ico.buffer.slice(ico.byteOffset, ico.byteOffset + ico.byteLength)
    const png = icoToPng(ab)
    const metadata = await sharp(Buffer.from(png)).metadata()
    assert.strictEqual(metadata.format, 'png')
  })

  it('respects size parameter', async () => {
    const ico = new Uint8Array(readFileSync(join(fixturesDir, 'github.ico')))
    // github.ico has 16x16 and 32x32 entries
    const png16 = icoToPng(ico, 16)
    const png32 = icoToPng(ico, 32)
    const meta16 = await sharp(Buffer.from(png16)).metadata()
    const meta32 = await sharp(Buffer.from(png32)).metadata()
    assert.strictEqual(meta16.width, 16)
    assert.strictEqual(meta32.width, 32)
  })

  it('throws on invalid input', () => {
    assert.throws(() => icoToPng(new Uint8Array([0, 0, 0])), /Invalid ICO/)
  })
})
