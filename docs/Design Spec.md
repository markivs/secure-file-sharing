# Design Specification: Secure Distributed File Sharing Daemon

## Purpose
A fully client-side, distributed, encrypted file-sharing application designed as a security exercise. Files are securely stored in encrypted shards on Google Docs, emphasizing real-time collaboration, security through cryptography, and robust integrity verification via Merkle trees.

---

## Goals and Requirements

### Functional Requirements
- **Real-Time Collaboration:** 
  - Support for frequent, small, concurrent edits similar to Google Docs.
  - Operational Transformation (OT) for conflict resolution.

- **Encrypted Data Storage:** 
  - End-to-end encryption using AES-256-GCM.
  - Each shard is independently encrypted; observable metadata headers remain plaintext.

- **Integrity Verification:** 
  - Use global Merkle Trees spanning all shards.
  - Merkle root published publicly for cross-validation.

- **Dynamic Sharding:** 
  - Shards stored in individual Google Docs documents.
  - Load balanced via consistent hashing (dynamic shard discovery).
  - Minimal reallocation of data upon shard addition/removal.

- **Client-Side Logic Only:** 
  - Entirely client-driven; no server logic beyond Google Docs storage API.

- **Observable Metadata:** 
  - Headers, public keys, chunk hashes openly stored to ensure transparency and auditability.

---

### Non-Functional Requirements
- **Security & Privacy:**
  - Strong confidentiality (end-to-end encrypted payloads).
  - Digital signatures (ECDSA or Ed25519) to authenticate observable metadata.
  - Robust key rotation and revocation procedures.

- **Performance:**
  - Low latency real-time sync for small incremental edits.
  - Lightweight and efficient client daemon operation.

- **Scalability:**
  - Distributed design ensures horizontal scalability.
  - Minimal overhead during shard reallocation events.

- **Resilience:**
  - Tamper-evident structures (Merkle tree verification).
  - Client-driven security measures ensure robustness against MITM attacks.

---

## Constraints & Assumptions
- No dedicated backend servers; Google Docs acts only as dumb storage.
- All clients maintain local cryptographic keys independently.
- Operational Transform preferred over CRDTs due to simplicity and familiarity.
- External user OAuth credentials injected into clients explicitly.

---

## Key Design Choices
- Encryption: AES-256-GCM for data, ECDSA for metadata signatures.
- Sharding: Consistent hashing with virtual nodes.
- Real-time sync: Operational Transformation (OT).
- Key Rotation: Monthly periodic rotation plus event-driven revocation/rotation.

---

## Next Steps
- Detailed client class implementation (`DriveClient`).
- Proof-of-concept with a minimal Google Docs shard interaction.
- Integration of Merkle tree integrity logic.
