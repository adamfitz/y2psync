# identity Specification

## Purpose

Define how y2psync establishes device identity and sync group membership without compromising user anonymity. The system uses two separate identity concepts: a collision-resistant anonymous Peer ID for P2P network identification, and a Sync Group Key derived from the user's Master Sync Key passphrase for group authentication and encryption.

## Requirements

### Requirement: Peer ID (Device Identity on P2P Network)
Every device SHALL generate a unique, anonymous Peer ID that cannot be linked to the user, the Master Sync Key, or any other device.

#### Scenario: Peer ID generation on first launch
- GIVEN the application is launched for the first time
- WHEN the local identity store is empty
- THEN the system SHALL generate a new Peer ID
- AND the Peer ID SHALL be a random 256-bit value generated from a cryptographically secure RNG
- AND the Peer ID SHALL be stored locally and never transmitted in plaintext alongside any identifying information
- AND the Peer ID SHALL NOT be derived from or deterministically linked to the Master Sync Key

#### Scenario: Peer ID collision resistance
- GIVEN two independently-operated y2psync devices
- WHEN both generate Peer IDs
- THEN the probability of collision SHALL be negligible (less than 2^-128) due to 256-bit random generation

#### Scenario: Peer ID persistence
- GIVEN a Peer ID has been generated
- WHEN the application is restarted
- THEN the same Peer ID SHALL be reused for the lifetime of the application data on that device
- AND the Peer ID SHALL NOT change unless the application data is wiped

#### Scenario: Peer ID cannot de-anonymise
- GIVEN an attacker observes the P2P network
- WHEN they capture a Peer ID
- THEN they SHALL NOT be able to determine: the user's identity, the user's Master Sync Key, the device manufacturer, the device model, the device's geographic location (beyond IP geolocation inherent to the network), or any association with other Peer IDs belonging to the same user

### Requirement: Master Sync Key (Group Authentication)
The system SHALL use a user-chosen passphrase (Master Sync Key) to authenticate devices as belonging to the same sync group and to derive encryption keys.

#### Scenario: Master Sync Key creation
- GIVEN the user wants to enable sync
- WHEN they create a Master Sync Key
- THEN the system SHALL accept a passphrase of at least 12 characters with no maximum length
- AND the system SHALL NOT transmit the passphrase in plaintext over the network
- AND the system SHALL NOT store the passphrase in plaintext on disk

#### Scenario: Sync Group Key derivation
- GIVEN a Master Sync Key is provided
- WHEN the system needs to authenticate or encrypt
- THEN the system SHALL derive a Sync Group Key from the Master Sync Key using a key derivation function (Argon2id or equivalent)

#### Scenario: Rendezvous Tag derivation
- GIVEN a Master Sync Key is provided
- WHEN the system needs to discover peer devices on the DHT
- THEN the system SHALL derive a Rendezvous Tag from the Master Sync Key using a one-way hash, using a different derivation path than the Sync Group Key

#### Scenario: Device admission to sync group
- GIVEN a device with Peer ID A wants to join the sync group
- WHEN it presents a valid Sync Group Key derived from the Master Sync Key
- THEN existing group members SHALL accept it as a group member
- AND the new device SHALL receive the encrypted current state from existing members

### Requirement: No Central Identity Provider
The system SHALL NOT use any centralised identity provider, OAuth provider, or account system.

#### Scenario: No account creation
- GIVEN the user launches the application
- WHEN they navigate to settings
- THEN there SHALL be no option to create an account, sign in, or register with any service
- AND the only authentication mechanism SHALL be the local Master Sync Key passphrase

### Requirement: Cryptographic Agility
The system SHALL use well-vetted, audited cryptographic primitives.

#### Scenario: Algorithms used
- GIVEN the system performs cryptographic operations
- THEN the Peer ID generation SHALL use a CSPRNG (e.g., /dev/urandom, Java SecureRandom, Go crypto/rand)
- AND key derivation SHALL use Argon2id
- AND network encryption SHALL use AEAD (e.g., ChaCha20-Poly1305 or AES-256-GCM)
- AND the initial handshake SHALL use a Noise protocol or similar authenticated key exchange
