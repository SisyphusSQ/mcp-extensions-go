import assert from "node:assert/strict";
import { resolve } from "node:path";
import { test } from "node:test";
import { Client } from "@modelcontextprotocol/client";
import { StdioClientTransport } from "@modelcontextprotocol/client/stdio";

// This is SDK interoperability evidence, not a real OpenAI form renderer.
const capabilities = {
  elicitation: { form: {} },
  extensions: { "openai/elicitation": { form: {} } },
};

async function choose(t, answer, options = {}) {
  const advertisedCapabilities = options.capabilities ?? capabilities;
  const client = new Client({ name: "typescript-form-interop", version: "1" }, {
    capabilities: advertisedCapabilities,
    versionNegotiation: { mode: options.legacy ? "legacy" : { pin: "2026-07-28" } },
  });
  let requests = 0;
  if (advertisedCapabilities.elicitation) client.setRequestHandler("elicitation/create", request => {
    requests++;
    assert.deepEqual(request.params.requestedSchema, { type: "object", properties: {} });
    const schema = request.params._meta["openai/elicitation"].requestedSchema;
    assert.equal(schema.properties.code.pattern, "^[A-Z]{3}$");
    assert.deepEqual(schema.properties.code["x-openai-suggestions"], [{ const: "ABC", title: "Example" }]);
    assert.equal(schema.properties.images["x-openai-input"].options[0].uri, "demo://image");
    return answer;
  });
  t.after(() => client.close());
  await client.connect(new StdioClientTransport({ command: resolve("../../bin/form-mrtr"), args: [] }));
  const result = await client.callTool({ name: "choose", arguments: {} });
  return { result, requests };
}

for (const action of ["accept", "cancel", "decline"]) {
  test(`Go server / TypeScript client: ${action}`, async t => {
    const content = { code: "ABC", images: ["demo://image"] };
    const { result, requests } = await choose(t, { action, ...(action === "accept" ? { content } : {}) });
    assert.ok(!result.isError, JSON.stringify(result));
    assert.equal(requests, 1);
    assert.deepEqual(result.structuredContent, action === "accept" ? content : { action });
  });
}

for (const content of [{ code: "bad", images: ["demo://image"] }, { code: "ABC", images: ["demo://unlisted"] }, { code: "ABC" }]) {
  test(`Go server rejects ${JSON.stringify(content)}`, async t => {
    const { result, requests } = await choose(t, { action: "accept", content });
    assert.equal(result.isError, true);
    assert.equal(requests, 1);
  });
}

for (const options of [{ capabilities: { elicitation: { form: {} } } }, { capabilities: { extensions: capabilities.extensions } }, { legacy: true }]) {
  test(`Unsupported client never receives a form: ${JSON.stringify(options)}`, async t => {
    const { result, requests } = await choose(t, { action: "cancel" }, options);
    assert.equal(result.isError, true);
    assert.equal(requests, 0);
  });
}
