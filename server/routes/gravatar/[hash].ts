import * as z from "zod/v4";
import crypto from "crypto";

function getHash(email: string) {
  return crypto
    .createHash("md5")
    .update(email.trim().toLowerCase())
    .digest("hex");
}

const paramSchema = z.object({
  hash: z.union([
    z.string().regex(/^([a-f0-9]{32}|[a-f0-9]{64})$/i),
    z.email().transform((email) => getHash(email)),
  ]),
});

const querySchema = z.object({
  size: z.coerce.number().int().min(1).max(2048).optional(),
  s: z.coerce.number().int().min(1).max(2048).catch(100),
});

export default defineResponseHandler(async (event) => {
  try {
    const query = await getValidatedQuery(event, querySchema.parse);
    const params = await getValidatedRouterParams(event, paramSchema.safeParse);

    if (!params.success) throw new Error();

    const response = await fetch(
      `https://cravatar.com/avatar/${params.data.hash}?s=${
        query.size || query.s
      }`
    );

    if (response.ok) {
      return Buffer.from(await response.arrayBuffer());
    } else {
      throw new Error();
    }
  } catch {
    return useStorage("assets:server").getItemRaw("fallback/gravatar.png");
  }
});
