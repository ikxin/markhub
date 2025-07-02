export default defineResponseHandler(async (event) => {
  const hash = getRouterParam(event, "hash");
  const { size = 100, s = 100 } = getQuery(event);

  try {
    const fetchUrl = `https://secure.gravatar.com/avatar/${hash}.png?s=${
      size || s || 100
    }`;
    const response = await fetch(fetchUrl);

    if (response.ok) {
      return Buffer.from(await response.arrayBuffer());
    } else {
      throw new Error();
    }
  } catch {
    const fallbackKey = "fallback/gravatar.png";
    return useStorage("assets:server").getItemRaw(fallbackKey);
  }
});
