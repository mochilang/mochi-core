# mochi-core

Core of the Mochi language: parser, type checker, and runtime v3.

| Package | Description |
|---------|-------------|
| `parser/` | Lexer and parser (participle-based) |
| `ast/` | AST node types, printer, and converter |
| `types/` | Type checker and type inference |
| `types/plan/` | Typed query plan nodes |
| `diagnostic/` | Position-aware compiler diagnostics |
| `runtime/mod/` | Module root discovery |
| `golden/` | Golden-file test helpers |
| `runtime3/rust/` | Mochi runtime v3 (Rust) |

## Build

```
go build ./...
go test ./...
```

## Rust runtime

```
cd runtime3/rust
cargo build
cargo test
```
