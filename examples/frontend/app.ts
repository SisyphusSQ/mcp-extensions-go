import { App } from "@modelcontextprotocol/ext-apps";
import { OpenAIExtensions, OpenAIFileEntrypointInputSchema } from "@openai/mcp-extensions/app";

const app = new App({ name: "Go workspace", version: "0.0.0-dev" }, { availableDisplayModes: ["inline", "fullscreen"] });
const extensions = new OpenAIExtensions(app);
const element = (id: string): HTMLElement => {
  const value = document.getElementById(id);
  if (!value) throw new Error(`Missing element: ${id}`);
  return value;
};
const status = (message: string) => { element("status").textContent = message; };
const output = (value: unknown) => { element("output").textContent = JSON.stringify(value, null, 2); };
const actions = Array.from(document.querySelectorAll<HTMLButtonElement>("button"));
let connected = false;
let busy = false;
let currentValues: { units: string; showGrid: boolean } | undefined;
let displayMode: string | undefined;

function refreshActions() {
  for (const button of actions) button.disabled = !connected || busy;
  (element("save") as HTMLButtonElement).disabled = !connected || busy || !currentValues;
  (element("context") as HTMLButtonElement).disabled = !connected || busy || !extensions.modelContext;
  (element("message") as HTMLButtonElement).disabled = !connected || busy || !extensions.message;
  const display = element("display") as HTMLButtonElement;
  const supported = app.getHostContext()?.availableDisplayModes?.includes("fullscreen") ?? false;
  display.disabled = !connected || busy || displayMode === "fullscreen" || !supported;
  display.textContent = displayMode === "fullscreen" ? "Already fullscreen" : supported ? "Open fullscreen" : "Fullscreen unavailable";
  element("display-mode").textContent = displayMode ?? "unknown";
}

async function run(action: () => Promise<unknown>) {
  busy = true;
  refreshActions();
  status("Working…");
  try {
    output(await action());
    status("Ready");
  } catch (error) {
    status(error instanceof Error ? error.message : "Operation failed");
  } finally {
    busy = false;
    refreshActions();
  }
}

async function call(name: string, args: Record<string, unknown>) {
  const result = await app.callServerTool({ name, arguments: args });
  if (result.isError) throw new Error(result.content.map(item => item.type === "text" ? item.text : "").join(" "));
  return result.structuredContent;
}

async function settings(name: string, args: Record<string, unknown>) {
  const result = await call(name, args);
  const values = result?.values as Record<string, unknown> | undefined;
  if (typeof values?.units !== "string" || typeof values.showGrid !== "boolean") throw new Error("Invalid settings response");
  currentValues = { units: values.units, showGrid: values.showGrid };
  (element("units") as HTMLSelectElement).value = values.units;
  (element("grid") as HTMLInputElement).checked = values.showGrid;
  return result;
}

async function saveSettings() {
  if (!currentValues) throw new Error("Read settings before saving");
  const units = (element("units") as HTMLSelectElement).value;
  const showGrid = (element("grid") as HTMLInputElement).checked;
  const set: Record<string, unknown> = {};
  if (units !== currentValues.units) set.units = units;
  if (showGrid !== currentValues.showGrid) set.showGrid = showGrid;
  // Send only changed fields so another App's unrelated patch is preserved.
  return Object.keys(set).length ? settings("settings.update", { set }) : { values: currentValues };
}

// Register listeners before connecting so the first result is never missed.
app.ontoolresult = result => { output(result.structuredContent); };
app.ontoolinput = async ({ arguments: args }) => {
  const input = OpenAIFileEntrypointInputSchema.safeParse(args);
  if (!input.success) return;
  if (!extensions.resources) {
    status("The host does not support file resources");
    return;
  }
  await run(async () => {
    const resource = await extensions.resources!.read({ uri: input.data.file.resourceUri, representation: "text" });
    // DOM text output preserves the trust boundary for file contents.
    return { file: input.data.file.name, contents: resource.contents };
  });
};
app.onhostcontextchanged = () => {
  element("deep-link").textContent = extensions.deepLink.getCurrent()?.url ?? "/";
  displayMode = app.getHostContext()?.displayMode;
  refreshActions();
};

element("read").addEventListener("click", () => void run(() => settings("settings.read", {})));
element("save").addEventListener("click", () => void run(saveSettings));
element("search").addEventListener("click", () => void run(() => call("search_mentions", { query: (element("query") as HTMLInputElement).value })));
element("context").addEventListener("click", () => void run(async () => {
  return extensions.modelContext!.update({ content: [{ type: "text", text: "Selected demo part: bolt", _meta: { "openai/title": "Bolt" } }] });
}));
element("message").addEventListener("click", () => void run(async () => {
  return extensions.message!.send({ role: "user", content: [{ type: "text", text: "Tell me about the selected demo part." }] });
}));
element("display").addEventListener("click", () => void run(async () => {
  const result = await app.requestDisplayMode({ mode: "fullscreen" });
  displayMode = result.mode;
  if (result.mode !== "fullscreen") throw new Error(`Host kept ${result.mode} mode; fullscreen was not opened`);
  return result;
}));

refreshActions();
if (window.parent === window) {
  status("Open this App through an MCP host");
} else {
  try {
    await app.connect();
    connected = true;
    displayMode = app.getHostContext()?.displayMode;
    element("deep-link").textContent = extensions.deepLink.getCurrent()?.url ?? "/";
    // Refresh persisted controls without replacing the initial launch result.
    try { await settings("settings.read", {}); }
    catch (error) {
      status(error instanceof Error ? error.message : "Settings load failed");
      refreshActions();
      throw error;
    }
    status("Connected");
    refreshActions();
  } catch (error) {
    status(error instanceof Error ? error.message : "Host initialization failed");
  }
}
