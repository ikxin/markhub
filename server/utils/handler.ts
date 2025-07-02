import type { EventHandler, EventHandlerRequest } from "h3";

export const defineResponseHandler = <T extends EventHandlerRequest, D>(
  handler: EventHandler<T, D>
): EventHandler<T, D> =>
  defineEventHandler<T>(async (event) => {
    const buffer = await handler(event);

    event.node.res.setHeader("Content-Type", "image/png");
    event.node.res.setHeader("Cache-Control", `max-age=${60 * 60 * 24 * 30}`);

    return buffer;
  });
