export default defineEventHandler(async (event) => {
  const username = getRouterParam(event, "user");
  const { size = 200 } = getQuery(event);

  if (!username) {
    return {
      statusCode: 400,
      message: "Username is required",
    };
  }

  const githubAvatarUrl = `https://github.com/${username}.png?size=${size}`;

  try {
    const response = await fetch(githubAvatarUrl);

    if (!response.ok) {
      return {
        statusCode: response.status,
        message: "Failed to fetch GitHub avatar",
      };
    }

    const imageBuffer = await response.arrayBuffer();

    event.node.res.setHeader("Content-Type", "image/png");
    event.node.res.setHeader("Cache-Control", "max-age=3600");

    return new Uint8Array(imageBuffer);
  } catch (error) {
    return {
      statusCode: 500,
      message: "Server error",
    };
  }
});
