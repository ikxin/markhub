export default defineEventHandler((event) => {
  // 定义需要检查 referer 的受保护路径
  const protectedPaths: string[] = [];

  // 检查当前路径是否需要 referer 验证
  const needsRefererCheck = protectedPaths.some((protectedPath) =>
    event.path.startsWith(protectedPath)
  );

  // 如果不需要验证则直接返回
  if (!needsRefererCheck) {
    return;
  }

  const allowedDomains =
    process.env.NUXT_ALLOWED_DOMAINS?.split(",")
      .map((item) => item.trim())
      .filter(Boolean) || [];
  const allowEmptyReferer = process.env.NUXT_ALLOW_EMPTY_REFERER === "true";

  const referer = getRequestHeader(event, "referer");

  // 如果允许空 referer 且当前请求没有 referer,则允许访问
  if (allowEmptyReferer && !referer) {
    return;
  }

  // 验证 referer 是否在允许的域名列表中
  if (!referer) {
    throw createError({
      statusCode: 403,
      message: "Invalid request source",
    });
  }

  // 将域名规则转换为正则表达式
  const isAllowed = allowedDomains.some((domain) => {
    const pattern = domain
      .replace(/\./g, "\\.") // 转义点号
      .replace(/\*/g, ".*"); // 将星号转换为正则表达式通配符
    const regex = new RegExp(pattern);
    return regex.test(new URL(referer).hostname);
  });

  if (!isAllowed) {
    throw createError({
      statusCode: 403,
      message: "Invalid request source",
    });
  }
});
