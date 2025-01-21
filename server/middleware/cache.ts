export default defineEventHandler((event) => {
  event.node.res.setHeader("Cache-Control", `max-age=${60 * 60 * 24 * 30}`);
});
