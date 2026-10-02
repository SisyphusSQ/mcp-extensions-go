"""Generate fixtures using the pinned official OpenAI extension source.

Run with PYTHONPATH pointing to upstream python/src and dependencies matching
forms/testdata/python-parity.json, inside an isolated development environment.
"""

import hashlib
import inspect
import json
from collections.abc import Sequence
from importlib.metadata import version
from pathlib import Path

import openai_mcp_form_protocol
from openai_mcp_form_protocol import (
    FormField, FormSchema, complete_field_submission, is_valid_value,
    validate_file_selections,
)

COMMIT = "900032d8bd7c1566202d0cb1666986584f932043"
SOURCE_SHA256 = "e807a74ff4fa9470cc33ad883381ace0798cc507ef43688796f5560575c25646"
source = Path(inspect.getfile(openai_mcp_form_protocol)).read_bytes()
if hashlib.sha256(source).hexdigest() != SOURCE_SHA256:
    raise ValueError("PYTHONPATH must use the pinned OpenAI form protocol source")
cases: list[dict[str, object]] = []


def add(
    name: str,
    field: dict[str, object],
    values: Sequence[object] = (),
    uploads: Sequence[tuple[dict[str, object], Sequence[str]]] = (),
) -> None:
    schema = {"type": "object", "properties": {"value": field}, "required": ["value"]}
    row = {"name": name, "schema": schema, "valid": False, "values": [], "uploads": []}
    try:
        parsed = FormSchema[FormField].model_validate(schema)
    except ValueError:
        cases.append(row)
        return
    row["valid"] = True
    prop = parsed.properties["value"]
    for value in values:
        valid = is_valid_value(prop, value)
        if prop.file_input is not None:
            try:
                validate_file_selections(parsed, {"value": value})
                valid = True
            except ValueError:
                valid = False
        row["values"].append({"value": value, "valid": valid})
    for content, uploaded in uploads:
        try:
            result = complete_field_submission(prop, "value", content, tuple(uploaded))
            row["uploads"].append({"content": content, "uploaded": uploaded, "valid": True, "result": result})
        except ValueError:
            row["uploads"].append({"content": content, "uploaded": uploaded, "valid": False})
    cases.append(row)


