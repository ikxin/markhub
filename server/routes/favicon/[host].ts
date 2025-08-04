import * as z from 'zod/v4'
import icoToPng from 'ico-to-png'
import sharp from 'sharp'

const paramSchema = z.object({
  host: z.union([
    z.ipv4(),
    z.ipv6(),
    z.string().regex(z.regexes.hostname),
    z.string().regex(z.regexes.domain),
  ]),
})

export default defineResponseHandler(async (event) => {
  const params = await getValidatedRouterParams(event, paramSchema.safeParse)
  if (!params.success)
    return useStorage('assets:server').getItemRaw('fallback/favicon.png')

  try {
    return await getIcoByLinkTag(params.data.host)
  } catch {
    /* next */
  }

  try {
    return await getIcoByFavicon(params.data.host)
  } catch {
    return useStorage('assets:server').getItemRaw('fallback/favicon.png')
  }
})

const defaultSize = { width: 100, height: 100 }

const getIcoByLinkTag = async (host?: string) => {
  const source = await fetch(`http://${host}`).then((res) => res.text())
  const linkMatch = source?.match(
    /((<link[^>]+rel=.(icon|shortcut icon|alternate icon|apple-touch-icon)[^>]+>))/g,
  )
  const hrefMatch = linkMatch?.toString().match(/href=["']([^"']+)["']/i)

  if (!hrefMatch) throw new Error()

  const fetchUrl = new URL(hrefMatch[1], `http://${host}`)
  fetchUrl.search = ''
  const response = await fetch(fetchUrl.toString())

  if (response.ok) {
    const buffer = Buffer.from(await response.arrayBuffer())
    return fetchUrl.toString().endsWith('.ico')
      ? sharp(await icoToPng(buffer, 64)).resize(defaultSize)
      : sharp(buffer).resize(defaultSize)
  } else {
    throw new Error()
  }
}

const getIcoByFavicon = async (host?: string) => {
  const fetchUrl = new URL('/favicon.ico', `http://${host}`)
  const response = await fetch(fetchUrl.toString())

  if (response.ok) {
    const buffer = Buffer.from(await response.arrayBuffer())
    return sharp(await icoToPng(buffer, 64)).resize(defaultSize)
  } else {
    throw new Error()
  }
}
