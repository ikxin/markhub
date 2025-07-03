export default defineResponseHandler(async (event) => {
  const number = getRouterParam(event, 'number')
  const { s, spec } = getQuery(event)

  try {
    const fetchUrl = `https://q1.qlogo.cn/g?b=qq&nk=${number}&s=${
      s || spec || 100
    }`
    const response = await fetch(fetchUrl)

    if (response.ok) {
      return Buffer.from(await response.arrayBuffer())
    } else {
      throw new Error()
    }
  } catch {
    const fallbackKey = 'fallback/qq.png'
    return useStorage('assets:server').getItemRaw(fallbackKey)
  }
})
