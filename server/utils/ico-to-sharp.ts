const PNG_SIG = /* @__PURE__ */ new Uint8Array([
  0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
])

const JPEG_SIG = /* @__PURE__ */ new Uint8Array([0xff, 0xd8, 0xff])

const GIF_SIG = /* @__PURE__ */ new Uint8Array([0x47, 0x49, 0x46])

const WEBP_SIG = /* @__PURE__ */ new Uint8Array([0x52, 0x49, 0x46, 0x46])

function isPng(b: Uint8Array): boolean {
  if (b.length < 8) return false
  for (let i = 0; i < 8; i++) if (b[i] !== PNG_SIG[i]) return false
  return true
}

function isJpeg(b: Uint8Array): boolean {
  if (b.length < 3) return false
  for (let i = 0; i < 3; i++) if (b[i] !== JPEG_SIG[i]) return false
  return true
}

function isGif(b: Uint8Array): boolean {
  if (b.length < 3) return false
  for (let i = 0; i < 3; i++) if (b[i] !== GIF_SIG[i]) return false
  return true
}

function isWebp(b: Uint8Array): boolean {
  if (b.length < 12) return false
  for (let i = 0; i < 4; i++) if (b[i] !== WEBP_SIG[i]) return false
  return b[8] === 0x57 && b[9] === 0x45 && b[10] === 0x42 && b[11] === 0x50
}

function isSharpSupported(b: Uint8Array): boolean {
  return isPng(b) || isJpeg(b) || isGif(b) || isWebp(b)
}

function r16(b: Uint8Array, o: number): number {
  return b[o] | (b[o + 1] << 8)
}

function r32(b: Uint8Array, o: number): number {
  return (b[o] | (b[o + 1] << 8) | (b[o + 2] << 16) | (b[o + 3] << 24)) >>> 0
}

function w32(b: Uint8Array, o: number, v: number): void {
  b[o] = (v >>> 24) & 0xff
  b[o + 1] = (v >>> 16) & 0xff
  b[o + 2] = (v >>> 8) & 0xff
  b[o + 3] = v & 0xff
}

/**
 * Convert image data to sharp-compatible format.
 *
 * Handles standard ICO files with BMP entries (1/4/8/24/32-bit),
 * ICO files with embedded PNG/JPEG entries, and PNG/JPEG/GIF/WebP files.
 *
 * @param input - Image file content as Uint8Array or ArrayBuffer
 * @param size  - Preferred icon size in pixels (defaults to largest available)
 * @returns Image data compatible with sharp
 */
export function icoToSharp(
  input: Uint8Array | ArrayBuffer,
  size?: number,
): Uint8Array {
  const buf = input instanceof Uint8Array ? input : new Uint8Array(input)

  // Pass through sharp-supported formats directly
  if (isSharpSupported(buf)) return buf

  // Validate ICO header: reserved=0, type=1 (icon)
  if (buf.length < 6 || r16(buf, 0) !== 0 || r16(buf, 2) !== 1)
    throw new Error('Invalid ICO file')

  const count = r16(buf, 4)
  if (count === 0) throw new Error('ICO contains no images')

  // Select best matching entry (defaults to largest)
  let bestIdx = 0
  let bestScore = -Infinity
  for (let i = 0; i < count; i++) {
    const e = 6 + i * 16
    const w = buf[e] || 256 // 0 means 256
    const score = size != null ? -Math.abs(w - size) * 0x10000 + w : w
    if (score > bestScore) {
      bestScore = score
      bestIdx = i
    }
  }

  const e = 6 + bestIdx * 16
  const offset = r32(buf, e + 12)
  const imgSize = r32(buf, e + 8)
  const img = buf.subarray(offset, offset + imgSize)

  // Entry may contain embedded sharp-supported image
  if (isSharpSupported(img)) return img

  // Decode BMP DIB data and encode as PNG
  return dibToPng(img)
}

// --- BMP DIB to PNG conversion ---

function dibToPng(d: Uint8Array): Uint8Array {
  const hdrSz = r32(d, 0)
  const w = r32(d, 4)
  // biHeight is signed int32, doubled in ICO (XOR mask + AND mask)
  const h = Math.abs(d[8] | (d[9] << 8) | (d[10] << 16) | (d[11] << 24)) >> 1
  const bits = r16(d, 14)
  const clrUsed = r32(d, 32)

  const nColors = clrUsed || (bits <= 8 ? 1 << bits : 0)
  const palOff = hdrSz
  const pixOff = palOff + nColors * 4
  const xorStr = ((w * bits + 31) >>> 5) << 2 // row stride padded to 4 bytes
  const andStr = ((w + 31) >>> 5) << 2
  const andOff = pixOff + xorStr * h

  const rgba = new Uint8Array(w * h * 4)

  for (let y = 0; y < h; y++) {
    const sy = h - 1 - y // BMP is bottom-up
    const row = pixOff + sy * xorStr
    for (let x = 0; x < w; x++) {
      const di = (y * w + x) << 2

      if (bits === 32) {
        const si = row + (x << 2)
        rgba[di] = d[si + 2] // R (BMP stores BGRA)
        rgba[di + 1] = d[si + 1] // G
        rgba[di + 2] = d[si] // B
        rgba[di + 3] = d[si + 3] // A
      } else if (bits === 24) {
        const si = row + x * 3
        rgba[di] = d[si + 2]
        rgba[di + 1] = d[si + 1]
        rgba[di + 2] = d[si]
        rgba[di + 3] = 0xff
      } else {
        // Palette-based: 1, 4, or 8 bits per pixel
        let ci: number
        if (bits === 8) {
          ci = d[row + x]
        } else if (bits === 4) {
          const bv = d[row + (x >>> 1)]
          ci = x & 1 ? bv & 0x0f : bv >>> 4
        } else {
          // 1-bit
          ci = (d[row + (x >>> 3)] >>> (7 - (x & 7))) & 1
        }
        const pi = palOff + (ci << 2)
        rgba[di] = d[pi + 2] // R (palette stores BGRX)
        rgba[di + 1] = d[pi + 1] // G
        rgba[di + 2] = d[pi] // B
        rgba[di + 3] = 0xff
      }
    }
  }

  // Handle transparency
  if (bits === 32) {
    // Check if alpha channel is actually used
    let hasAlpha = false
    for (let i = 3; i < rgba.length; i += 4) {
      if (rgba[i]) {
        hasAlpha = true
        break
      }
    }
    // If all alpha values are 0, use AND mask instead
    if (!hasAlpha) {
      for (let i = 3; i < rgba.length; i += 4) rgba[i] = 0xff
      applyAndMask(rgba, d, andOff, w, h, andStr)
    }
  } else {
    applyAndMask(rgba, d, andOff, w, h, andStr)
  }

  return buildPng(w, h, rgba)
}

