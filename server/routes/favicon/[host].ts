import sharp from 'sharp'
import icoToPng from 'ico-to-png'

export default defineResponseHandler(async (event) => {
  const host = getRouterParam(event, 'host')

  try {
    return await getIcoByLinkTag(host)
  } catch (error) {}

  try {
    return await getIcoByFavicon(host)
  } catch (error) {}

  return useStorage('assets:server').getItemRaw('fallback/favicon.png')
})

const defaultSize = { width: 100, height: 100 }

const getIcoByLinkTag = async (host?: string) => {
  const source = await fetch(`http://${host}`).then((res) => res.text())
  const linkMatch = source?.match(
    /((<link[^>]+rel=.(icon|shortcut icon|alternate icon|apple-touch-icon)[^>]+>))/g,
  )
  const hrefMatch = linkMatch?.toString().match(/href=["']([^"']+)["']/i)

  if (!hrefMatch) throw new Error()

  const fetchUrl = new URL(hrefMatch[1], `http://${host}`).toString()
  const response = await fetch(fetchUrl)

  if (response.ok) {
    const buffer = Buffer.from(await response.arrayBuffer())
    return fetchUrl.endsWith('.ico')
      ? sharp(await icoToPng(buffer, 64)).resize(defaultSize)
      : sharp(buffer).resize(defaultSize)
  } else {
    throw new Error()
  }
}

const getIcoByFavicon = async (host?: string) => {
  const fetchUrl = new URL('/favicon.ico', `http://${host}`).toString()
  const response = await fetch(fetchUrl)

  if (response.ok) {
    const buffer = Buffer.from(await response.arrayBuffer())
    return sharp(await icoToPng(buffer, 64)).resize(defaultSize)
  } else {
    throw new Error()
  }
}
