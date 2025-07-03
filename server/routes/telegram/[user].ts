export default defineResponseHandler(async (event) => {
  const user = getRouterParam(event, 'user')

  try {
    const fetchUrl = `https://t.me/i/userpic/320/${user}.jpg`
    const response = await fetch(fetchUrl)

    if (response.ok) {
      return Buffer.from(await response.arrayBuffer())
    } else {
      throw new Error()
    }
  } catch {
    const fallbackKey = 'fallback/telegram.png'
    return useStorage('assets:server').getItemRaw(fallbackKey)
  }
})
