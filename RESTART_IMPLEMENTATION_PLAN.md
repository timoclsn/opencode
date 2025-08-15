# `/restart` Command Implementation Plan

## Overview

Implement a `/restart` command that allows users to restart opencode to pick up new versions after updates (homebrew or global npm). The command should exit the current process and spawn a fresh opencode instance from PATH.

## Use Case

After updating opencode via `brew upgrade opencode` or `npm update -g opencode`, users can type `/restart` to pick up the new version without manually exiting and restarting.

## Architecture

The implementation involves two main components:

1. **TUI (Go)**: Add the restart command and handle it by exiting with a special exit code
2. **CLI wrapper (TypeScript)**: Detect the special exit code and re-exec opencode from PATH

## Implementation Details

### 1. TUI Command (Go)

#### File: `packages/tui/internal/commands/command.go`

Add the new command constant and default entry:

```go
const (
    // ... existing commands
    AppRestartCommand           CommandName = "app_restart"
)

// In the LoadFromConfig function, add to defaults slice:
{
    Name:        AppRestartCommand,
    Description: "restart opencode",
    Trigger:     []string{"restart"},
    // No keybinding - only available via /restart trigger
},
```

#### File: `packages/tui/internal/tui/tui.go`

Add restart handling in the `executeCommand` function:

```go
case commands.AppRestartCommand:
    // Exit with special code to signal restart
    os.Exit(222)
```

#### File: `packages/tui/internal/components/chat/editor.go`

In the `Submit()` function, add restart handling alongside existing quit commands:

```go
switch value {
case "exit", "quit", "q", ":q":
    return m, tea.Quit
case "restart":
    return m, util.CmdHandler(commands.ExecuteCommandMsg(m.app.Commands[commands.AppRestartCommand]))
}
```

### 2. CLI Glue (TypeScript)

#### File: `packages/opencode/src/cli/cmd/tui.ts`

Define the restart exit code constant:

```typescript
const RESTART_EXIT_CODE = 222
```

Modify the TUI handler to detect restart and re-exec:

```typescript
// After spawning TUI process
const code = await proc.exited
server.stop()

if (code === RESTART_EXIT_CODE) {
  if (Installation.isDev() || Installation.isSnapshot()) {
    // In dev/snapshot, just continue the loop to relaunch
    continue
  } else {
    // Production: re-exec from PATH to pick up updates
    const cleanEnv = { ...process.env }
    delete cleanEnv.OPENCODE_BIN_PATH // Let wrapper resolve fresh binary

    const restartArgs = [
      "tui",
      ...(args.project ? [args.project] : []),
      ...(args.model ? ["--model", args.model] : []),
      ...(args.prompt ? ["--prompt", args.prompt] : []),
      ...(args.mode ? ["--mode", args.mode] : []),
      ...(args.session ? ["--session", args.session] : []),
      ...(args.continue ? ["--continue"] : []),
      ...(args.port !== 0 ? ["--port", args.port.toString()] : []),
      ...(args.hostname !== "127.0.0.1" ? ["--hostname", args.hostname] : []),
    ]

    Bun.spawn({
      cmd: ["opencode", ...restartArgs],
      cwd: process.cwd(),
      stdout: "inherit",
      stderr: "inherit",
      stdin: "inherit",
      env: cleanEnv,
    })

    // Exit current process - let new one take over
    process.exit(0)
  }
}

// Otherwise continue with existing logic (break if done)
```

## User Experience

- `/restart` will automatically appear in command completions when typing `/`
- `/restart` will be listed in the help dialog
- Users can type either `/restart` or just `restart` followed by Enter
- The command works the same way as `/exit` but restarts instead of quitting
- Existing update toast already says "restart to apply" - no changes needed

## Technical Notes

### Exit Code Choice

- Using exit code `222` to signal restart (arbitrary choice, unlikely to conflict)
- Both Go and TypeScript need to use the same code

### Cross-Platform Considerations

- Removing `OPENCODE_BIN_PATH` ensures npm/brew wrappers recalculate to updated binary
- Spawning `opencode` from PATH works on all platforms (Windows PATHEXT handles .cmd)
- stdio inheritance preserves terminal behavior

### Development vs Production

- **Development**: No re-exec attempted, just continues the loop to relaunch TUI/server
- **Production**: Re-execs `opencode` from PATH to pick up package manager updates
- **Snapshot builds**: Treated like development (no re-exec)

## Files to Modify

### Go Files

- `packages/tui/internal/commands/command.go` - Add restart command definition
- `packages/tui/internal/tui/tui.go` - Handle restart command execution
- `packages/tui/internal/components/chat/editor.go` - Handle plain "restart" input

### TypeScript Files

- `packages/opencode/src/cli/cmd/tui.ts` - Detect restart exit code and re-exec

## Validation Steps

1. **Basic functionality**: Run `opencode tui`, type `/restart`, confirm process restarts
2. **Command discovery**: Verify `/restart` appears in completions and help
3. **Plain text**: Verify typing just `restart` + Enter works
4. **Update scenario**: Update opencode, run `/restart`, confirm new version loads
5. **Cross-platform**: Test on macOS, Linux, Windows terminals
6. **IDE integration**: Test inside VS Code terminal
7. **Development mode**: Test that dev builds continue loop instead of re-exec

## Edge Cases

- If re-exec spawn fails (rare), current process will have already stopped - user can relaunch manually
- Network/permission issues during re-exec should be handled gracefully
- Preserve all original CLI arguments when restarting (project path, model, session, etc.)
