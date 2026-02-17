# Features

## Completed

| Feature | Package | Description |
|---------|---------|-------------|
| Lint | `bufcheck` | Configurable protobuf linting with 80+ built-in rules |
| Breaking | `bufcheck` | Breaking change detection between proto versions |
| Generate | `bufgen` | Code generation orchestration with plugin management |
| Format | `bufformat` | Protobuf file formatter |
| Build | `bufimage` | Proto compilation to FileDescriptorSet images |
| Export | `bufctl` | Export resolved proto files |
| Convert | `bufconvert` | Convert between proto serialization formats |
| LSP | `buflsp` | Language Server Protocol for IDE integration |
| Workspaces | `bufworkspace` | Multi-module workspace support |
| BSR Push/Pull | `bufmodule` | Push/pull modules to Buf Schema Registry |
| Dependencies | `bufmodule` | Dependency resolution with lock files |
| Custom Plugins | `bufplugin` | User-defined lint and breaking change plugins |
| Policy | `bufpolicy` | Policy management and enforcement |
| Remote Plugins | `bufremoteplugin` | Docker-based remote plugin execution |
| WASM Plugins | `wasm` | WebAssembly plugin execution |
| curl | `bufcurl` | gRPC/Connect RPC invocation |
| Studio Agent | `bufstudioagent` | BSR Studio integration proxy |
| Offline Mode | `bufcli` | Cache-backed operation without network |

## Proposed

| Feature | Priority | Description |
|---------|----------|-------------|
| Standalone package extraction | P3 | Extract `storage`, `protovalidate`, `protoyaml` as independent modules |
| Improved offline errors | P3 | Better error messages when cache misses occur in offline mode |
| Flag-based config | P3 | Replace config file-based options with CLI flags (noted in bufcobra TODOs) |
