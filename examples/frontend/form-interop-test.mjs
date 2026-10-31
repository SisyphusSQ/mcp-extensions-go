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

for (const [name, content, valid] of [
  ["ordinary numbers", { count: 42, measurement: 0.125, exact: "42" }, true],
  ["JS normalizes integral decimal", { count: 1.0, measurement: 1e2, exact: "1.0" }, true],
  ["JS normalizes integral exponent", { count: 1e0, measurement: 1e-2, exact: "1" }, true],
  ["exact string pattern rejects exponent", { count: 1, measurement: 1e-2, exact: "1e0" }, false],
  ["large exact text", { count: 1, measurement: 1.25, exact: "9007199254740993" }, true],
  ["decimal exact text", { count: 1, measurement: 0.1, exact: "0.12345678901234567890123456789" }, true],
  ["integer string rejected", { count: "42", measurement: 1, exact: "42" }, false],
  ["number string rejected", { count: 42, measurement: "1.25", exact: "42" }, false],
  ["number cannot replace exact string", { count: 1, measurement: 1, exact: 42 }, false],
  ["fraction is not integer", { count: 1.5, measurement: 1, exact: "42" }, false],
]) {
  test(`Numeric MRTR: ${name}`, async t => {
    const client = new Client({ name: "numeric-interop", version: "1" }, {
      capabilities, versionNegotiation: { mode: { pin: "2026-07-28" } },
    });
    t.after(() => client.close());
    client.setRequestHandler("elicitation/create", request => {
      const schema = request.params._meta["openai/elicitation"].requestedSchema;
      assert.equal(schema.properties.count.type, "integer");
      assert.equal(schema.properties.measurement.type, "number");
      assert.equal(schema.properties.exact.type, "string");
      return { action: "accept", content };
    });
    await client.connect(new StdioClientTransport({ command: resolve("../../bin/form-mrtr"), args: [] }));
    const result = await client.callTool({ name: "numeric", arguments: {} });
    assert.equal(!result.isError, valid, JSON.stringify(result));
    if (valid) assert.deepEqual(result.structuredContent, content);
  });
}
