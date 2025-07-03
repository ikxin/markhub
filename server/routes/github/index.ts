export default defineResponseHandler(async (event) => {
  const { id } = getQuery(event)

  try {
    const fetchUrl = `https://avatars.githubusercontent.com/u/${id}?size=100`
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
