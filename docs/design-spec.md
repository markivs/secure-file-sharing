# Design Specification: Secure Distributed File Sharing Daemon

## Purpose
A fully client-side, distributed, encrypted file-sharing application designed as a security and distributed systems exercise. Files are securely stored in encrypted shards on Google Docs, emphasizing real-time collaboration, robust encryption (MITM-resistant), and transparent distributed algorithms and concepts.

---

## Goals and Requirements

### Functional Requirements
- **Real-Time Collaboration:** 
  - Frequent, small, concurrent edits via Operational Transformation (OT).

- **Encrypted Data Storage:** 
  - End-to-end AES-256-GCM encryption per shard.
  - Transparent metadata headers remain observable.

- **Integrity Verification:** 
  - Global Merkle trees across all shards.
  - Publicly observable Merkle roots for integrity validation.

- **Dynamic Sharding:** 
  - Shards stored independently in Google Docs files.
  - Load balancing via consistent hashing with dynamic shard discovery.

- **Client-Side Logic Only:** 
  - No additional servers or managed infrastructure beyond Google Docs storage.

- **Observable Metadata:** 
  - Open headers explicitly contain public keys, chunk hashes, Merkle roots, and signatures.

### **Distributed Event Queue (New)**
- Shards explicitly serve as lightweight distributed event logs.
- Clients publish encrypted events and subscribe to shard payload data.
- Explicit offset tracking and event processing confirmation within metadata.

---

### Non-Functional Requirements
- **Security & Privacy:** Strong confidentiality (AES-256-GCM), digital signatures on observable metadata, clear key rotation and revocation.
- **Performance:** Latency minimized (polling ~1–5 sec), efficient incremental updates.
- **Scalability:** Horizontal scalability via shard partitioning, minimal data movement during reallocation.
- **Resilience:** Robust MITM resistance, Merkle-based tamper evidence, conflict resolution via OT.

---

## Constraints & Assumptions
- No dedicated backend servers.
- Client-driven cryptographic key management explicitly.
- Operational Transform chosen for simplicity.
- OAuth credentials externally injected into clients.

---

## Key Design Choices
- Encryption: AES-256-GCM payload, ECDSA metadata signatures.
- Sharding: Consistent hashing.
- Real-time sync: OT-based.
- Key rotation: Monthly/event-driven hybrid rotation schedule.
- Distributed event log explicitly via encrypted shard payloads.

---

## Next Steps
- Detailed client implementation (`DriveClient`).
- Proof-of-concept event log within Google Docs payload.
- Merkle tree integrity and incremental synchronization logic implementation.
