export default defineEventHandler(async (event) => {
  const host = getRouterParam(event, "host")!;

  let buffer: Buffer = await generateStringIco(host?.charAt(0).toUpperCase());

  try {
    buffer = await getFaviconIco(host);
  } catch (error) {
    buffer = await getLinkTagIco(host);
  }

  event.node.res.setHeader("Content-Type", "image/png");
  return buffer;
});
