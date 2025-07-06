FROM node:lts as builder

WORKDIR /app

COPY . .

RUN corepack enable \
  && pnpm install \
  && NITRO_PRESET=node-server pnpm run build

FROM node:lts

COPY --from=builder /app/.output /app/.output

WORKDIR /app

EXPOSE 3000

ENV HOST=0.0.0.0
ENV PORT=3000

CMD ["node", ".output/server/index.mjs"]
