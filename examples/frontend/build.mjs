import { build } from "esbuild";
import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";

await mkdir("dist", { recursive: true });
const result = await build({ entryPoints: ["app.ts"], bundle: true, write: false, format: "esm", target: "es2022", minify: true });
// Escape HTML closing tags without changing the JavaScript string contents.
const script = result.outputFiles[0].text.replace(/<\/script/gi, "<\\/script");
const hash = createHash("sha256").update(script).digest("base64");
const template = await readFile("index.html", "utf8");
await writeFile("dist/app.html", template.replace("SCRIPT_HASH", `'sha256-${hash}'`).replace("APP_SCRIPT", () => script));
console.log("Built dist/app.html using official App and OpenAI extensions");
