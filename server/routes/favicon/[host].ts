export default defineEventHandler(async (event) => {
  // 获取路由参数
  const host = getRouterParam(event, "host");

  // 验证主机名是否存在
  if (!host) {
    throw createError({
      statusCode: 400,
      message: "Host is required",
    });
  }

  // 设置通用响应头
  const setResponseHeaders = () => {
    event.node.res.setHeader("Content-Type", "image/png");
    event.node.res.setHeader("Cache-Control", "max-age=3600, public");
  };

  // 获取图标
  const getFavicon = async (url: string) => {
    const response = await fetch(url);
    if (!response.ok) {
      throw createError({
        statusCode: response.status,
        message: `Failed to fetch favicon from ${url}`,
      });
    }
    return response.arrayBuffer().then(buffer => new Uint8Array(buffer));
  };

  try {
    // 尝试直接获取 favicon.ico
    try {
      const favicon = await getFavicon(`http://${host}/favicon.ico`);
      setResponseHeaders();
      return favicon;
    } catch {
      // 如果获取失败,尝试从 HTML 中解析
      const html = await fetch(`http://${host}`).then(res => res.text());
      const faviconMatch = html.match(/<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+)["'][^>]*>/i);
      console.log(faviconMatch);
      if (faviconMatch && faviconMatch[1]) {
        let faviconUrl = faviconMatch[1];
        
        // 处理相对路径
        if (faviconUrl.startsWith('/')) {
          faviconUrl = `http://${host}${faviconUrl}`;
        } else if (!faviconUrl.startsWith('http')) {
          faviconUrl = `http://${host}/${faviconUrl}`;
        }

        const favicon = await getFavicon(faviconUrl);
        setResponseHeaders();
        return favicon;
      }

      // 如果都失败了返回默认图标
      throw createError({
        statusCode: 404,
        message: "Favicon not found",
      });
    }
  } catch (error) {
    throw createError({
      statusCode: 500,
      message: "Failed to fetch favicon",
    });
  }
});
