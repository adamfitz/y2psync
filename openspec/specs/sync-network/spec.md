# sync-network Specification

## Purpose

Define how y2psync devices discover each other and synchronise playlist and subscription data over a decentralised peer-to-peer network, with timestamp-based conflict resolution and no central server.

## Requirements

### Requirement: Peer Discovery Without Central Server
Devices SHALL discover each other using a combination of DHT rendezvous and LAN multicast.

#### Scenario: WAN discovery via DHT
- GIVEN a device is online and has a Master Sync Key configured
- WHEN it searches for peer devices
- THEN the system SHALL publish a Rendezvous Tag (derived from the Master Sync Key) on a public Distributed Hash Table
- AND SHALL subscribe to the same Rendezvous Tag to find other devices advertising the same tag
- AND SHALL discover all online devices sharing the same Master Sync Key

#### Scenario: LAN discovery via mDNS
- GIVEN two devices are on the same local network
- WHEN both have the same Master Sync Key configured
- THEN the system SHALL additionally discover peers via mDNS (or equivalent LAN multicast protocol)
- AND SHALL prefer LAN connections over WAN connections when both are available

#### Scenario: No central rendezvous server
- GIVEN the system is operating
- WHEN any device joins the network
- THEN there SHALL be no central server that all devices must register with
- AND discovery SHALL function using only the public DHT and optional LAN multicast

### Requirement: Encrypted Synchronisation Channel
All sync traffic SHALL be encrypted end-to-end between devices.

#### Scenario: Authenticated key exchange
- GIVEN Device A and Device B have discovered each other
- WHEN they establish a sync connection
- THEN they SHALL perform an authenticated key exchange using the Sync Group Key
- AND SHALL negotiate a session key for encrypting all subsequent data

#### Scenario: Payload encryption
- GIVEN a sync session is established
- WHEN data is transmitted between devices
- THEN all payload data SHALL be encrypted using AEAD (ChaCha20-Poly1305 or AES-256-GCM)
- AND the encrypted payload SHALL include authentication to prevent tampering

### Requirement: Data Synchronisation
Devices SHALL synchronise playlist and subscription data bidirectionally, with timestamp-based conflict resolution.

#### Scenario: Full sync on connect
- GIVEN two devices establish a sync session for the first time (or after extended disconnection)
- WHEN the session is established
- THEN each device SHALL send its full dataset (all playlists with all entries, all subscription lists with all entries) to the other device
- AND each device SHALL merge the received data with its local data

#### Scenario: Incremental sync
- GIVEN two devices have previously synced
- WHEN they establish a sync session
- THEN each device SHALL send only the changes since the last sync (new entries, deleted entries, modified entry timestamps)

#### Scenario: Timestamp-based merge for entries
- GIVEN Device A has entry E with timestamp T_A and Device B has entry E with timestamp T_B
- WHEN merging
- THEN the entry with the later (newer) timestamp SHALL be kept on both devices
- AND the entry with the earlier (older) timestamp SHALL be discarded

#### Scenario: No-duplicates merge
- GIVEN Device A has entries E1, E2 and Device B has entries E1, E3
- WHEN merging playlists or subscription lists
- THEN the resulting set SHALL contain E1, E2, E3
- AND SHALL NOT contain duplicate E1
- AND deduplication SHALL use the stable YouTube ID (video ID for playlists, channel ID for subscriptions) as the identity key, NOT the title or URL

#### Scenario: Deletion sync
- GIVEN a user deletes an entry on Device A
- WHEN the next sync occurs
- THEN Device B SHALL also delete that entry
- AND a deleted entry SHALL NOT be resurrected by an older copy on another device (deletions SHALL be tracked via a tombstone mechanism with timestamp)

#### Scenario: Out-of-order sync
- GIVEN devices A and B both modify data while offline
- WHEN they reconnect
- THEN the timestamp-based merge SHALL handle concurrent modifications deterministically: each entry's timestamp determines which copy wins
- AND the sync SHALL converge to the same state on all devices

### Requirement: Offline Operation
Devices SHALL operate fully offline and sync opportunistically when connectivity is available.

#### Scenario: Full offline read/write
- GIVEN a device has no network connectivity
- WHEN the user creates, edits, or deletes playlists and subscriptions
- THEN all operations SHALL succeed immediately against the local data store
- AND the changes SHALL be queued for sync when connectivity is restored

#### Scenario: Queued sync on reconnect
- GIVEN a device has accumulated changes while offline
- WHEN network connectivity is restored
- THEN the system SHALL automatically attempt to discover peers and sync all queued changes
- AND SHALL not require user intervention to trigger sync

#### Scenario: Partial connectivity
- GIVEN only some peer devices are online
- WHEN a device syncs with available peers
- THEN changes SHALL propagate through the network as each peer connects with others (gossip-style propagation)
- AND eventually all devices SHALL converge

### Requirement: Sync Group Membership
Devices SHALL be able to leave and join sync groups.

#### Scenario: New device joins group
- GIVEN a device with the correct Master Sync Key
- WHEN it connects to any existing group member
- THEN the existing member SHALL share the full dataset with the new device
- AND the new device SHALL have a complete copy of all playlists and subscriptions

#### Scenario: Device leaves group
- GIVEN a user changes their Master Sync Key on one device
- WHEN the device next connects to peers with the old key
- THEN the device SHALL be unable to authenticate with the old peer group
- AND SHALL no longer participate in sync
