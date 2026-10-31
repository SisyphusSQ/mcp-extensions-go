// A local protocol fixture using the official AppBridge, not an OpenAI host.
import { AppBridge, PostMessageTransport, McpUiMessageRequestSchema } from "@modelcontextprotocol/ext-apps/app-bridge";
import { OpenAIMessageParamsSchema, type OpenAIModelContextHostState } from "@openai/mcp-extensions/app";

const iframe = document.getElementById("app") as HTMLIFrameElement;
const events: Array<{ method: string; params?: unknown }> = [];
// Extend the host request schema through the official protected hook. The core
// message schema alone strips OpenAI metadata before invoking onmessage.
class FixtureBridge extends AppBridge {
  enableOpenAIMessages() {
    this.replaceRequestHandler(McpUiMessageRequestSchema.extend({ params: OpenAIMessageParamsSchema }), async request => {
      events.push({ method: "ui/message", params: request.params });
      return {};
    });
  }
}
let sequence = 0;
const disabled = new URLSearchParams(window.location.search).has("noextensions");
const parameters = new URLSearchParams(window.location.search);
const initialMode = parameters.has("fullscreen") ? "fullscreen" : "inline";
let context: OpenAIModelContextHostState | undefined = parameters.has("restore") ? {
  updateId: "restored-1", content: [{ type: "text", text: "Washer", _meta: { "example/selectionId": "washer" } }],
} : parameters.has("cleared") ? null : undefined;
const bridge = new FixtureBridge(null, { name: "Local protocol fixture", version: "0" }, {
  serverTools: {}, serverResources: {}, message: {}, updateModelContext: { text: {} },
  experimental: disabled ? {} : { "openai/modelContext": {}, "openai/message": {}, "openai/resource": {} },
}, { hostContext: { platform: parameters.has("mobile") ? "mobile" : parameters.has("web") ? "web" : "desktop", displayMode: initialMode, availableDisplayModes: parameters.has("nofullscreen") ? ["inline"] : ["inline", "fullscreen"], "openai/deepLink": { url: "/parts?tag=bolt" }, ...(context !== undefined ? { "openai/modelContext": context } : {}) } });

async function proxy(method: "tools/call" | "resources/read", params: unknown) {
  const response = await fetch("/rpc", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ method, params }) });
  if (!response.ok) throw new Error("Local MCP proxy failed");
  return response.json();
}

bridge.oncalltool = async params => {
  events.push({ method: "tools/call", params });
  if (parameters.has("settingsfail") && params.name === "settings.read") {
    return { isError: true, content: [{ type: "text", text: "Settings fixture unavailable" }] };
  }
  return proxy("tools/call", params);
};
bridge.onreadresource = async params => {
  events.push({ method: "resources/read", params });
  if (params.uri === "host-resource://fixture") {
    return { contents: [{ uri: params.uri, mimeType: "text/plain", text: "<img src=x onerror=alert(1)> fixture", _meta: { "openai/resource": { writable: false, etag: "v1" } } }] };
  }
  return proxy("resources/read", params);
};
bridge.enableOpenAIMessages();
bridge.onupdatemodelcontext = async params => {
  events.push({ method: "ui/update-model-context", params });
  const updateId = `update-${++sequence}`;
  context = { ...params, updateId };
  bridge.setHostContext({ "openai/modelContext": context });
  if (parameters.has("remove-during-update")) {
    context = null;
    bridge.setHostContext({ "openai/modelContext": null });
  }
  return { _meta: { "openai/modelContext": { updateId } } };
};
bridge.onrequestdisplaymode = async params => {
  events.push({ method: "ui/request-display-mode", params });
  if (parameters.has("keepinline")) return { mode: "inline" };
  bridge.setHostContext({ displayMode: params.mode });
  return { mode: params.mode };
};
bridge.oninitialized = async () => {
  events.push({ method: "ui/notifications/initialized" });
  await bridge.sendToolInput({ arguments: parameters.has("initialfile") ? { file: { name: "fixture.txt", resourceUri: "host-resource://fixture" } } : {} });
  const result = await proxy("tools/call", { name: "open_workspace", arguments: {} });
  await bridge.sendToolResult(result);
};
Object.assign(window, { fixture: {
  events,
  context: (ids: string[] | null) => {
    context = ids === null ? null : { updateId: `host-${++sequence}`, content: (context?.content ?? []).filter(block => ids.includes(String(block._meta?.["example/selectionId"]))) };
    bridge.setHostContext({ "openai/modelContext": context });
  },
  deepLink: (url: string) => bridge.setHostContext({ "openai/deepLink": { url } }),
  file: () => bridge.sendToolInput({ arguments: { file: { name: "fixture.txt", resourceUri: "host-resource://fixture" } } }),
} });
// Connect before loading the document so ui/initialize cannot race the listener.
await bridge.connect(new PostMessageTransport(iframe.contentWindow!, iframe.contentWindow!));
iframe.src = `/app${parameters.has("webmcp") ? "?webmcp" : ""}`;
