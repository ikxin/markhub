import * as z from 'zod/v4'
import crypto from 'crypto'

function getHash(email: string) {
  return crypto
    .createHash('md5')
    .update(email.trim().toLowerCase())
    .digest('hex')
}

const paramSchema = z.object({
  hash: z.union([
    z.string().regex(/^([a-f0-9]{32}|[a-f0-9]{64})$/i),
    z.email().transform((email) => getHash(email)),
  ]),
})

/**
 * Source https://docs.gravatar.com/sdk/images
 */
const querySchema = z.object({
  // Size
  s: z.string().optional().default('100'),
  size: z.string().optional(),
  // Default Image
  default: z.string().optional(),
  // Force Default
  f: z.string().optional(),
  forcedefault: z.string().optional(),
  // Rating
  r: z.string().optional().default('g'),
  rating: z.string().optional(),
  // Initials
  initials: z.string().optional(),
  name: z.string().optional(),
})

export default defineResponseHandler(async (event) => {
  try {
    const query = await getValidatedQuery(event, querySchema.safeParse)

    const params = await getValidatedRouterParams(event, paramSchema.safeParse)
    if (!params.success) throw new Error()

    const queryParams = new URLSearchParams({ ...query.data })

    const fetchUrl = `https://secure.gravatar.com/avatar/${params.data.hash}?${queryParams}`
    const response = await fetch(fetchUrl)

    if (response.ok) {
      return Buffer.from(await response.arrayBuffer())
    } else {
      throw new Error()
    }
  } catch {
    return useStorage('assets:server').getItemRaw('fallback/gravatar.png')
  }
})
