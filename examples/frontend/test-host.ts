// A local protocol fixture using the official AppBridge, not an OpenAI host.
import { AppBridge, PostMessageTransport } from "@modelcontextprotocol/ext-apps/app-bridge";

const iframe = document.getElementById("app") as HTMLIFrameElement;
const events: Array<{ method: string; params?: unknown }> = [];
let sequence = 0;
const disabled = new URLSearchParams(window.location.search).has("noextensions");
const parameters = new URLSearchParams(window.location.search);
const initialMode = parameters.has("fullscreen") ? "fullscreen" : "inline";
const bridge = new AppBridge(null, { name: "Local protocol fixture", version: "0" }, {
  serverTools: {}, serverResources: {}, message: {}, updateModelContext: { text: {} },
  experimental: disabled ? {} : { "openai/modelContext": {}, "openai/message": {}, "openai/resource": {} },
}, { hostContext: { displayMode: initialMode, availableDisplayModes: parameters.has("nofullscreen") ? ["inline"] : ["inline", "fullscreen"], "openai/deepLink": { url: "/parts?tag=bolt" } } });

async function proxy(method: "tools/call" | "resources/read", params: unknown) {
  const response = await fetch("/rpc", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ method, params }) });
  if (!response.ok) throw new Error("Local MCP proxy failed");
  return response.json();
}

bridge.oncalltool = async params => { events.push({ method: "tools/call", params }); return proxy("tools/call", params); };
bridge.onreadresource = async params => {
  events.push({ method: "resources/read", params });
  if (params.uri === "host-resource://fixture") {
    return { contents: [{ uri: params.uri, mimeType: "text/plain", text: "<img src=x onerror=alert(1)> fixture", _meta: { "openai/resource": { writable: false, etag: "v1" } } }] };
  }
  return proxy("resources/read", params);
};
bridge.onmessage = async params => { events.push({ method: "ui/message", params }); return {}; };
bridge.onupdatemodelcontext = async params => {
  events.push({ method: "ui/update-model-context", params });
  return { _meta: { "openai/modelContext": { updateId: `update-${++sequence}` } } };
};
bridge.onrequestdisplaymode = async params => {
  events.push({ method: "ui/request-display-mode", params });
  if (parameters.has("keepinline")) return { mode: "inline" };
  bridge.setHostContext({ displayMode: params.mode });
  return { mode: params.mode };
};
bridge.oninitialized = async () => {
  events.push({ method: "ui/notifications/initialized" });
  await bridge.sendToolInput({ arguments: {} });
  const result = await proxy("tools/call", { name: "open_workspace", arguments: {} });
  await bridge.sendToolResult(result);
};
Object.assign(window, { fixture: {
  events,
  deepLink: (url: string) => bridge.setHostContext({ "openai/deepLink": { url } }),
  file: () => bridge.sendToolInput({ arguments: { file: { name: "fixture.txt", resourceUri: "host-resource://fixture" } } }),
} });
// Connect before loading the document so ui/initialize cannot race the listener.
await bridge.connect(new PostMessageTransport(iframe.contentWindow!, iframe.contentWindow!));
iframe.src = "/app";
