# System Architecture: Secure Distributed File Sharing Daemon

## Overview
A client-side, fully distributed architecture that securely manages encrypted file shards stored in Google Docs, coordinated without a centralized server.

---

## System Components

### 1. Client Daemon Application
A long-running Go-based daemon, entirely responsible for system logic:

- **Key Management**
  - Local secure generation, storage, and rotation of cryptographic keys.
  - Client-managed asymmetric key pairs (ECDSA or Ed25519).

- **Shard Coordination (Consistent Hashing)**
  - Dynamically discovers shard lists stored on Google Drive.
  - Determines shard assignment via consistent hashing.

- **Encryption & Decryption**
  - AES-256-GCM payload encryption/decryption.
  - Encrypted payload chunks stored in shards.

- **Metadata Handling**
  - Observable headers (plaintext), digitally signed for tamper resistance.
  - Public keys, Merkle roots, and chunk hashes publicly observable.

- **Integrity Verification (Merkle Tree Management)**
  - Global Merkle tree computed locally by each client.
  - Merkle root hashes published in shard metadata for verification.

- **Conflict Resolution (Operational Transformation)**
  - Ensures real-time synchronization and conflict-free concurrent edits.
  - Client-side Operational Transformation algorithm resolves concurrent edits.

### 2. Google Docs Storage (Dumb Storage Layer)
Acts purely as storage backend:

- Stores encrypted payload chunks (AES-256-GCM ciphertext).
- Stores observable metadata headers (plaintext, digitally signed).
- Triggers real-time notifications to subscribed clients upon edits.

---

## Data Structures

### Shard Structure (per Google Doc)
```plaintext
==== SHARD METADATA HEADER ====
Shard ID:                shard-UUID
Shard Version:           1
Merkle Root:             sha256(root)
Protocol Version:        1.0
Timestamp (latest edit): 2025-07-20T18:04:23Z

Participant Public Keys:
- ClientA: <pubkey_base64>
- ClientB: <pubkey_base64>

Shard Digital Signature:
- Signed by: ClientA
- Signature: <base64_ecdsa_sig>

==== ACTIVE SESSION INFO ====
Active Clients:
- ClientA (Last seen: timestamp)
- ClientB (Last seen: timestamp)

Key Rotation Events:
- Last Rotation: timestamp
- Next Rotation: timestamp

==== PAYLOAD DATA ====
Chunk ID | Chunk Hash (SHA-256) | AES-256-GCM Encrypted Payload (base64)
------------------------------------------------------------------------
1        | e9a3b5c...           | ciphertext_base64
2        | 2c7d4f1...           | ciphertext_base64
