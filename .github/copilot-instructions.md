# Plenc - Copilot Instructions

## Overview

Plenc is a Go serialization library based on protobuf wire format, but using Go struct tags instead of `.proto` files. It prioritizes performance over protobuf compatibility.

## Architecture

### Package Structure
- **`plenc`** (root): Public API (`Marshal`, `Unmarshal`), codec registry, default `Plenc` instance
- **`plenccodec`**: Core codec implementations (`Codec` interface, type-specific codecs)
- **`plenccore`**: Wire protocol (varints, wire types, tag encoding)
- **`null`**: Optional codecs for `github.com/unravelin/null` types
- **`cmd/plenctag`**: CLI tool to auto-add plenc tags to structs

### Key Concepts

**Codec Interface** ([plenccodec/codec.go](plenccodec/codec.go)): All types are encoded via codecs implementing:
```go
type Codec interface {
    Omit(ptr unsafe.Pointer) bool           // Skip zero values
    Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error)
    Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte
    Size(ptr unsafe.Pointer, tag []byte) int
    WireType() plenccore.WireType
    // ...
}
```

**Wire Types** ([plenccore/wire.go](plenccore/wire.go)): `WTVarInt`, `WT64`, `WT32`, `WTLength`, `WTSlice` (plenc-specific for efficient slice handling)

**Struct Codec Generation**: Codecs for structs are auto-generated at first use via reflection ([plenccodec/struct.go](plenccodec/struct.go))

## Struct Tag Conventions

```go
type Example struct {
    A int     `plenc:"1"`           // Required: unique index (never reuse after field removal)
    B string  `plenc:"-"`           // Excluded from encoding
    C string  `plenc:"2,intern"`    // String interning for repeated values
    D int     `plenc:"3,flat"`      // Non-zigzag encoding (for always-positive ints)
}
```

**Rules:**
- All exported fields MUST have a `plenc` tag (error if missing)
- Index numbers must be unique within a struct
- Never reuse index numbers from removed fields
- Field names can change; types cannot

## Development Workflow

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./plenccodec/...

# Auto-add plenc tags to structs
go run ./cmd/plenctag -w myfile.go
```

## Adding New Codecs

1. Implement `plenccodec.Codec` interface
2. Register with `plenc.RegisterCodec(reflect.TypeOf(MyType{}), myCodec{})`
3. For tag-specific variants: `plenc.RegisterCodecWithTag(typ, "mytag", codec)`

Example pattern from [plenccodec/int.go](plenccodec/int.go):
```go
type IntCodec[T int | int8 | int16 | int32 | int64] struct{}
// Implement all Codec methods using unsafe.Pointer for zero-allocation access
```

## Performance Patterns

- **unsafe.Pointer**: Used extensively for zero-allocation field access
- **Pre-computed tags**: Struct fields cache their encoded tag bytes
- **Size-then-append**: `Size()` called first to pre-allocate exact buffer size
- **Interning**: Use `intern` tag for strings with limited distinct values

## Testing Patterns

- Fuzz testing with `github.com/google/gofuzz` for round-trip verification
- Golden files in [plenccodec/testdata/](plenccodec/testdata/) for encoding stability
- Use `github.com/google/go-cmp/cmp` for struct comparisons

## Protobuf Compatibility

Plenc diverges from protobuf in slice encoding (`WTSlice`). For proto compatibility:
```go
var p plenc.Plenc
p.ProtoCompatibleArrays = true  // Use standard protobuf repeated field encoding
p.ProtoCompatibleTime = true    // Use google.protobuf.Timestamp format
p.RegisterDefaultCodecs()
```

## Common Gotchas

- Empty slices unmarshal as `nil`, not `[]T{}`
- Slices of pointers with nil elements are not fully supported
- Maps are pointer-like types; don't pass `*map[K]V` to Marshal
- Recursive types are supported (via wrapped registry during codec building)
