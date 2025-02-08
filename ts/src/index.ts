import { Env } from "bun";
import { Context } from "elysia";
import { app } from "./handlers";

export default {
	async fetch(request: Request, env: Env, ctx: Context): Promise<Response> {
		return await app.handle(request);
	},
};