function applyAndMask(
  rgba: Uint8Array,
  d: Uint8Array,
  off: number,
  w: number,
  h: number,
  stride: number,
): void {
  for (let y = 0; y < h; y++) {
    const row = off + (h - 1 - y) * stride // AND mask is also bottom-up
    for (let x = 0; x < w; x++) {
      const bi = row + (x >>> 3)
      if (bi >= d.length) return
      if ((d[bi] >>> (7 - (x & 7))) & 1) {
        rgba[(y * w + x) * 4 + 3] = 0 // transparent
      }
    }
  }
}

// --- Minimal PNG encoder ---

const CRC_TBL = /* @__PURE__ */ (() => {
  const t = new Uint32Array(256)
  for (let n = 0; n < 256; n++) {
    let c = n
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
    t[n] = c >>> 0
  }
  return t
})()

function crc32(b: Uint8Array, off: number, len: number): number {
  let c = ~0
  for (let i = off; i < off + len; i++)
    c = CRC_TBL[(c ^ b[i]) & 0xff] ^ (c >>> 8)
  return ~c >>> 0
}

function adler32(b: Uint8Array): number {
  let a = 1
  let s = 0
  for (let i = 0; i < b.length; i++) {
    a = (a + b[i]) % 65521
    s = (s + a) % 65521
  }
  return ((s << 16) | a) >>> 0
}

function zlibStore(raw: Uint8Array): Uint8Array {
  const len = raw.length
  const max = 0xffff
  const n = Math.max(1, Math.ceil(len / max))
  const out = new Uint8Array(2 + n * 5 + len + 4)
  let p = 0

  // Zlib header (RFC 1950)
  out[p++] = 0x78
  out[p++] = 0x01

  // Deflate stored blocks (RFC 1951)
  for (let i = 0, off = 0; i < n; i++) {
    const sz = Math.min(len - off, max)
    out[p++] = i === n - 1 ? 1 : 0 // BFINAL
    out[p++] = sz & 0xff
    out[p++] = (sz >>> 8) & 0xff
    const nsz = sz ^ 0xffff
    out[p++] = nsz & 0xff
    out[p++] = (nsz >>> 8) & 0xff
    out.set(raw.subarray(off, off + sz), p)
    p += sz
    off += sz
  }

  // Adler-32 checksum (big-endian)
  w32(out, p, adler32(raw))
  return out
}

function pngChunk(
  buf: Uint8Array,
  p: number,
  type: number,
  data: Uint8Array | null,
  len: number,
): number {
  w32(buf, p, len)
  w32(buf, p + 4, type)
  if (data && len > 0) buf.set(data.subarray(0, len), p + 8)
  w32(buf, p + 8 + len, crc32(buf, p + 4, 4 + len))
  return p + 12 + len
}

function buildPng(w: number, h: number, rgba: Uint8Array): Uint8Array {
  // Build filtered scanlines (filter byte 0 = None per row)
  const rowLen = 1 + w * 4
  const raw = new Uint8Array(h * rowLen)
  for (let y = 0; y < h; y++) {
    raw[y * rowLen] = 0 // filter: None
    raw.set(rgba.subarray(y * w * 4, (y + 1) * w * 4), y * rowLen + 1)
  }

  const zlib = zlibStore(raw)

  // IHDR data: width(4) + height(4) + depth(1) + colorType(1) + comp(1) + filter(1) + interlace(1)
  const ihdr = new Uint8Array(13)
  w32(ihdr, 0, w)
  w32(ihdr, 4, h)
  ihdr[8] = 8 // bit depth
  ihdr[9] = 6 // color type: RGBA

  // Assemble PNG: signature(8) + IHDR(25) + IDAT(12+zlib) + IEND(12)
  const png = new Uint8Array(8 + 25 + 12 + zlib.length + 12)
  png.set(PNG_SIG)
  let p = 8
  p = pngChunk(png, p, 0x49484452, ihdr, 13) // IHDR
  p = pngChunk(png, p, 0x49444154, zlib, zlib.length) // IDAT
  p = pngChunk(png, p, 0x49454e44, null, 0) // IEND
  return png
}
