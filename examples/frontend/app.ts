import { App } from "@modelcontextprotocol/ext-apps";
import { OpenAIExtensions, OpenAIFileEntrypointInputSchema, type OpenAIModelContextHostState } from "@openai/mcp-extensions/app";
import { z } from "zod";

const app = new App({ name: "Go workspace", version: "0.0.3" }, { availableDisplayModes: ["inline", "fullscreen"] });
const extensions = new OpenAIExtensions(app);
const element = (id: string): HTMLElement => {
  const value = document.getElementById(id);
  if (!value) throw new Error(`Missing element: ${id}`);
  return value;
};
const status = (message: string) => { element("status").textContent = message; };
const output = (value: unknown) => { element("output").textContent = JSON.stringify(value, null, 2); };
const entries = [
  { id: "bolt", title: "Bolt", text: "Selected demo part: bolt", group: "Parts" },
  { id: "washer", title: "Washer", text: "Selected demo part: washer", group: "Parts" },
  { id: "note", title: "Workshop note", text: "Compare the selected parts before assembly.", group: "Notes" },
  { id: "workspace", title: "Workspace", text: "The user is viewing the Go workspace.", group: undefined },
];
const selected = new Set<string>(["bolt"]);
const selectionInputs = new Map<string, HTMLInputElement>();
for (const entry of entries) {
  const label = document.createElement("label");
  const input = document.createElement("input");
  input.type = "checkbox";
  input.id = `select-${entry.id}`;
  input.addEventListener("change", () => {
    if (input.checked) selected.add(entry.id); else selected.delete(entry.id);
  });
  label.append(input, document.createTextNode(entry.title));
  element("selections").append(label);
  selectionInputs.set(entry.id, input);
}
const actions = Array.from(document.querySelectorAll<HTMLButtonElement>("button"));
let connected = false;
let busy = false;
let currentValues: { units: string; showGrid: boolean } | undefined;
let displayMode: string | undefined;
let submittedIds = new Set<string>();
let ownUpdateId: string | undefined;
let contextPending = false;
let pendingHostState: OpenAIModelContextHostState | undefined;

function refreshActions() {
  for (const button of actions) button.disabled = !connected || busy;
  for (const [id, input] of selectionInputs) {
    input.checked = selected.has(id);
    input.disabled = !connected || busy || !extensions.modelContext;
  }
  (element("save") as HTMLButtonElement).disabled = !connected || busy || !currentValues;
  (element("context") as HTMLButtonElement).disabled = !connected || busy || !extensions.modelContext;
  for (const id of ["message", "new-message", "draft", "new-draft"]) {
    (element(id) as HTMLButtonElement).disabled = !connected || busy || !extensions.message;
  }
  const display = element("display") as HTMLButtonElement;
  const supported = app.getHostContext()?.availableDisplayModes?.includes("fullscreen") ?? false;
  display.disabled = !connected || busy || displayMode === "fullscreen" || !supported;
  display.textContent = displayMode === "fullscreen" ? "Already fullscreen" : supported ? "Open fullscreen" : "Fullscreen unavailable";
  element("display-mode").textContent = displayMode ?? "unknown";
}

