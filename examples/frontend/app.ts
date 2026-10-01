import { App } from "@modelcontextprotocol/ext-apps";
import { OpenAIExtensions, OpenAIFileEntrypointInputSchema } from "@openai/mcp-extensions/app";

const app = new App({ name: "Go workspace", version: "0.0.0-dev" }, {});
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

function refreshActions() {
  for (const button of actions) button.disabled = !connected || busy;
  (element("context") as HTMLButtonElement).disabled = !connected || busy || !extensions.modelContext;
  (element("message") as HTMLButtonElement).disabled = !connected || busy || !extensions.message;
}

async function run(action: () => Promise<unknown>) {
  busy = true;
  refreshActions();
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
  if (typeof values?.units === "string") (element("units") as HTMLSelectElement).value = values.units;
  if (typeof values?.showGrid === "boolean") (element("grid") as HTMLInputElement).checked = values.showGrid;
  return result;
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
  refreshActions();
};

element("read").addEventListener("click", () => void run(() => settings("settings.read", {})));
element("save").addEventListener("click", () => void run(() => settings("settings.update", {
  set: { units: (element("units") as HTMLSelectElement).value, showGrid: (element("grid") as HTMLInputElement).checked },
})));
element("search").addEventListener("click", () => void run(() => call("search_mentions", { query: (element("query") as HTMLInputElement).value })));
element("context").addEventListener("click", () => void run(async () => {
  return extensions.modelContext!.update({ content: [{ type: "text", text: "Selected demo part: bolt", _meta: { "openai/title": "Bolt" } }] });
}));
element("message").addEventListener("click", () => void run(async () => {
  return extensions.message!.send({ role: "user", content: [{ type: "text", text: "Tell me about the selected demo part." }] });
}));
element("display").addEventListener("click", () => void run(() => app.requestDisplayMode({ mode: "fullscreen" })));

refreshActions();
if (window.parent === window) {
  status("Open this App through an MCP host");
} else {
  try {
    await app.connect();
    connected = true;
    element("deep-link").textContent = extensions.deepLink.getCurrent()?.url ?? "/";
    status("Connected");
    refreshActions();
  } catch (error) {
    status(error instanceof Error ? error.message : "Host initialization failed");
  }
}
