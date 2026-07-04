# OpenSpec — y2psync Specification Index

This directory contains Gherkin-style specifications for the y2psync application. Each spec describes the behavior of a major capability domain.

## Specs

| Spec | File | Covers |
|------|------|--------|
| [Overview](specs/y2psync-overview/spec.md) | `specs/y2psync-overview/spec.md` | Application purpose, high-level requirements, platform targets |
| [Identity](specs/identity/spec.md) | `specs/identity/spec.md` | Master Sync Key, Peer ID generation, anonymous P2P identity, group membership |
| [Sync Network](specs/sync-network/spec.md) | `specs/sync-network/spec.md` | P2P peer discovery, DHT rendezvous, mDNS LAN fallback, encrypted sync protocol, conflict resolution via timestamps |
| [Playlist Management](specs/playlist-management/spec.md) | `specs/playlist-management/spec.md` | Playlist CRUD, video entry management, delta import, deduplication via YouTube video ID |
| [Subscription Management](specs/subscription-management/spec.md) | `specs/subscription-management/spec.md` | Channel subscription CRUD, share-to-add, delta import, deduplication via YouTube channel ID |
| [Mobile App](specs/mobile-app/spec.md) | `specs/mobile-app/spec.md` | Android-specific behaviours: share intent receiver, playlist URL import, subscription import |
| [Desktop App](specs/desktop-app/spec.md) | `specs/desktop-app/spec.md` | Desktop-specific behaviours: YouTube URL parsing, playlist import, channel subscription management |

## Format

Each spec follows: **Purpose** → **Requirements** → **Scenarios** (Given/When/Then with SHALL/SHOULD/MAY per RFC 2119).

## Key Design Principles

- **No YouTube APIs** — y2psync NEVER calls YouTube OAuth or YouTube Data API v3. All interaction with YouTube content is via user-supplied URLs, processed through the Android share intent system (mobile) or manual URL entry (desktop).
- **Anonymous P2P Identity** — Each device generates a local random Peer ID (256-bit) with collision probability negligible. The Master Sync Key passphrase is used only for sync group authentication and encryption, NEVER for identity. Peer ID cannot be linked to the passphrase or de-anonymise the user.
- **Timestamp-Based Merge** — Each list entry (video in a playlist, channel in a subscription list) carries a created-at timestamp. On sync conflict, the entry with the newest timestamp wins. The playlist/subscription list itself has its own creation timestamp distinct from its entries.
- **Delta-Only Import** — Importing a YouTube playlist or subscription list NEVER overwrites existing entries and NEVER produces duplicates. Deduplication uses stable YouTube IDs (video ID, channel ID, playlist ID).
