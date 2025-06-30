import { tool } from "ai"
import { MCP } from "../mcp"
import { z } from "zod"

export const mcpList = tool({
  description:
    "List all MCP servers and their current status (enabled/disabled)",
  parameters: z.object({}),
  execute: async () => {
    const servers = await MCP.list()
    return {
      servers,
      description:
        "Current MCP server status. Use mcpToggle tool to enable/disable servers at runtime.",
    }
  },
})

export const mcpToggle = tool({
  description: "Enable or disable an MCP server at runtime",
  parameters: z.object({
    name: z.string().describe("Name of the MCP server to toggle"),
    enabled: z
      .boolean()
      .describe("Whether to enable (true) or disable (false) the server"),
  }),
  execute: async ({ name, enabled }) => {
    const result = await MCP.toggle(name, enabled)
    return {
      ...result,
      message: `MCP server '${name}' has been ${enabled ? "enabled" : "disabled"}`,
    }
  },
})
