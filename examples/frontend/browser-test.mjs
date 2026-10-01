import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";
import { once } from "node:events";
import { createServer } from "node:http";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { spawn } from "node:child_process";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";
import { build } from "esbuild";

// Use an installed browser; this test never downloads browsers or global tools.
const playwright = process.env.PLAYWRIGHT_MODULE_PATH
  ? await import(pathToFileURL(process.env.PLAYWRIGHT_MODULE_PATH).href)
  : await import("playwright");
const executablePath = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH;
const token = randomBytes(32).toString("hex");
const child = spawn(resolve("../../bin/mcp-extensions-http"), [], {
  env: { ...process.env, MCP_BEARER_TOKEN: token, MCP_LISTEN_ADDR: "127.0.0.1:0", MCP_APP_HTML: resolve("dist/app.html") },
  stdio: ["ignore", "ignore", "pipe"],
});
let browser, local, client, page;
try {
  const endpoint = await new Promise((resolveEndpoint, reject) => {
    const timer = setTimeout(() => reject(new Error("Go server startup timed out")), 10000);
    child.once("error", error => { clearTimeout(timer); reject(error); });
    child.once("exit", code => { clearTimeout(timer); reject(new Error(`Go server exited: ${code}`)); });
    child.stderr.on("data", chunk => {
      const address = /address=(127\.0\.0\.1:\d+)/.exec(chunk.toString())?.[1];
      if (address) { clearTimeout(timer); resolveEndpoint(`http://${address}/mcp`); }
    });
  });
  client = new Client({ name: "local-browser-test", version: "0" });
  await client.connect(new StreamableHTTPClientTransport(new URL(endpoint), { requestInit: { headers: { Authorization: `Bearer ${token}` } } }));
  const resource = await client.readResource({ uri: "ui://workspace/home.html" });
  assert.equal(resource.contents[0].mimeType, "text/html;profile=mcp-app");
  const appHTML = resource.contents[0].text;
  assert.equal(typeof appHTML, "string");
  assert.ok(!appHTML.includes(token), "App received a server credential");
  const hostJS = (await build({ entryPoints: ["test-host.ts"], bundle: true, write: false, format: "esm", target: "es2022" })).outputFiles[0].text;
  const hostHTML = '<!doctype html><title>Local bridge fixture</title><iframe id="app" sandbox="allow-scripts" style="width:900px;height:900px"></iframe><script type="module" src="/host.js"></script>';
  const names = new Set(["open_workspace", "open_file", "settings.read", "settings.update", "search_mentions"]);
  local = createServer(async (req, res) => {
    try {
      if (req.method === "GET" && (req.url === "/" || req.url?.startsWith("/?"))) {
        res.setHeader("Content-Type", "text/html"); res.end(hostHTML);
      } else if (req.method === "GET" && req.url === "/host.js") {
        res.setHeader("Content-Type", "text/javascript"); res.end(hostJS);
      } else if (req.method === "GET" && req.url === "/app") {
        res.setHeader("Content-Type", "text/html"); res.end(appHTML);
      } else if (req.method === "GET" && req.url === "/favicon.ico") {
        res.writeHead(204); res.end();
      } else if (req.method === "POST" && req.url === "/rpc") {
        const origin = `http://127.0.0.1:${local.address().port}`;
        if (req.headers.origin !== origin || req.headers["content-type"] !== "application/json") { res.writeHead(403); res.end(); return; }
        const chunks = []; let bytes = 0;
        for await (const chunk of req) { bytes += chunk.length; if (bytes > 1 << 20) throw new Error("Request too large"); chunks.push(chunk); }
        const { method, params } = JSON.parse(Buffer.concat(chunks).toString());
        let result;
        if (method === "tools/call" && names.has(params?.name)) result = await client.callTool(params);
        else if (method === "resources/read" && ["parts://bolt", "parts://washer"].includes(params?.uri)) result = await client.readResource(params);
        else { res.writeHead(403); res.end(); return; }
        res.setHeader("Content-Type", "application/json"); res.end(JSON.stringify(result));
      } else { res.writeHead(404); res.end(); }
    } catch { res.writeHead(500); res.end("Local fixture request failed"); }
  });
  local.listen(0, "127.0.0.1"); await once(local, "listening");
  browser = await playwright.chromium.launch({ headless: true, ...(executablePath ? { executablePath } : {}) });
  page = await browser.newPage();
  const errors = [];
  page.on("pageerror", error => errors.push(error.message));
  page.on("console", message => { if (message.type() === "error") console.error(message.text()); });
  await page.goto(`http://127.0.0.1:${local.address().port}/`);
  const app = page.frameLocator("#app");
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor({ timeout: 10000 });
  await app.locator("#output").filter({ hasText: "Welcome to your workspace" }).waitFor();
  assert.equal(await app.locator("#deep-link").textContent(), "/parts?tag=bolt");
  await app.locator("#read").click();
  await app.locator("#output").filter({ hasText: '"units": "mm"' }).waitFor();
  await app.locator("#units").selectOption("in"); await app.locator("#grid").check(); await app.locator("#save").click();
  await app.locator("#output").filter({ hasText: '"units": "in"' }).waitFor();
  await app.locator("#read").click();
  await app.locator("#output").filter({ hasText: '"showGrid": true' }).waitFor();
  assert.equal(await app.locator("#units").inputValue(), "in");
  assert.equal(await app.locator("#grid").isChecked(), true);
  await app.locator("#query").fill("bolt"); await app.locator("#search").click();
  await app.locator("#output").filter({ hasText: "parts://bolt" }).waitFor();
  await app.locator("#context").click(); await app.locator("#output").filter({ hasText: "update-1" }).waitFor();
  await app.locator("#message").click();
  await page.waitForFunction(() => window.fixture.events.some(e => e.method === "ui/message"));
  await app.locator("#display").click(); await app.locator("#output").filter({ hasText: "fullscreen" }).waitFor();
  await page.evaluate(() => window.fixture.deepLink("/parts/washer"));
  await app.locator("#deep-link").filter({ hasText: "/parts/washer" }).waitFor();
  await page.evaluate(() => window.fixture.file());
  await app.locator("#output").filter({ hasText: "fixture.txt" }).waitFor();
  assert.equal(await app.locator("#output img").count(), 0, "File HTML became executable DOM");
  const events = await page.evaluate(() => window.fixture.events);
  const message = events.find(e => e.method === "ui/message");
  assert.equal(message.params.role, "user");
  assert.equal(events.filter(e => e.method === "tools/call" && e.params.name === "open_workspace").length, 0, "App redundantly called the initial tool");
  assert.equal(events.find(e => e.method === "ui/update-model-context").params.content[0]._meta["openai/title"], "Bolt");
  assert.equal(events.find(e => e.method === "resources/read").params._meta["openai/resource"].representation, "text");
  assert.deepEqual(errors, []);
  await page.goto(`http://127.0.0.1:${local.address().port}/?noextensions`);
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
  assert.equal(await app.locator("#context").isDisabled(), true);
  assert.equal(await app.locator("#message").isDisabled(), true);
  const standalone = await browser.newPage();
  await standalone.goto(`http://127.0.0.1:${local.address().port}/app`);
  await standalone.locator("#status").filter({ hasText: "Open this App through an MCP host" }).waitFor();
  console.log("PASS: official App/AppBridge handshake, Go HTTP settings/search, initial result, display/deep-link updates, context/message payloads, file text, absent capabilities, standalone notice");
} catch (error) {
  if (page) {
    for (const frame of page.frames()) console.error("Fixture frame:", await frame.locator("body").innerText().catch(() => "unavailable"));
  }
  throw error;
} finally {
  await browser?.close();
  if (local) { local.closeAllConnections(); await new Promise(resolveClose => local.close(resolveClose)); }
  await client?.close();
  if (child.exitCode === null) { const exited = once(child, "exit"); child.kill("SIGTERM"); await exited; }
}
