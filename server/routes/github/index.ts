import sharp from "sharp";

export default defineEventHandler(async (event) => {
  const { id } = getQuery(event);

  const githubAvatarUrl = `https://avatars.githubusercontent.com/u/${id}?size=100`;

  const response = await fetch(githubAvatarUrl);

  event.node.res.setHeader("Content-Type", "image/png");

  return sharp(Buffer.from(await response.arrayBuffer()))
    .png()
    .toBuffer();
});
