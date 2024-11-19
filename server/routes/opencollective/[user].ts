export default defineEventHandler(async (event) => {
  // 获取路由参数和查询参数
  const username = getRouterParam(event, "user");
  const { size = 128 } = getQuery(event);

  // 验证用户名是否存在
  if (!username) {
    throw createError({
      statusCode: 400,
      message: "Username is required",
    });
  }

  // 设置通用响应头
  const setResponseHeaders = () => {
    event.node.res.setHeader("Content-Type", "image/png");
    event.node.res.setHeader("Cache-Control", "max-age=3600, public");
  };

  // 获取头像图片
  const getAvatar = async (url: string) => {
    const response = await fetch(url);
    if (!response.ok) {
      throw createError({
        statusCode: response.status,
        message: `Failed to fetch avatar from ${url}`,
      });
    }
    return new Uint8Array(await response.arrayBuffer());
  };

  try {
    // 尝试获取用户头像
    const avatarUrl = `https://images.opencollective.com/${username}/avatar.png?height=${size}&width=${size}`;
    try {
      const avatar = await getAvatar(avatarUrl);
      setResponseHeaders();
      return avatar;
    } catch {
      // 如果获取用户头像失败,使用默认头像
      const defaultAvatarUrl = `https://images.opencollective.com/opencollective/avatar.png?height=${size}&width=${size}`;
      const defaultAvatar = await getAvatar(defaultAvatarUrl);
      setResponseHeaders();
      return defaultAvatar;
    }
  } catch (error) {
    throw createError({
      statusCode: 500,
      message: "Failed to fetch avatar",
    });
  }
});