add("integer enum", {"type": "integer", "enum": [1, 2]}, [1, 2, 3, True, 1.0, "1"])
add("number enum", {"type": "number", "enum": [1, 2.5]}, [1, 1.0, 2.5, 3, True])
add("boolean enum", {"type": "boolean", "enum": [False]}, [False, True, 0])
add("array enum", {"type": "array", "items": {"type": "string"}, "enum": [["a"], []]}, [["a"], [], ["b"]])
add("mixed invalid enum", {"type": "integer", "enum": [1, True]})
add("empty enum", {"type": "string", "enum": []})
add("duplicate top enum", {"type": "string", "enum": ["a", "a"]}, ["a", "b"])
add("blank labels", {"type": "string", "enum": ["a"], "enumNames": [""]}, ["a", "b"])
add("blank rich title", {"type": "string", "oneOf": [{"const": "a", "title": ""}]}, ["a", "b"])
add("choices and suggestions", {"type": "string", "enum": ["a"], "oneOf": [{"const": "a", "title": "A"}], "x-openai-suggestions": [{"const": "b", "title": "B"}]}, ["a", "b"])
add("suggestions are hints", {"type": "string", "minLength": 2, "x-openai-suggestions": [{"const": "a", "title": "A"}]}, ["custom", "a"])
add("annotations", {"type": "string", "examples": [None, 1, {"nested": [None]}], "$comment": "note", "_meta": {"nullable": None}, "deprecated": False, "readOnly": True, "writeOnly": False}, ["", "text"])
add("item annotations forbidden", {"type": "array", "items": {"type": "string", "readOnly": True}})
add("choice item annotations forbidden", {"type": "array", "items": {"type": "string", "enum": ["a"], "title": "A"}})
add("free arrays", {"type": "array", "items": {"type": "string", "minLength": 2, "x-openai-suggestions": [{"const": "x", "title": ""}]}, "uniqueItems": True}, [["ok", "custom"], ["ok", "ok"], ["x"], [1]])
add("multi-select", {"type": "array", "items": {"enum": ["a", "b"]}}, [["a"], ["a", "a"], ["other"], []])
add("conflicting multi choices", {"type": "array", "items": {"enum": ["a"], "anyOf": [{"const": "a", "title": "A"}]}})
add("duplicate multi choices", {"type": "array", "items": {"enum": ["a", "a"]}})
add("exclusive form bound forbidden", {"type": "number", "exclusiveMinimum": 0})
add("multipleOf form forbidden", {"type": "integer", "multipleOf": 2})
add("numeric bounds", {"type": "integer", "minimum": 1, "maximum": 3}, [0, 1, 3, 4, 1.0, "2"])
add("nullable annotations", {"type": "string", "title": None, "description": None, "examples": None, "_meta": None, "deprecated": None}, ["text"])
add("null default forbidden", {"type": "string", "default": None})
add("null pattern forbidden", {"type": "string", "pattern": None})
add("nullable bound", {"type": "string", "minLength": None}, ["text"])
add("wrong constraint type", {"type": "boolean", "minLength": None})
add("image grammar", {"type": "string", "oneOf": [{"const": "a", "title": "A", "x-openai-thumbnail": {"src": "data:image/png;base64,Y"}}]}, ["a"])
add("HTTP image forbidden", {"type": "string", "oneOf": [{"const": "a", "title": "A", "x-openai-thumbnail": {"src": "http://example.com/a"}}]})
add("date", {"type": "string", "format": "date"}, ["2024-02-29", "0000-01-01", "2025-02-29", "2026-1-01"])
add("date-time", {"type": "string", "format": "date-time"}, ["2026-10-02t01:02:03.1z", "0000-01-01T00:00:00Z", "2026-10-02T01:02:03+23:59", "2026-10-02T01:02:03+24:00"])
add("email", {"type": "string", "format": "email"}, ["a@example.com", "a@localhost", "a@127.0.0.1", "A <a@example.com>", "a@-example.com", "a@example.com\n", "first.last+tag@example.com", '"a"@example.com', "用户@例子.公司"])
add("URI", {"type": "string", "format": "uri"}, ["urn:x:part", "file:///a", "https://[::1]/a", "relative", "urn:x:%xy", "https://example.com\n"])
resource = {"type": "resource", "options": [{"uri": "parts://a", "name": "", "extra": {"nullable": None}, "_meta": {"null": None}}]}
add("resource descriptors", {"type": "string", "format": "uri", "x-openai-input": resource}, ["parts://a", "host://file"])
user = {**resource, "userOptions": {"accept": [".txt", "image/*"]}}
add("host selections and uploads", {"type": "array", "items": {"type": "string", "format": "uri"}, "minItems": 1, "maxItems": 2, "uniqueItems": True, "x-openai-input": user}, [["parts://a"], ["host://file"], ["relative"], ["parts://a", "parts://a"]], [({}, ["host://file"]), ({"value": ["parts://a"]}, ["host://file"]), ({}, ["host://file", "host://file"]), ({"value": ["parts://a"]}, ["host://file", "host://second"])])
add("defaults exclude host selections", {"type": "string", "format": "uri", "default": "host://file", "x-openai-input": user})
add("implicit selection", {"type": "array", "items": {"type": "string", "format": "uri"}, "x-openai-input": {**resource, "selection": "implicit"}}, [["host://file"]])
add("implicit default forbidden", {"type": "array", "items": {"type": "string", "format": "uri"}, "default": [], "x-openai-input": {**resource, "selection": "implicit"}})
add("opaque accept extension", {"type": "string", "format": "uri", "x-openai-input": {**resource, "userOptions": {"accept": ["."]}}})
add("duplicate accept", {"type": "string", "format": "uri", "x-openai-input": {**resource, "userOptions": {"accept": [".TXT", ".txt"]}}})
add("whitespace accept forbidden", {"type": "string", "format": "uri", "x-openai-input": {**resource, "userOptions": {"accept": [" text/plain"]}}})
add("empty suggestions", {"type": "string", "x-openai-suggestions": []}, ["custom"])
add("duplicate suggestions", {"type": "string", "x-openai-suggestions": [{"const": "a", "title": "A"}, {"const": "a", "title": "A2"}]}, ["a", "custom"])
add("null multi type", {"type": "array", "items": {"type": None, "enum": ["a"]}}, [["a"]])
add("exact keyword casing", {"type": "string", "READONLY": True})
add("empty format forbidden", {"type": "string", "format": ""})
add("empty selection forbidden", {"type": "array", "items": {"type": "string", "format": "uri"}, "x-openai-input": {**resource, "selection": ""}})
add("empty user kind forbidden", {"type": "string", "format": "uri", "x-openai-input": {**resource, "userOptions": {"kind": ""}}})
add("enum and rich choice intersection", {"type": "string", "enum": ["a", "b"], "oneOf": [{"const": "a", "title": "A"}]})
add("precise integer enum", {"type": "integer", "enum": [9007199254740993]}, [9007199254740993, 9007199254740992])
for name, extra in [("duplicate required names", {"required": ["value", "value"]}), ("custom schema annotation", {"$schema": "urn:example:form-schema"})]:
    schema = {"type": "object", "properties": {"value": {"type": "string"}}, **extra}
    parsed = FormSchema[FormField].model_validate(schema)
    cases.append({"name": name, "schema": schema, "valid": True, "values": [], "uploads": []})
fixture = {"upstream_commit": COMMIT, "upstream_source_sha256": SOURCE_SHA256, "dependencies": {name: version(name) for name in ("pydantic", "email-validator", "rfc3986-validator", "mcp-types")}, "cases": cases}
Path("forms/testdata").mkdir(exist_ok=True)
Path("forms/testdata/python-parity.json").write_text(json.dumps(fixture, indent=2, ensure_ascii=False) + "\n")
print(f"Generated {len(cases)} schema cases and {sum(len(row['values']) + len(row['uploads']) for row in cases)} value/submission cases")
