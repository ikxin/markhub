function generateRandomIcon(val?: string) {
  // Limit to within 200 to make the color not too bright.
  const [r, g, b] = Array.from({ length: 3 }, () =>
    Math.floor(Math.random() * 200)
  );

  const svg = `<svg width="64" height="64" xmlns="http://www.w3.org/2000/svg">
    <rect width="64" height="64" rx="16" ry="16" fill="rgb(${r},${g},${b})"/>
    <text 
      x="32"
      y="32"
      font-family="Arial" 
      font-size="32"
      font-weight="bold"
      fill="#FFFFFF"
      text-anchor="middle"
      alignment-baseline="central"
      dominant-baseline="central"
    >${val}</text>
  </svg>`;

  return Buffer.from(svg);
}

export default defineEventHandler(async (event) => {
  const host = getRouterParam(event, "host");

  try {
    const html = await fetch(`http://${host}`).then((res) => res.text());

    // Try to find icon link tag
    const linkTag = html.match(
      /<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+)["'][^>]*>/i
    );

    if (linkTag && linkTag[1]) {
      let iconUrl = linkTag[1].trim();

      if (iconUrl.startsWith("//")) {
        iconUrl = `http:${iconUrl}`;
      }

      if (iconUrl.startsWith("./")) {
        iconUrl = `http://${host}${iconUrl.slice(1)}`;
      }

      if (iconUrl.startsWith("/")) {
        iconUrl = `http://${host}${iconUrl}`;
      }

      if (!iconUrl.startsWith("http")) {
        iconUrl = `http://${host}/${iconUrl}`;
      }

      const res = await fetch(iconUrl);
      if (res.ok) {
        event.node.res.setHeader(
          "Content-Type",
          iconUrl.endsWith(".svg") ? "image/svg+xml" : "image/png"
        );
        return res.body;
      }
    }

    // Fallback to favicon.ico
    const res = await fetch(`http://${host}/favicon.ico`);
    if (res.ok) {
      event.node.res.setHeader("Content-Type", "image/png");
      return res.body;
    }
  } catch (error) {
    event.node.res.setHeader("Content-Type", "image/svg+xml");
    return generateRandomIcon(host?.charAt(0).toUpperCase());
  }
});
