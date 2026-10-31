// Match the optional API used by the pinned official WebMCP example.
interface Document {
  readonly modelContext?: {
    registerTool(tool: {
      name: string;
      description: string;
      inputSchema: Record<string, unknown>;
      annotations?: { readOnlyHint?: boolean };
      execute(input: unknown): Promise<string>;
    }): void;
  };
}
