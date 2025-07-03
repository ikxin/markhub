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

const LINK_REGEX =
  /((<link[^>]+rel=.(icon|shortcut icon|alternate icon|apple-touch-icon)[^>]+>))/i

const HREF_REGEX = /href=["']([^"']+)["']/i

const getIcoByLinkTag = async (host?: string) => {
  const html = await fetch(`http://${host}`).then((res) => res.text())

  const link = html.match(LINK_REGEX)
  if (!link) throw new Error()

  const href = link[1].match(HREF_REGEX)
  if (!href) throw new Error()

  let [_, iconUrl] = href
  iconUrl = new URL(iconUrl, `http://${host}`).toString()

  const response = await fetch(iconUrl)

  if (!response.ok) throw new Error()

  const buffer = Buffer.from(await response.arrayBuffer())

  if (iconUrl.endsWith('.ico')) {
    return icoToPng(buffer, 64)
  } else {
    return sharp(buffer).resize({ width: 64, height: 64 }).toBuffer()
  }
}

const getIcoByFavicon = async (host?: string) => {
  const buffer = Buffer.from(
    await fetch(`http://${host}/favicon.ico`).then((res) => res.arrayBuffer()),
  )

  return icoToPng(buffer, 64)
}
