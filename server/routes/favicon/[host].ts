export default defineEventHandler(async (event) => {
  const host = getRouterParam(event, "host")!;

  const defaultIco = await generateStringIco(host?.charAt(0).toUpperCase());

  event.node.res.setHeader("Content-Type", "image/png");

  try {
    return await getIcoByLinkTag(host);
  } catch (error) {}

  try {
    return await getIcoByFavicon(host);
  } catch (error) {}

  try {
    return await getIcoByWebmaster(host);
  } catch (error) {}

  try {
    return await getIcoByGoogle(host);
  } catch (error) {}

  return defaultIco;
});
