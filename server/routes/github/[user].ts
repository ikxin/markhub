export default defineResponseHandler(async (event) => {
  const user = getRouterParam(event, 'user')

  try {
    const fetchUrl = `https://github.com/${user}.png?size=100`
    const response = await fetch(fetchUrl)

    if (response.ok) {
      return Buffer.from(await response.arrayBuffer())
    } else {
      throw new Error()
    }
  } catch {
    const fallbackKey = 'fallback/github.png'
    return useStorage('assets:server').getItemRaw(fallbackKey)
  }
})