// Registered WebMCP tools and UI actions share concurrency/error handling.
async function perform(action: () => Promise<unknown>) {
  if (busy) throw new Error("Another operation is in progress");
  busy = true;
  refreshActions();
  status("Working…");
  try {
    const result = await action();
    output(result);
    status("Ready");
    return result;
  } catch (error) {
    status(error instanceof Error ? error.message : "Operation failed");
    throw error;
  } finally {
    busy = false;
    refreshActions();
  }
}
async function run(action: () => Promise<unknown>) {
  try { await perform(action); } catch { /* perform already displays the error. */ }
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
  return Object.keys(set).length ? settings("settings.update", { set }) : { values: currentValues };
}
function view() {
  return { query: (element("query") as HTMLInputElement).value, selected: [...selected], settings: currentValues, displayMode };
}
async function search(query: string) {
  (element("query") as HTMLInputElement).value = query;
  return call("search_mentions", { query });
}
function hostIds(context: OpenAIModelContextHostState) {
  return new Set((context?.content ?? []).flatMap(block => {
    const id = block._meta?.["example/selectionId"];
    return typeof id === "string" && entries.some(entry => entry.id === id) ? [id] : [];
  }));
}
function restoreContext() {
  const context = extensions.modelContext?.getCurrent();
  if (context === undefined) return;
  selected.clear();
  for (const id of hostIds(context)) selected.add(id);
  submittedIds = new Set(selected);
  ownUpdateId = context?.updateId;
}
function reconcileContext(context: OpenAIModelContextHostState | undefined) {
  if (context === undefined) return;
  if (contextPending) { pendingHostState = context; return; }
  if (context !== null && context.updateId === ownUpdateId) return;
  const kept = hostIds(context);
  // Only remove previously submitted items; preserve unsent local selections.
  for (const id of submittedIds) if (!kept.has(id)) selected.delete(id);
  submittedIds = kept;
  refreshActions();
}
async function shareContext() {
  if (!extensions.modelContext) throw new Error("The host does not support model context");
  const previousIds = submittedIds;
  submittedIds = new Set(selected);
  contextPending = true;
  try {
    const result = await extensions.modelContext.update({
      content: entries.filter(entry => selected.has(entry.id)).map(entry => ({
        type: "text" as const, text: entry.text,
        _meta: { "openai/title": entry.title, "example/selectionId": entry.id, ...(entry.group ? { "openai/group": { title: entry.group } } : {}) },
      })),
    });
    ownUpdateId = result?.updateId;
    return result;
  } catch (error) {
    submittedIds = previousIds;
    throw error;
  } finally {
    contextPending = false;
    const context = pendingHostState;
    pendingHostState = undefined;
    reconcileContext(context);
  }
}
async function selectPart(part: "bolt" | "washer", include: boolean) {
  if (!extensions.modelContext) throw new Error("The host does not support model context");
  if (include) selected.add(part); else selected.delete(part);
  await shareContext();
  return view();
}
function registerAppTools() {
  if (!document.modelContext || app.getHostContext()?.platform !== "desktop") {
    element("webmcp").textContent = "Unavailable on this host";
    return;
  }
  const empty = z.strictObject({});
  const query = z.strictObject({ query: z.string() });
  const selection = z.strictObject({ part: z.enum(["bolt", "washer"]), selected: z.boolean() });
  document.modelContext.registerTool({
    name: "read_view", description: "Read this workspace's query, selected context, settings and display mode.",
    inputSchema: z.toJSONSchema(empty), annotations: { readOnlyHint: true },
    async execute(input) { empty.parse(input); return JSON.stringify(view()); },
  });
  document.modelContext.registerTool({
    name: "search_parts", description: "Search readable demo parts in this workspace.", inputSchema: z.toJSONSchema(query),
    async execute(input) { const args = query.parse(input); return JSON.stringify(await perform(() => search(args.query))); },
  });
  document.modelContext.registerTool({
    name: "select_part", description: "Select or remove a demo part in this workspace's attached model context.", inputSchema: z.toJSONSchema(selection),
    async execute(input) { const args = selection.parse(input); return JSON.stringify(await perform(() => selectPart(args.part, args.selected))); },
  });
  element("webmcp").textContent = "Registered (Desktop expanded App required)";
}

// Register listeners before connecting so the first result is never missed.
app.ontoolresult = result => { output(result.structuredContent); };
app.ontoolinput = async ({ arguments: args }) => {
  const input = OpenAIFileEntrypointInputSchema.safeParse(args);
  if (!input.success) return;
  if (!extensions.resources) { status("The host does not support file resources"); return; }
  await run(async () => {
    const resource = await extensions.resources!.read({ uri: input.data.file.resourceUri, representation: "text" });
    return { file: input.data.file.name, contents: resource.contents };
  });
};
app.onhostcontextchanged = () => {
  element("deep-link").textContent = extensions.deepLink.getCurrent()?.url ?? "/";
  displayMode = app.getHostContext()?.displayMode;
  reconcileContext(extensions.modelContext?.getCurrent());
  refreshActions();
};
element("read").addEventListener("click", () => void run(() => settings("settings.read", {})));
element("save").addEventListener("click", () => void run(saveSettings));
element("search").addEventListener("click", () => void run(() => search((element("query") as HTMLInputElement).value)));
element("context").addEventListener("click", () => void run(shareContext));
for (const [id, target, send] of [["message", "active", true], ["new-message", "new", true], ["draft", "active", false], ["new-draft", "new", false]] as const) {
  element(id).addEventListener("click", () => void run(async () => extensions.message!.send({
    role: "user", content: [{ type: "text", text: "Tell me about the selected demo part." }], _meta: { "openai/message": { target, send } },
  })));
}
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
    restoreContext();
    registerAppTools();
    await settings("settings.read", {});
    status("Connected");
    refreshActions();
  } catch (error) {
    status(error instanceof Error ? error.message : "Host initialization failed");
    refreshActions();
  }
}
