export default defineResponseHandler(async (event) => {
  const user = getRouterParam(event, 'user')

  try {
    const fetchUrl = `https://images.opencollective.com/${user}/avatar.png?width=100&height=100`
    const response = await fetch(fetchUrl)

    if (response.ok) {
      return Buffer.from(await response.arrayBuffer())
    } else {
      throw new Error()
    }
  } catch {
    const fallbackKey = 'fallback/opencollective.png'
    return useStorage('assets:server').getItemRaw(fallbackKey)
  }
})
