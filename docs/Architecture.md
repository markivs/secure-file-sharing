# System Architecture: Secure Distributed File Sharing Daemon

## Overview
Fully client-side distributed architecture enabling secure collaboration and distributed message queuing via encrypted shards stored explicitly in Google Docs.

---

## System Components

### 1. Client Daemon Application
Client-only Go daemon manages:

- **Key Management:** Locally stored keys, rotation, revocation explicitly client-managed.
- **Shard Coordination (Consistent Hashing):** Dynamic shard discovery, balancing, explicit shard reallocation.
- **Encryption/Decryption (AES-256-GCM):** Payload encryption explicitly per shard.
- **Metadata Management:** Transparent metadata headers explicitly observable and signed.
- **Integrity Verification (Merkle Trees):** Explicit global Merkle tree spanning shards, Merkle root verification.
- **Conflict Resolution (Operational Transformation):** Ensuring consistent concurrent edits explicitly.

### 2. Google Docs Storage Layer (Dumb Storage)
Pure storage backend, explicitly structured:

- Observable plaintext headers explicitly for transparency.
- AES-256-GCM encrypted payloads explicitly used for secure file data and distributed event logs.

---

## Data Structures

### Shard Document Structure (Event-log payload explicitly included)

```plaintext
==== SHARD METADATA HEADER ====
Shard ID:                shard-UUID
Shard Version:           1
Merkle Root:             sha256(root)
Protocol Version:        1.0
Timestamp (latest edit): timestamp

Participant Public Keys:
- ClientA: <pubkey_base64>
- ClientB: <pubkey_base64>

Shard Digital Signature:
- Signed by: ClientA
- Signature: <base64_signature>

==== ACTIVE SESSION INFO ====
Active Clients:
- ClientA (Last seen: timestamp)
- ClientB (Last seen: timestamp)

Processing Offsets (New):
- ClientA: last_processed_chunk=5
- ClientB: last_processed_chunk=4

==== PAYLOAD DATA (Encrypted Event Log) ====
Chunk ID | Chunk Hash (SHA-256) | AES-256-GCM Encrypted Event Batch
----------------------------------------------------------------------
1        | hash                  | ciphertext_base64 (events 1–10)
2        | hash                  | ciphertext_base64 (events 11–20)
...
