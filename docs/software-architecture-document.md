# Software Architecture Document — y2psync

**Version:** 1.0  
**Date:** 2026-07-04  
**Status:** Draft

---

## Table of Contents

1. [Introduction and Goals](#1-introduction-and-goals)
2. [Constraints](#2-constraints)
3. [Context and Scope](#3-context-and-scope)
4. [Solution Strategy](#4-solution-strategy)
5. [Building Block View](#5-building-block-view)
6. [Runtime View](#6-runtime-view)
7. [Deployment View](#7-deployment-view)
8. [Crosscutting Concepts](#8-crosscutting-concepts)
9. [Architectural Decisions](#9-architectural-decisions)
10. [Quality Requirements](#10-quality-requirements)
11. [Risks and Technical Debt](#11-risks-and-technical-debt)
12. [Glossary](#12-glossary)

---

## 1. Introduction and Goals

### 1.1 Requirements Overview

y2psync is a cross-platform application that enables users to save, organise, and synchronise YouTube playlists and channel subscriptions across their personal devices without any centralised server, user account, or dependency on YouTube APIs.

**Core features:**
- Create and manage named YouTube playlist collections
- Create and manage named YouTube channel subscription collections
- Add YouTube videos to playlists by URL (manual entry on desktop, via Android share intent on mobile)
- Import complete YouTube playlists by URL (delta-merge, no duplicates, no overwrites)
- Add YouTube channels to subscription lists by URL (via share intent or manual entry)
- Decentralised peer-to-peer sync across user's devices using a passphrase-based Master Sync Key
- Full offline operation with sync when connectivity is available
- Timestamp-based conflict resolution per entry

### 1.2 Quality Goals

| Priority | Goal | Description |
|----------|------|-------------|
| 1 | **Privacy** | No central server, no account, no YouTube API calls. Peer identity cannot be de-anonymised. All data encrypted in transit. |
| 2 | **Offline First** | All operations succeed locally without network. Sync is opportunistic and non-blocking. |
| 3 | **Data Integrity** | No data loss on sync. Timestamp-based conflict resolution ensures convergence. No duplicate entries ever created. |
| 4 | **Cross-Platform** | Mobile (Android) and desktop (Linux/Windows/macOS) share identical data model and sync protocol. |
| 5 | **Ease of Use** | Sync setup is a single passphrase. No server configuration, no account registration, no API keys. |

### 1.3 Stakeholders

| Stakeholder | Expectations |
|-------------|--------------|
| End User (Privacy-Conscious) | No tracking, no account required, full control over data |
| End User (YouTube Power User) | Save and organise playlists/subscriptions across multiple devices |
| Developer/Implementer | Clear specification, well-defined data model and sync protocol |
| Future iOS Port Team | Architecture supports adding iOS client without protocol changes |

---

## 2. Constraints

### 2.1 Technical Constraints

| Constraint | Rationale |
|------------|-----------|
| No YouTube API usage (neither OAuth nor Data API v3) | User privacy; no dependency on Google; app works without internet to Google |
| No central server | User privacy; no operational cost; no single point of failure |
| No user account system | Privacy; no email, no password (beyond local passphrase), no identity provider |
| Must work fully offline | Users may have limited connectivity; sync is only needed occasionally |
| Android minimum API 26 | Reasonable modern baseline; covers ~95% of active Android devices |
| Desktop targets: Linux, Windows, macOS | Cross-platform Qt/Go/Tauri/etc. must support all three |

### 2.2 Organisational Constraints

| Constraint | Rationale |
|------------|-----------|
| Mobile app in Kotlin | Android ecosystem standard; interoperates with Android share intents |
| Desktop app language free choice | Performance and developer preference; must implement same data model and sync protocol |
| Open source preferred | Community trust for privacy-focused application |

### 2.3 Conventions

- All timestamps in UTC, nanosecond precision where available, millisecond minimum
- All identifiers: local UUID v4 for internal IDs, YouTube's native IDs (video, playlist, channel) for external references
- All network communication encrypted with AEAD
- RFC 2119 keywords (SHALL, SHOULD, MAY) used in requirement definitions

---

## 3. Context and Scope

### 3.1 System Context

The **y2psync** system interacts with:
- **Users** (via Desktop UI and Android Mobile UI)
- **YouTube** (public web pages only, no API — read-only scraping of publicly visible metadata)
- **Other y2psync peer instances** (via P2P network)
- **Android Share Intent System** (mobile only, receive URLs from YouTube app)

#### C4 Context Diagram (Level 1)

```mermaid
C4Context
  title System Context diagram for y2psync

  Person(user, "User", "A person who wants to save and sync YouTube playlists and subscriptions")
  System_Boundary(y2psync_system, "y2psync") {
    System(y2psync, "y2psync", "Cross-platform app for saving & syncing YouTube playlists and subscriptions")
  }
  System_Ext(youtube, "YouTube (Web)", "Public YouTube website (scraped for video/playlist/channel IDs)")
  System_Ext(android_share, "Android Share System", "Intent system that forwards URLs from YouTube app")
  System_Ext(peers, "Other y2psync Peers", "Other instances of y2psync on the user's devices")
  System_Ext(dht, "Public DHT Network", "Distributed Hash Table for peer discovery")

  Rel(user, y2psync, "Uses", "Desktop UI / Mobile UI")
  Rel(y2psync, youtube, "Scrapes public page metadata", "HTTPS, read-only")
  Rel(y2psync, android_share, "Receives shared URLs via intent", "ACTION_SEND")
  Rel(y2psync, peers, "Syncs playlists & subscriptions", "Encrypted P2P")
  Rel(y2psync, dht, "Peer discovery via rendezvous tag", "DHT protocol")
  UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```

### 3.2 Business Context

y2psync operates as a **personal data management tool** — it does not facilitate sharing or publishing content. All data belongs to the single user and remains on their devices. The app has no business model based on data collection, advertising, or user tracking.

### 3.3 Technical Context

External interfaces:
- **YouTube Web Scrape** — Outbound HTTPS GET to `youtube.com` for publicly visible page HTML. No authentication, no cookies required. Rate-limited to reasonable human-like frequency.
- **Public DHT** — Outbound connections to a public Distributed Hash Table (e.g., libp2p DHT or Mainline DHT) for peer discovery.
- **P2P Sync Connections** — Direct encrypted connections between y2psync peers, initiated after DHT discovery or mDNS LAN discovery.
- **Android Share Intent** — Inbound `ACTION_SEND` with `text/plain` containing a YouTube URL, from any Android application (typically the official YouTube app).

---

## 4. Solution Strategy

### 4.1 Technology Decisions

| Domain | Decision | Rationale |
|--------|----------|-----------|
| Mobile platform | Kotlin + Jetpack Compose | Android-first; modern UI toolkit; native share intent support |
| Desktop platform | Go with Fyne (or Rust with Tauri) | Cross-platform; good P2P library ecosystem; single binary deployment |
| Local database | SQLite (Room on Android, go-sqlite3 or similar on desktop) | Ubiquitous, embedded, zero-configuration, offline-first |
| P2P networking | libp2p (or custom subset using Mainline DHT) | Mature DHT implementation; NAT traversal; encrypted streams |
| Encryption | ChaCha20-Poly1305 + Argon2id + Noise Protocol | Modern, audited, no hardware dependency, constant-time |
| Serialization | Protocol Buffers (protobuf) | Compact, cross-platform, schema-enforced, backward-compatible |

### 4.2 Architecture Pattern

**Local-first with opportunistic sync.** The application follows an offline-first architecture where the local database is the single source of truth. Sync is a background process that reconciles multiple local databases into a consistent state using timestamp-based conflict resolution.

The P2P layer uses a **gossip-style propagation** model: when a device syncs with any peer, it exchanges all changes since the last sync. Changes propagate through the network as devices connect with each other. There is no master node or central coordinator.

### 4.3 Identity Architecture

Two separate identity concepts are maintained to satisfy the contradictory requirements of "identify devices via passphrase" and "anonymous, unlinkable, non-colliding P2P identity":

1. **Peer ID** — 256-bit random value generated locally, never derived from the passphrase. Used on the P2P network as the device's address. Cannot be linked to the Master Sync Key.
2. **Sync Group Key** — Derived from the Master Sync Key via Argon2id. Used to authenticate group membership and encrypt sync data.
3. **Rendezvous Tag** — Derived from the Master Sync Key via a one-way hash (different derivation path from Sync Group Key). Used for DHT peer discovery.

---

## 5. Building Block View

### 5.1 Container Diagram (Level 2)

```mermaid
C4Container
  title Container diagram for y2psync

  Person(user, "User", "A y2psync user")

  System_Boundary(y2psync_mobile, "y2psync Mobile (Android)") {
    Container(mobile_ui, "Mobile UI", "Kotlin + Jetpack Compose", "Provides user interface for managing playlists & subscriptions")
    Container(mobile_db, "Local Database", "Room (SQLite)", "Stores playlists, subscriptions, entries, sync metadata")
    Container(mobile_sync, "Sync Engine", "Kotlin", "Manages peer discovery, sync sessions, conflict resolution")
    Container(mobile_share, "Share Receiver", "Kotlin (Intent Service)", "Handles incoming ACTION_SEND intents from YouTube app")
    Container(mobile_p2p, "P2P Agent", "Kotlin (libp2p bindings)", "Manages DHT discovery and encrypted P2P connections")
  }

  System_Boundary(y2psync_desktop, "y2psync Desktop") {
    Container(desktop_ui, "Desktop UI", "Go + Fyne / Rust + Tauri", "Provides user interface for managing playlists & subscriptions")
    Container(desktop_db, "Local Database", "SQLite", "Stores playlists, subscriptions, entries, sync metadata")
    Container(desktop_sync, "Sync Engine", "Go/Rust", "Manages peer discovery, sync sessions, conflict resolution")
    Container(desktop_p2p, "P2P Agent", "Go/Rust (libp2p)", "Manages DHT discovery and encrypted P2P connections")
  }

  System_Ext(youtube, "YouTube (Web)", "Public YouTube HTML pages")
  System_Ext(dht_network, "Public DHT", "Peer discovery infrastructure")
  System_Ext(android_share_sys, "Android Share System", "Intent dispatch from YouTube app")

  Rel(user, mobile_ui, "Interacts with", "Touch UI")
  Rel(user, desktop_ui, "Interacts with", "Keyboard & Mouse")

  Rel(mobile_ui, mobile_db, "Reads/Writes", "SQL")
  Rel(desktop_ui, desktop_db, "Reads/Writes", "SQL")

  Rel(mobile_ui, mobile_share, "Receives URLs from", "Intent callback")
  Rel(mobile_share, android_share_sys, "Receives shared URLs", "ACTION_SEND")

  Rel(mobile_ui, mobile_sync, "Triggers sync", "API call")
  Rel(desktop_ui, desktop_sync, "Triggers sync", "API call")

  Rel(mobile_sync, mobile_p2p, "Uses", "Sync protocol")
  Rel(desktop_sync, desktop_p2p, "Uses", "Sync protocol")

  Rel(mobile_p2p, dht_network, "Advertises/subscribes", "Rendezvous protocol")
  Rel(desktop_p2p, dht_network, "Advertises/subscribes", "Rendezvous protocol")

  Rel(mobile_p2p, desktop_p2p, "Encrypted sync", "Noise + AEAD")
  Rel(mobile_p2p, mobile_p2p, "Encrypted sync", "Noise + AEAD (peer-to-peer)")

  Rel(mobile_ui, youtube, "Scrapes page metadata", "HTTPS")
  Rel(desktop_ui, youtube, "Scrapes page metadata", "HTTPS")

  UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```

### 5.2 Component Diagram (Level 3) — Sync Engine

```mermaid
C4Component
  title Component diagram for y2psync Sync Engine

  Container_Boundary(sync_engine, "Sync Engine") {
    Component(discovery, "Peer Discovery", "DHT + mDNS", "Finds peer devices sharing the same Rendezvous Tag")
    Component(handshake, "Handshake Manager", "Noise Protocol", "Authenticated key exchange using Sync Group Key")
    Component(sync_session, "Sync Session", "Protocol Buffers", "Manages a sync session with a single peer")
    Component(merge, "Merge Engine", "Timestamp-based CRDT", "Resolves conflicts; newest timestamp wins per entry")
    Component(change_log, "Change Log", "Append-only journal", "Records all local mutations for incremental sync")
    Component(serializer, "Data Serializer", "Protocol Buffers", "Serializes/deserializes sync payloads")
    Component(sync_scheduler, "Sync Scheduler", "Background worker", "Triggers sync on network availability and periodically")
  }

  Container_Boundary(db, "Local Database") {
    Component(playlist_repo, "Playlist Repository", "SQL DAO", "CRUD for playlists and video entries")
    Component(sub_repo, "Subscription Repository", "SQL DAO", "CRUD for subscription lists and channel entries")
    Component(sync_repo, "Sync Metadata Repository", "SQL DAO", "Tracks last sync timestamps per peer, tombstones")
  }

  Rel(discovery, handshake, "Initiates", "Peer found")
  Rel(handshake, sync_session, "Creates", "Authenticated")
  Rel(sync_session, serializer, "Uses", "payload encoding")
  Rel(sync_session, merge, "Calls", "conflict resolution")
  Rel(merge, change_log, "Reads", "local changes")
  Rel(merge, playlist_repo, "Reads/Writes", "playlist data")
  Rel(merge, sub_repo, "Reads/Writes", "subscription data")
  Rel(merge, sync_repo, "Reads/Writes", "sync metadata")
  Rel(sync_scheduler, discovery, "Triggers", "on schedule/event")
  Rel(sync_scheduler, sync_session, "Triggers", "on peers found")

  UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```

### 5.3 Data Model (Entity Relationship)

```
PlaylistList
  - id: UUID (PK)
  - name: string
  - created_at: timestamp (UTC)

PlaylistEntry
  - id: UUID (PK)
  - playlist_list_id: UUID (FK → PlaylistList.id)
  - youtube_video_id: string (YouTube's stable video ID)
  - display_title: string (optional, from page scrape)
  - created_at: timestamp (UTC)  ← birth timestamp per entry
  - sort_order: integer
  - is_deleted: boolean (tombstone)
  - deleted_at: timestamp (nullable)

SubscriptionList
  - id: UUID (PK)
  - name: string
  - created_at: timestamp (UTC)

SubscriptionEntry
  - id: UUID (PK)
  - subscription_list_id: UUID (FK → SubscriptionList.id)
  - youtube_channel_id: string (YouTube's stable channel ID)
  - channel_name: string (optional, from page scrape)
  - channel_url: string
  - created_at: timestamp (UTC)  ← birth timestamp per entry
  - is_deleted: boolean (tombstone)
  - deleted_at: timestamp (nullable)

SyncMetadata
  - peer_id: string
  - last_sync_timestamp: timestamp (UTC)
  - last_sync_status: string
```

---

## 6. Runtime View

### 6.1 Scenario: First-Time Sync Between Two Devices

```mermaid
C4Dynamic
  title Dynamic diagram — First-time sync between two devices

  Container_Boundary(dev_a, "Device A (New, has data)") {
    Component(da_sync, "Sync Engine A", "")
    Component(da_p2p, "P2P Agent A", "")
    Component(da_db, "Local DB A", "")
  }

  Container_Boundary(dev_b, "Device B (New, empty)") {
    Component(db_sync, "Sync Engine B", "")
    Component(db_p2p, "P2P Agent B", "")
    Component(db_db, "Local DB B", "")
  }

  Rel(da_p2p, db_p2p, "1. Discover via DHT Rendezvous Tag", "")
  Rel(da_p2p, db_p2p, "2. Establish Noise Protocol handshake", "")
  Rel(da_p2p, db_p2p, "3. Authenticate with Sync Group Key", "")
  Rel(da_sync, db_sync, "4. Open encrypted sync session", "")
  Rel(da_sync, da_db, "5. Read full dataset", "")
  Rel(da_sync, db_sync, "6. Send full dataset (serialized protobuf)", "")
  Rel(db_sync, db_db, "7. Write all data to local DB", "")
  Rel(db_sync, db_db, "8. Record sync timestamp", "")
  Rel(db_sync, da_sync, "9. Confirm sync complete", "")
  UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="2")
```

### 6.2 Scenario: User Saves Video to Playlist (Mobile, via Share)

```mermaid
C4Dynamic
  title Dynamic diagram — Save YouTube video via share intent

  Person(user, "User")
  System_Ext(youtube_app, "YouTube App")
  Container_Boundary(y2psync_mobile, "y2psync Mobile") {
    Component(share_rx, "Share Receiver", "")
    Component(mobile_ui, "Mobile UI", "")
    Component(mobile_db, "Local DB", "")
    Component(mobile_sync, "Sync Engine", "")
  }

  Rel(user, youtube_app, "1. Opens YouTube video and taps Share", "")
  Rel(youtube_app, share_rx, "2. ACTION_SEND with video URL", "")
  Rel(share_rx, mobile_ui, "3. Parse video ID, show playlist picker", "")
  Rel(user, mobile_ui, "4. Selects target playlist", "")
  Rel(mobile_ui, mobile_db, "5. Insert entry with video ID + timestamp", "")
  Rel(mobile_ui, user, "6. Show confirmation", "")
  Rel(mobile_sync, mobile_db, "7. (Background) Detect change, queue for sync", "")
  UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```

### 6.3 Scenario: Delta Import of YouTube Playlist (Desktop)

```mermaid
sequenceDiagram
  actor User
  participant DesktopUI
  participant DB as Local DB
  participant Scraper as Page Scraper

  User->>DesktopUI: Paste YouTube playlist URL
  DesktopUI->>Scraper: Fetch public playlist page
  Scraper-->>DesktopUI: Return list of video IDs
  DesktopUI->>DesktopUI: Ask user: new or existing playlist?
  User->>DesktopUI: Select existing playlist
  DesktopUI->>DB: Query existing video IDs in target playlist
  DB-->>DesktopUI: Return set of existing IDs
  DesktopUI->>DesktopUI: Compute delta (new IDs only)
  DesktopUI->>DB: Insert new entries with timestamps
  DesktopUI-->>User: Show import summary (X new, Y skipped)
```

---

## 7. Deployment View

### 7.1 Mobile Deployment (Android)

```
┌──────────────────────────────────────┐
│         Android Device                │
│  ┌────────────────────────────────┐  │
│  │  y2psync APK                    │  │
│  │  ┌──────────┐ ┌─────────────┐  │  │
│  │  │ App UI   │ │ Sync Engine │  │  │
│  │  └──────────┘ └─────────────┘  │  │
│  │  ┌──────────┐ ┌─────────────┐  │  │
│  │  │ P2P Agent│ │ Room DB     │  │  │
│  │  └──────────┘ └─────────────┘  │  │
│  └────────────────────────────────┘  │
│  ┌────────────────────────────────┐  │
│  │ Android OS                     │  │
│  │ - WorkManager (bg sync)        │  │
│  │ - Intent System (share rx)     │  │
│  │ - Network Stack                │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

### 7.2 Desktop Deployment

```
┌──────────────────────────────────────┐
│         Desktop Machine               │
│  (Linux / Windows / macOS)           │
│  ┌────────────────────────────────┐  │
│  │  y2psync Desktop Binary         │  │
│  │  ┌──────────┐ ┌─────────────┐  │  │
│  │  │ App UI   │ │ Sync Engine │  │  │
│  │  └──────────┘ └─────────────┘  │  │
│  │  ┌──────────┐ ┌─────────────┐  │  │
│  │  │ P2P Agent│ │ SQLite DB   │  │  │
│  │  └──────────┘ └─────────────┘  │  │
│  └────────────────────────────────┘  │
│  ┌────────────────────────────────┐  │
│  │ OS                             │  │
│  │ - Network Stack                │  │
│  │ - File System (DB file)        │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

### 7.3 Network Topology

```mermaid
C4Deployment
  title Deployment diagram for y2psync

  Deployment_Node(mobile, "Android Device", "Android 8+") {
    Container(mobile_app, "y2psync Mobile", "Kotlin APK")
  }

  Deployment_Node(desktop, "Desktop Computer", "Linux / Windows / macOS") {
    Container(desktop_app, "y2psync Desktop", "Native binary")
  }

  Deployment_Node(lan, "Local Network", "LAN / WiFi") {
    Container_Ext(mdns, "mDNS Discovery", "Multicast DNS")
  }

  Deployment_Node(wan, "Internet", "WAN") {
    Container_Ext(dht, "Public DHT", "libp2p / Mainline DHT")
    Container_Ext(youtube_web, "YouTube", "Web servers")
  }

  Rel(mobile_app, mdns, "LAN peer discovery", "mDNS")
  Rel(desktop_app, mdns, "LAN peer discovery", "mDNS")
  Rel(mobile_app, dht, "WAN peer discovery", "DHT rendezvous")
  Rel(desktop_app, dht, "WAN peer discovery", "DHT rendezvous")
  Rel(mobile_app, desktop_app, "Direct P2P sync", "Encrypted (Noise + AEAD)")
  Rel(mobile_app, mobile_app, "Direct P2P sync", "Encrypted (Noise + AEAD)")
  Rel(desktop_app, desktop_app, "Direct P2P sync", "Encrypted (Noise + AEAD)")
  Rel(mobile_app, youtube_web, "Scrape metadata", "HTTPS")
  Rel(desktop_app, youtube_web, "Scrape metadata", "HTTPS")
```

---

## 8. Crosscutting Concepts

### 8.1 Domain Model

The domain model is shared across all client platforms and is serialised via Protocol Buffers for network sync.

**Core entities:**
- `PlaylistList` — a named collection of video entries
- `PlaylistEntry` — a single YouTube video reference with timestamp
- `SubscriptionList` — a named collection of channel entries
- `SubscriptionEntry` — a single YouTube channel reference with timestamp
- `SyncMetadata` — per-peer tracking of last sync state

**Identity values:**
- `PeerID` — 256-bit random identifier, immutable for device lifetime
- `SyncGroupKey` — derived from Master Sync Key, used for encryption
- `RendezvousTag` — derived from Master Sync Key (separate path), used for discovery

### 8.2 Persistence

- SQLite on all platforms
- Room on Android (with Repository pattern)
- go-sqlite3 or rusqlite on desktop
- Schema migrations handled by versioned migrations
- All timestamps stored as ISO 8601 UTC

### 8.3 Security

- **At rest:** Local database not encrypted by default (relies on device-level encryption). MAY add SQLCipher in future.
- **In transit:** All P2P traffic encrypted with Noise Protocol handshake + AEAD (ChaCha20-Poly1305)
- **Key derivation:** Argon2id for Master Sync Key → Sync Group Key
- **Identity:** Peer ID generated from CSPRNG; never derived from or linkable to passphrase
- **No secrets in code:** No hardcoded keys, tokens, or API credentials

### 8.4 Synchronisation Protocol

The sync protocol uses a **pull-push** model:

1. **Handshake:** Noise Protocol `KK` pattern (both sides have pre-shared key = Sync Group Key)
2. **Exchange:** Each peer sends its full change log since last sync timestamp
3. **Merge:** Timestamp-based conflict resolution — for each entry, the newer `created_at` wins
4. **Tombstones:** Deletions are recorded with timestamp to prevent resurrection
5. **Serialization:** Protocol Buffers over Noise-encrypted stream

### 8.5 Error Handling

- All network operations have timeouts (configurable, default 30s for sync, 10s for DHT)
- Network failures are non-fatal; changes remain queued locally
- Sync conflicts that cannot be resolved by timestamp are resolved deterministically by comparing Peer IDs lexicographically
- Database corruption detected via SQLite integrity check; user prompted to rebuild from sync

### 8.6 Logging and Debugging

- Log levels: ERROR, WARN, INFO, DEBUG
- Log destination: file in application data directory (`y2psync.log`)
- Log rotation: 3 files of 5MB each
- No personal data or User IDs written to logs
- Sync session details (peer ID, timestamp, entry count) logged at INFO level

### 8.7 Testing

- Unit tests for merge engine, URL parsing, data model
- Integration tests for sync protocol (two in-memory peers exchange data)
- Android instrumentation tests for share intent handling
- Desktop CLI flag for headless sync-to-stdout for automated testing

---

## 9. Architectural Decisions

### 9.1 ADR-001: Two Separate Identities (Peer ID vs Sync Group Key)

**Context:** The requirements demand both "identify devices via user passphrase" and "anonymous, unlinkable, non-colliding P2P identity".

**Decision:** Maintain two separate identity concepts. Peer ID is a random 256-bit value with no relation to the passphrase. Sync Group Key is derived from the passphrase via Argon2id.

**Consequences:**
- + Peer ID cannot be used to de-anonymise users
- + Peer ID collisions are practically impossible (256-bit random space)
- + Passphrase changes only change the Sync Group Key, not the Peer ID
- − Requires careful implementation to never leak the relationship between Peer ID and Sync Group Key

### 9.2 ADR-002: Timestamp-Based Conflict Resolution (Not CRDT)

**Context:** The previous specification considered CRDT-based merge. The user clarified that per-entry timestamps should determine conflict resolution.

**Decision:** Use per-entry `created_at` timestamps as the conflict resolution authority. The entry with the newer timestamp always wins. Ties broken by lexicographic Peer ID comparison.

**Consequences:**
- + Simpler than CRDT implementation
- + Deterministic convergence
- + User-intuitive ("I added it later, so I meant it")
- − Requires synchronised clocks (NTP assumed; drift accepted up to reasonable tolerance)
- − Cannot handle true concurrent edits to the same entry at the same nanosecond (extremely rare, tiebroken by Peer ID)

### 9.3 ADR-003: No YouTube API Dependency

**Context:** The user explicitly requires that y2psync NEVER touches YouTube APIs.

**Decision:** All YouTube content interaction is limited to: (a) URL parsing to extract stable IDs, (b) scraping publicly visible HTML pages for display metadata only. No API keys, no OAuth, no authenticated requests to Google.

**Consequences:**
- + Complete user privacy — no Google account needed
- + No rate limiting from API quotas
- + Application works without internet connectivity to Google
- − Page scraping is fragile; YouTube HTML changes may break metadata extraction
- − Cannot programmatically discover user's existing YouTube subscriptions (must be user-shared)

### 9.4 ADR-004: REST API / serverless P2P (no central server)

**Context:** The application must sync without any centralised server.

**Decision:** Use P2P networking with DHT-based peer discovery and direct connections. No relay servers, no STUN/TURN (except what libp2p provides internally).

**Consequences:**
- + No operational cost
- + No central point of failure or surveillance
- − NAT traversal may fail on restrictive networks (some users may need LAN sync only)
- − Both devices must be online simultaneously for sync to occur

### 9.5 ADR-005: Protocol Buffers for Serialization

**Context:** Cross-platform data serialization needed for sync protocol.

**Decision:** Use Protocol Buffers (protobuf v3) for all sync payloads. Schema definitions stored in a shared `.proto` file.

**Consequences:**
- + Compact binary format
- + Schema enforcement and backward compatibility
- + Code generation for Kotlin, Go, Rust, C++
- − Requires schema management discipline

---

## 10. Quality Requirements

### 10.1 Quality Tree

```
Quality Model for y2psync
├── Privacy (highest priority)
│   ├── No data sent to central servers
│   ├── Peer identity cannot be de-anonymised
│   ├── All sync traffic encrypted
│   └── No YouTube API calls
├── Reliability
│   ├── No data loss on sync
│   ├── Deterministic conflict resolution
│   └── Offline operation never blocks the user
├── Usability
│   ├── Sync setup = one passphrase
│   ├── Share intent works from YouTube app (mobile)
│   └── Delta import "just works" (no duplicates)
├── Performance
│   ├── Local operations are instant (< 100ms)
│   ├── Sync completes in reasonable time (< 30s for typical data)
│   └── Background sync doesn't drain battery
└── Portability
    ├── Android mobile (primary)
    ├── Desktop Linux, Windows, macOS
    └── Future: iOS
```

### 10.2 Quality Scenarios

| Scenario | Quality | Given | When | Then |
|----------|---------|-------|------|------|
| Sync convergence | Reliability | Three devices make concurrent offline changes | They sync pairwise | All three converge to identical state |
| Share intent handling | Usability | User is in YouTube app | Shares a video to y2psync | Video added to selected playlist in < 2 seconds |
| Privacy leak | Privacy | Attacker monitors P2P traffic | Collects all observed Peer IDs and payloads | Cannot link any Peer ID to a user identity or Master Sync Key |
| Offline resilience | Reliability | Device has no network | User adds 100 videos to playlists | All saved locally; sync queued; no errors |
| Delta import | Usability | Playlist has 5 entries; YouTube playlist has 5 entries with 3 overlapping | User imports YouTube playlist | Result has 7 entries (no duplicates, existing 5 preserved) |

---

## 11. Risks and Technical Debt

### 11.1 Known Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| YouTube page structure changes break metadata scraping | Medium | Medium | Graceful degradation; display YouTube IDs instead of titles if scrape fails |
| NAT traversal fails for some users | Medium | Medium | Support LAN-only sync via mDNS; document port forwarding needs |
| Two devices generate identical timestamps for the same entry | Low | Low | Tie-breaking by Peer ID ensures deterministic resolution |
| User forgets Master Sync Key | Medium | High | Encourage export of sync settings; consider a written backup mechanism |
| Master Sync Key is weak/guessable | Medium | High | Enforce minimum 12-character passphrase; show strength indicator; use Argon2id to slow brute-force |
| Clock drift causes unexpected conflict resolution | Low | Medium | Use NTP-adjusted UTC; sync protocol includes clock skew tolerance |

### 11.2 Technical Debt (Known)

| Item | Description | Plan |
|------|-------------|------|
| iOS not yet supported | Architecture supports it, but implementation pending | Design protocol and data model to be iOS-compatible from day one |
| No automated recovery from database corruption | SQLite is robust but not immune | Add periodic integrity checks in future release |
| No web UI | Currently only native apps | A web client could be added using the same sync protocol |
| No export/import of data as JSON | Users may want portable backup | Add in v1.1 milestone |

---

## 12. Glossary

| Term | Definition |
|------|------------|
| **Master Sync Key** | User-chosen passphrase used to authenticate and encrypt sync between devices. Never transmitted in plaintext. |
| **Sync Group Key** | Cryptographic key derived from the Master Sync Key via Argon2id, used for group authentication and data encryption. |
| **Peer ID** | 256-bit random identifier generated on each device. Used as the device's address on the P2P network. Cannot be linked to the Master Sync Key. |
| **Rendezvous Tag** | One-way hash derived from the Master Sync Key (separate path from Sync Group Key), used for DHT peer discovery. |
| **DHT** | Distributed Hash Table — a decentralised key-value store used for peer discovery without a central server. |
| **mDNS** | Multicast DNS — a protocol for discovering devices on the local network without configuration. |
| **Tombstone** | A marker indicating that an entry has been deleted, retained to prevent resurrection during sync. Includes deletion timestamp. |
| **Delta Import** | Importing entries from an external source (YouTube URL) without overwriting existing entries or creating duplicates. |
| **Created-at Timestamp** | Per-entry UTC timestamp set when the entry is first added. Used as the authoritative value for conflict resolution during sync. |
| **Noise Protocol** | A framework for building cryptographic protocols with authenticated key exchange. Used for P2P handshake. |
| **AEAD** | Authenticated Encryption with Associated Data — encryption that provides both confidentiality and integrity. |
| **Share Intent** | Android's inter-app communication mechanism (`ACTION_SEND`) used to receive URLs from the YouTube app. |
| **PlaylistList** | A named collection of YouTube video entries created by the user within y2psync. |
| **SubscriptionList** | A named collection of YouTube channel entries created by the user within y2psync. |
| **YouTube Video ID** | The stable, unique 11-character identifier for a YouTube video (e.g., `dQw4w9WgXcQ`). |
| **YouTube Channel ID** | The stable, unique 24-character identifier for a YouTube channel (e.g., `UC_x5XG1OV2P6uZZ5FSM9Ttw`). |
| **YouTube Playlist ID** | The stable identifier for a YouTube playlist (e.g., `PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf`). |
| **Argon2id** | A modern, memory-hard key derivation function resistant to GPU and ASIC attacks. |
| **ChaCha20-Poly1305** | An AEAD cipher providing encryption and authentication. |
| **libp2p** | A modular networking stack for P2P applications, providing DHT, NAT traversal, and encrypted streams. |

---

*This document follows the arc42 template (https://arc42.org) and uses C4 model notation (https://c4model.com) for architecture diagrams via Mermaid.*
