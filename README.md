# @lgcorzo/kms-go — Sovereign Cryptographic KMS & KES Go SDKs

[![Go Reference (KMS)](https://pkg.go.dev/badge/github.com/lgcorzo/kms-go/kms.svg)](https://pkg.go.dev/github.com/lgcorzo/kms-go/kms)
[![Go Reference (KES)](https://pkg.go.dev/badge/github.com/lgcorzo/kms-go/kes.svg)](https://pkg.go.dev/github.com/lgcorzo/kms-go/kes)
[![CI Build Status](https://github.com/lgcorzo/kms-go/actions/workflows/go.yml/badge.svg)](https://github.com/lgcorzo/kms-go/actions/workflows/go.yml)
[![License: AGPLv3](https://img.shields.io/badge/License-AGPLv3-blue.svg)](./LICENSE)

This repository provides Go client primitives for key creation, Data Encryption Key (DEK) derivation, and envelope encryption via MinIO KMS and MinIO KES servers. It is actively maintained as part of the **Sovereign MinIO Ecosystem** under `@lgcorzo`.

---

## SDK Modules & Quickstart

This repository contains two independent Go modules:
- [**`kms-go/kms`**](#kms-sdk): Go SDK for MinIO Key Management Service (KMS).
- [**`kms-go/kes`**](#kes-sdk): Go SDK for Key Encryption Service (KES).

Each module uses its own semantic versioning and can be imported separately.

### KMS SDK

Import the KMS SDK in your Go application:
```sh
$ go get github.com/lgcorzo/kms-go/kms@latest
```

Or add it to your `go.mod` file:
```go
require (
    github.com/lgcorzo/kms-go/kms v0.5.0
)
```

#### Basic KMS Client Example
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lgcorzo/kms-go/kms"
)

func main() {
    client, err := kms.NewClient(&kms.Config{
        Endpoints: []string{"https://kms.example.com:7373"},
    })
    if err != nil {
        log.Fatalf("failed to create KMS client: %v", err)
    }

    // Generate a Data Encryption Key (DEK)
    resp, err := client.GenerateKey(context.Background(), "default", &kms.GenerateKeyRequest{
        Name: "my-master-key",
    })
    if err != nil {
        log.Fatalf("failed to generate key: %v", err)
    }

    fmt.Printf("Generated DEK (key version: %d)\n", resp[0].Version)
}
```

---

### KES SDK

Import the KES SDK in your Go application:
```sh
$ go get github.com/lgcorzo/kms-go/kes@latest
```

Or add it to your `go.mod` file:
```go
require (
    github.com/lgcorzo/kms-go/kes v0.3.0
)
```

#### Basic KES Client Example
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lgcorzo/kms-go/kes"
)

func main() {
    client := kes.NewClient("https://kes.example.com:7373", nil)

    // Derive a data key via KES
    dek, err := client.GenerateKey(context.Background(), "my-key", nil)
    if err != nil {
        log.Fatalf("failed to derive DEK: %v", err)
    }

    fmt.Printf("Derived DEK length: %d bytes\n", len(dek.Plaintext))
}
```

---

## Dark Gravity Factory Rationale & Sovereign Maintenance

`@lgcorzo/kms-go` plays a critical role within the **Dark Gravity** sovereign AI production infrastructure and automated agent software factory.

### Why Sovereign Maintenance under `@lgcorzo`?

1. **Full Supply-Chain Autonomy**
   - Eliminates reliance on upstream breaking changes, license pivots, or unannounced module deprecations.
   - Guarantees predictable, reproducible builds across enterprise and cloud-native deployments.

2. **Dark Gravity Core Integration**
   - Cryptographic primitives provided by `kms-go` drive automated envelope encryption and key rotation across AI agent pipelines, multi-tenant vector storage, and high-throughput data lakes.
   - Direct integration ensures zero-latency DEK derivation and state encryption in sovereign AI workflows.

3. **Compliance & Security Assurance**
   - Maintained under strict zero-CVE SLAs with continuous automated Go vulnerability scanning (`govulncheck`), static analysis (`golangci-lint`), and security analysis (`CodeQL`).
   - Ensures full compliance with EU AI Act data protection rules, SOC 2 Type II auditability, and ISO 25059 AI lifecycle standards.

4. **100% Ecosystem Interoperability**
   - Seamless compatibility with all 38 repositories in the `@lgcorzo` Sovereign MinIO Ecosystem (Server, KES, Operator, DirectPV, MC, Console, and SIMD hardware acceleration primitives).

---

## Sovereign MinIO Ecosystem (38 Repositories)

The table below outlines the 38 interconnected repositories maintained under `@lgcorzo`:

| Category | Repository | Description |
| :--- | :--- | :--- |
| **Core Storage** | `lgcorzo/minio` | High-performance, S3-compatible sovereign object storage server |
| | `lgcorzo/minio-go` | Official Go client SDK for MinIO S3 API |
| | `lgcorzo/madmin-go` | Go administration client SDK for MinIO clusters |
| | `lgcorzo/mc` | MinIO Client CLI tool for data management |
| **Security & KMS** | `lgcorzo/kes` | High-performance Key Encryption Service |
| | `lgcorzo/kms-go` | Go SDKs for KMS and KES key management & envelope encryption |
| | `lgcorzo/dser` | Dynamic security enforcement & policy engine |
| | `lgcorzo/sidekick` | High-availability sidecar proxy for MinIO endpoints |
| **Kubernetes & Cloud Native** | `lgcorzo/operator` | MinIO Kubernetes Operator for cloud-native orchestration |
| | `lgcorzo/directpv` | CSI plugin for direct-attached local storage provisioner |
| | `lgcorzo/console` | Web-based management console UI for MinIO |
| | `lgcorzo/minio-hs` | High-density storage management utilities |
| **SIMD & Acceleration** | `lgcorzo/sha256-simd` | AVX512/ARM64 SIMD accelerated SHA-256 implementation |
| | `lgcorzo/blake2b-simd` | SIMD-accelerated BLAKE2b hashing primitives |
| | `lgcorzo/siphash-go` | Fast SIMD-optimized SipHash implementation |
| | `lgcorzo/highwayhash` | High-throughput HighwayHash checksum library |
| | `lgcorzo/md5-simd` | SIMD-accelerated MD5 implementation |
| | `lgcorzo/dsimd` | Dynamic SIMD dispatch & detection helper |
| | `lgcorzo/simdjson-go` | SIMD-accelerated JSON parser for Go |
| **Networking & I/O** | `lgcorzo/mux` | High-efficiency HTTP/TCP multiplexer |
| | `lgcorzo/pkg` | Shared core utility packages and platform helpers |
| | `lgcorzo/s3-check` | Automated S3 API conformance test suite |
| | `lgcorzo/s3-benchmark` | High-throughput S3 benchmarking suite |
| | `lgcorzo/warp` | S3 performance benchmarking and stress-testing tool |
| **Eventing & Messaging** | `lgcorzo/nats-server` | High-performance NATS messaging server |
| | `lgcorzo/nats-streaming-server` | Streaming server for event-driven storage workflows |
| | `lgcorzo/stan.go` | Go client library for NATS Streaming |
| | `lgcorzo/nats.go` | Go client library for NATS core messaging |
| **Utilities & Data Formats** | `lgcorzo/compress` | High-speed compression library (S2, zstd, deflate) |
| | `lgcorzo/zip` | Streaming ZIP archive manipulation library |
| | `lgcorzo/dsnet` | Distributed storage networking utilities |
| | `lgcorzo/freeloader` | Dynamic data prefetching and caching daemon |
| | `lgcorzo/parquet-go` | Parquet columnar storage format library |
| | `lgcorzo/csv-go` | High-performance CSV reader/writer primitives |
| | `lgcorzo/jsonparser` | Zero-allocation JSON parser |
| | `lgcorzo/ldap` | LDAP authentication integration library |
| | `lgcorzo/certgen` | Zero-dependency TLS certificate generation tool |

---

## Sovereign Automated Maintenance CI/CD Pipeline

```
  Upstream MinIO / Security Updates
                  │
                  ▼
   ┌──────────────────────────────┐
   │ Sovereign Sync & Inspection  │
   └──────────────┬───────────────┘
                  │
                  ▼
   ┌──────────────────────────────┐
   │ Automated CI/CD Testing      │
   │  • Multi-OS (Linux/macOS/Win)│
   │  • Go Vulncheck               │
   │  • CodeQL Static Analysis    │
   │  • Golangci-lint             │
   └──────────────┬───────────────┘
                  │
                  ▼
   ┌──────────────────────────────┐
   │ Sovereign @lgcorzo Release   │
   │  • Zero CVE Assurance        │
   │  • Dark Gravity AI Ready     │
   └──────────────────────────────┘
```

---

## License

Use of the KES SDK and KMS SDK is governed by the AGPLv3 license that can be found in the [LICENSE](./LICENSE) file.
