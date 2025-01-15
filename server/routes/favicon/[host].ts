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
    let res = await fetch(`http://${host}/favicon.ico`);

    if (res.ok && res.headers.get("content-type")?.includes("image")) {
      event.node.res.setHeader("Content-Type", "image/png");
      event.node.res.setHeader("Cache-Control", "max-age=3600, public");
      return res.body;
    }

    const html = await fetch(`http://${host}`).then((res) => res.text());

    const linkTag = html.match(
      /((<link[^>]+rel=.(icon|shortcut icon|alternate icon|apple-touch-icon)[^>]+>))/i
    );

    if (!linkTag) throw new Error();

    const linkRef = linkTag[1].match(/href=["']([^"']+)["']/i);

    if (!linkRef) throw new Error();

    let iconUrl = linkRef[1];

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

    res = await fetch(iconUrl);

    if (res.ok && res.headers.get("content-type")?.includes("image")) {
      event.node.res.setHeader("Content-Type", "image/png");
      event.node.res.setHeader("Cache-Control", "max-age=3600, public");
      return res.body;
    } else {
      throw new Error();
    }
  } catch (error) {
    event.node.res.setHeader("Content-Type", "image/svg+xml");
    return generateRandomIcon(host?.charAt(0).toUpperCase());
  }
});
