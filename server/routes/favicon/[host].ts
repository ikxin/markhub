export default defineEventHandler(async (event) => {
  const host = getRouterParam(event, "host")!;

  event.node.res.setHeader("Content-Type", "image/png");

  try {
    return await getIcoByLinkTag(host);
  } catch (error) {}

  try {
    return await getIcoByFavicon(host);
  } catch (error) {}

  const defaultIco = await generateStringIco(host?.charAt(0).toUpperCase());

  return defaultIco;
});
