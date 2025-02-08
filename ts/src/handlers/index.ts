import cors from "@elysiajs/cors";
import Elysia from "elysia";

export const app = new Elysia({ aot: false })
	.use(
		cors({
			aot: false,
			origin: "localhost:3000",
			credentials: true,
		}),
	)
	.onRequest(({ set }) => {
		set.headers["access-control-allow-credentials"] = "true";
	})
	.get("/ping", () => "pong");
