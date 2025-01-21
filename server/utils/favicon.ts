import sharp from "sharp";
import icoToPng from "ico-to-png";

const LINK_TAG_REGEX =
  /((<link[^>]+rel=.(icon|shortcut icon|alternate icon|apple-touch-icon)[^>]+>))/i;

const LINK_REF_REGEX = /href=["']([^"']+)["']/i;

export const getIcoByFavicon = async (host: string): Promise<Buffer> => {
  const data = await fetch(`http://${host}/favicon.ico`).then((res) =>
    res.arrayBuffer()
  );
  return await icoToPng(Buffer.from(data), 32);
};

export const getIcoByLinkTag = async (host: string): Promise<Buffer> => {
  const htmlStr = await fetch(`http://${host}`).then((res) => res.text());

  const linkTag = htmlStr.match(LINK_TAG_REGEX);
  if (!linkTag) throw new Error();

  const linkRef = linkTag[1].match(LINK_REF_REGEX);
  if (!linkRef) throw new Error();

  let [_, iconUrl] = linkRef;
  iconUrl = new URL(iconUrl, `http://${host}`).toString();

  const data = await fetch(iconUrl).then((res) => res.arrayBuffer());
  const buffer = Buffer.from(data);

  if (iconUrl.endsWith(".ico")) {
    return icoToPng(Buffer.from(buffer), 32);
  }

  return sharp(Buffer.from(buffer)).png().toBuffer();
};

export const getIcoByWebmaster = async (host: string): Promise<Buffer> => {
  const url = `https://favicon.im/${host}`;
  const data = await fetch(url).then((res) => res.arrayBuffer());
  return sharp(Buffer.from(data)).png().toBuffer();
};

export const getIcoByGoogle = async (host: string): Promise<Buffer> => {
  const url = `https://www.google.com/s2/favicons?domain=${host}&sz=64`;
  const data = await fetch(url).then((res) => res.arrayBuffer());
  return sharp(Buffer.from(data)).png().toBuffer();
};

export const generateStringIco = (val?: string): Promise<Buffer> => {
  const [r, g, b] = Array.from({ length: 3 }, () =>
    Math.floor(Math.random() * 200)
  );

  const svg = `<svg width="64" height="64">
    <rect width="64" height="64" rx="16" ry="16" fill="rgb(${r},${g},${b})"/>
    <text 
      x="32"
      y="44"
      font-family="Arial"
      font-size="36"
      font-weight="bold"
      fill="#FFFFFF"
      text-anchor="middle"
    >${process.env.VERCEL ? "" : val}</text>
  </svg>`;

  return sharp(Buffer.from(svg)).png().toBuffer();
};
