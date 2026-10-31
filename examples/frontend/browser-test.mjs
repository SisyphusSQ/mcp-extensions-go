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
      } else if (req.method === "GET" && (req.url === "/app" || req.url === "/app?webmcp")) {
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
  // This test-only injected API exercises registration/execution, not host discovery.
  const browserContext = await browser.newContext();
  await browserContext.addInitScript(() => {
    if (location.pathname !== "/app" || !location.search.includes("webmcp")) return;
    const tools = new Map();
    Object.defineProperty(document, "modelContext", { value: { registerTool(tool) {
      if (tools.has(tool.name)) throw new Error("Duplicate WebMCP tool");
      tools.set(tool.name, tool);
    } } });
    window.webmcpFixture = tools;
  });
  page = await browserContext.newPage();
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
  // Simulate another App changing an unrelated setting after this App's read.
  await client.callTool({ name: "settings.update", arguments: { set: { showGrid: false } } });
  await app.locator("#units").selectOption("mm"); await app.locator("#save").click();
  await app.locator("#output").filter({ hasText: '"units": "mm"' }).waitFor();
  assert.equal(await app.locator("#grid").isChecked(), false, "App overwrote another instance's unrelated patch");
  await app.locator("#query").fill("bolt"); await app.locator("#search").click();
  await app.locator("#output").filter({ hasText: "parts://bolt" }).waitFor();
  await app.locator("#context").click(); await app.locator("#output").filter({ hasText: "update-1" }).waitFor();
  await app.locator("#message").click();
  await page.waitForFunction(() => window.fixture.events.some(e => e.method === "ui/message"));
  await app.locator("#draft").click();
  await page.waitForFunction(() => window.fixture.events.filter(e => e.method === "ui/message").length === 2);
  await app.locator("#new-draft").click();
  await page.waitForFunction(() => window.fixture.events.filter(e => e.method === "ui/message").length === 3);
  await app.locator("#display").click(); await app.locator("#output").filter({ hasText: "fullscreen" }).waitFor();
  assert.equal(await app.locator("#display").textContent(), "Already fullscreen");
  assert.equal(await app.locator("#display").isDisabled(), true);
  await page.evaluate(() => window.fixture.deepLink("/parts/washer"));
  await app.locator("#deep-link").filter({ hasText: "/parts/washer" }).waitFor();
  await page.evaluate(() => window.fixture.file());
  await app.locator("#output").filter({ hasText: "fixture.txt" }).waitFor();
  assert.equal(await app.locator("#output img").count(), 0, "File HTML became executable DOM");
  const events = await page.evaluate(() => window.fixture.events);
  const message = events.find(e => e.method === "ui/message");
  assert.equal(message.params.role, "user");
  assert.deepEqual(events.filter(e => e.method === "ui/message").slice(1).map(e => e.params._meta["openai/message"]), [{ target: "active", send: false }, { target: "new", send: false }]);
  assert.equal(events.filter(e => e.method === "tools/call" && e.params.name === "open_workspace").length, 0, "App redundantly called the initial tool");
  assert.equal(events.find(e => e.method === "ui/update-model-context").params.content[0]._meta["openai/title"], "Bolt");
  assert.equal(events.find(e => e.method === "resources/read").params._meta["openai/resource"].representation, "text");
  assert.deepEqual(events.filter(e => e.method === "tools/call" && e.params.name === "settings.update")[1].params.arguments.set, { units: "mm" });
  assert.deepEqual(errors, []);
  await page.goto(`http://127.0.0.1:${local.address().port}/?noextensions`);
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
  assert.equal(await app.locator("#context").isDisabled(), true);
  assert.equal(await app.locator("#message").isDisabled(), true);
  assert.equal(await app.locator("#draft").isDisabled(), true);
  assert.equal(await app.locator("#new-draft").isDisabled(), true);
  await page.goto(`http://127.0.0.1:${local.address().port}/?mobile`);
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
  assert.equal(await app.locator("#message").isDisabled(), false);
  for (const id of ["draft", "new-draft", "new-message"]) assert.equal(await app.locator(`#${id}`).isDisabled(), false);
  for (const id of ["message", "new-message", "draft", "new-draft"]) {
    await app.locator(`#${id}`).click();
    await app.locator("#status").filter({ hasText: "Ready" }).waitFor();
  }
  assert.deepEqual((await page.evaluate(() => window.fixture.events)).filter(e => e.method === "ui/message").map(e => e.params._meta["openai/message"]), [
    { target: "active", send: true }, { target: "new", send: true }, { target: "active", send: false }, { target: "new", send: false },
  ]);
  // Groups, external single/group/all removal, and preserving unsent selections.
  await page.goto(`http://127.0.0.1:${local.address().port}/`);
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
  for (const id of ["washer", "note", "workspace"]) await app.locator(`#select-${id}`).check();
  await app.locator("#context").click();
  await app.locator("#status").filter({ hasText: "Ready" }).waitFor();
  const groups = (await page.evaluate(() => window.fixture.events)).find(e => e.method === "ui/update-model-context").params.content;
  assert.deepEqual(groups.map(b => b._meta["openai/group"]?.title), ["Parts", "Parts", "Notes", undefined]);
  await page.evaluate(() => window.fixture.context(["washer", "note", "workspace"]));
  await app.locator("#select-bolt:not(:checked)").waitFor();
  await page.evaluate(() => window.fixture.context(["note", "workspace"]));
  await app.locator("#select-washer:not(:checked)").waitFor();
  await app.locator("#context").click();
  await app.locator("#status").filter({ hasText: "Ready" }).waitFor();
  let contexts = (await page.evaluate(() => window.fixture.events)).filter(e => e.method === "ui/update-model-context");
  assert.deepEqual(contexts.at(-1).params.content.map(b => b._meta["example/selectionId"]), ["note", "workspace"]);
  await app.locator("#select-bolt").check();
  await page.evaluate(() => window.fixture.context(["workspace"]));
  await app.locator("#select-note:not(:checked)").waitFor();
  assert.equal(await app.locator("#select-bolt").isChecked(), true, "Unsent local selection was lost");
  await app.locator("#context").click();
  await app.locator("#status").filter({ hasText: "Ready" }).waitFor();
  await page.evaluate(() => window.fixture.context(null));
  await app.locator("#select-bolt:not(:checked)").waitFor();
  await app.locator("#select-workspace:not(:checked)").waitFor();
  for (const query of ["restore", "cleared", "remove-during-update"]) {
    await page.goto(`http://127.0.0.1:${local.address().port}/?${query}`);
    await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
    if (query === "restore") {
      assert.equal(await app.locator("#select-washer").isChecked(), true);
      assert.equal(await app.locator("#select-bolt").isChecked(), false);
    } else if (query === "cleared") {
      assert.equal(await app.locator("#selections input:checked").count(), 0);
    } else {
      await app.locator("#context").click();
      await app.locator("#status").filter({ hasText: "Ready" }).waitFor();
      await app.locator("#select-bolt:not(:checked)").waitFor();
    }
  }
  const other = await browserContext.newPage();
  await other.goto(`http://127.0.0.1:${local.address().port}/?restore`);
  await other.frameLocator("#app").locator("#status").filter({ hasText: "Connected" }).waitFor();
  await page.evaluate(() => window.fixture.context(null));
  assert.equal(await other.frameLocator("#app").locator("#select-washer").isChecked(), true, "Other App instance was cleared");
  await other.close();
  await page.goto(`http://127.0.0.1:${local.address().port}/?webmcp`);
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
  await app.locator("#webmcp").filter({ hasText: "Registered" }).waitFor();
  const appFrame = () => page.frames().find(frame => frame.url().includes("/app"));
  const invoke = (name, input) => appFrame().evaluate(async ({ name, input }) => {
    try { return { value: await window.webmcpFixture.get(name).execute(input) }; }
    catch (error) { return { error: error.message }; }
  }, { name, input });
  assert.deepEqual(await appFrame().evaluate(() => [...window.webmcpFixture.keys()]), ["read_view", "search_parts", "select_part"]);
  assert.equal(await appFrame().evaluate(() => window.webmcpFixture.get("read_view").annotations.readOnlyHint), true);
  assert.equal(JSON.parse((await invoke("read_view", {})).value).selected[0], "bolt");
  assert.ok((await invoke("read_view", { extra: true })).error);
  const search = await invoke("search_parts", { query: "washer" });
  assert.ok(search.value.includes("parts://washer"));
  assert.equal(await app.locator("#query").inputValue(), "washer");
  const selection = await invoke("select_part", { part: "washer", selected: true });
  assert.deepEqual(JSON.parse(selection.value).selected, ["bolt", "washer"]);
  assert.equal(await app.locator("#select-washer").isChecked(), true);
  for (const input of [{ part: "unknown", selected: true }, { part: "bolt", selected: "yes" }, { part: "bolt", selected: true, tool: "settings.update" }]) {
    assert.ok((await invoke("select_part", input)).error);
  }
  const concurrent = await appFrame().evaluate(async () => Promise.allSettled([
    window.webmcpFixture.get("search_parts").execute({ query: "bolt" }),
    window.webmcpFixture.get("select_part").execute({ part: "bolt", selected: false }),
  ]).then(results => results.map(result => result.status === "rejected" ? result.reason.message : "ok")));
  assert.deepEqual(concurrent, ["ok", "Another operation is in progress"]);
  await page.evaluate(() => window.fixture.context([]));
  await app.locator("#select-washer:not(:checked)").waitFor();
  assert.deepEqual(JSON.parse((await invoke("read_view", {})).value).selected, []);
  await app.locator("#select-bolt").check();
  await app.locator("#context").click();
  await app.locator("#status").filter({ hasText: "Ready" }).waitFor();
  assert.deepEqual(JSON.parse((await invoke("read_view", {})).value).selected, ["bolt"]);
  await page.goto(`http://127.0.0.1:${local.address().port}/?noextensions&webmcp`);
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
  assert.equal((await invoke("select_part", { part: "washer", selected: true })).error, "The host does not support model context");
  assert.equal(await app.locator("#select-washer").isChecked(), false);
  for (const query of ["mobile&webmcp", "web&webmcp", ""]) {
    await page.goto(`http://127.0.0.1:${local.address().port}/?${query}`);
    await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
    assert.equal(await app.locator("#webmcp").textContent(), "Unavailable on this host");
    if (query) assert.equal(await appFrame().evaluate(() => window.webmcpFixture.size), 0);
  }
  await page.goto(`http://127.0.0.1:${local.address().port}/?settingsfail&webmcp`);
  await app.locator("#status").filter({ hasText: "Settings fixture unavailable" }).waitFor();
  assert.equal(await app.locator("#save").isDisabled(), true);
  assert.equal(await app.locator("#search").isDisabled(), false);
  assert.ok((await invoke("search_parts", { query: "bolt" })).value.includes("parts://bolt"));
  await page.goto(`http://127.0.0.1:${local.address().port}/?initialfile`);
  await page.waitForFunction(() => window.fixture.events.some(e => e.method === "resources/read"));
  assert.equal(await app.locator("#output img").count(), 0);
  for (const [query, label] of [["fullscreen", "Already fullscreen"], ["nofullscreen", "Fullscreen unavailable"]]) {
    await page.goto(`http://127.0.0.1:${local.address().port}/?${query}`);
    await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
    assert.equal(await app.locator("#display").textContent(), label);
    assert.equal(await app.locator("#display").isDisabled(), true);
    assert.equal((await page.evaluate(() => window.fixture.events)).filter(e => e.method === "ui/request-display-mode").length, 0);
  }
  await page.goto(`http://127.0.0.1:${local.address().port}/?keepinline`);
  await app.locator("#status").filter({ hasText: "Connected" }).waitFor();
  await app.locator("#display").click();
  await app.locator("#status").filter({ hasText: "Host kept inline mode" }).waitFor();
  assert.equal(await app.locator("#display-mode").textContent(), "inline");
  const standalone = await browser.newPage();
  await standalone.goto(`http://127.0.0.1:${local.address().port}/app`);
  await standalone.locator("#status").filter({ hasText: "Open this App through an MCP host" }).waitFor();
  assert.deepEqual(errors, []);
  console.log("PASS: official AppBridge/Go HTTP, grouped context deletion/echo/remount/in-flight/isolation, mobile target/send matrix, WebMCP registration/validation/execution/live UI state/unsupported hosts, settings/files/display and standalone");
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
