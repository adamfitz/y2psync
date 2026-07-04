# mobile-app Specification

## Purpose

Define Android-specific behaviours of the y2psync mobile application, with particular focus on receiving shared content from the official YouTube application via the Android intent system.

## Requirements

### Requirement: Share Intent Receiver
The Android application SHALL register as a target for shared URLs from other applications, specifically the official YouTube application.

#### Scenario: Register intent filter
- GIVEN the Android application is installed
- WHEN the system checks available share targets
- THEN the application SHALL appear as a share target for `text/plain` content
- AND the intent filter SHALL match `ACTION_SEND` with `text/plain` MIME type

#### Scenario: Receive YouTube video URL
- GIVEN the user is in the official YouTube application
- WHEN they tap "Share" on a video and select y2psync
- THEN the application SHALL receive the YouTube video URL from the intent
- AND SHALL parse the URL to extract the YouTube video ID
- AND SHALL present a UI for the user to select which playlist to add the video to
- AND SHALL support creating a new playlist from this flow

#### Scenario: Receive YouTube playlist URL
- GIVEN the user is in the official YouTube application
- WHEN they tap "Share" on a playlist and select y2psync
- THEN the application SHALL receive the YouTube playlist URL
- AND SHALL parse the URL to extract the playlist ID
- AND SHALL present the user with options: import all videos to a new playlist, import all videos to an existing playlist, or cancel

#### Scenario: Receive YouTube channel URL
- GIVEN the user is in the official YouTube application
- WHEN they tap "Share" on a channel and select y2psync
- THEN the application SHALL receive the YouTube channel URL
- AND SHALL extract the channel ID
- AND SHALL present the user with options: add to an existing subscription list, create a new subscription list, or cancel

#### Scenario: Dispatch based on URL type
- GIVEN the application receives a shared URL
- WHEN the URL is parsed
- THEN the application SHALL determine whether it is a video URL, playlist URL, or channel URL
- AND SHALL route to the appropriate handler for each type

### Requirement: Local Playlist and Subscription Storage
The mobile application SHALL store all data locally using a local database.

#### Scenario: Local persistence
- GIVEN the application has data
- WHEN the application is closed and reopened
- THEN all playlists, subscription lists, and their entries SHALL be preserved
- AND SHALL be immediately available offline

#### Scenario: Database technology
- GIVEN the application needs local storage
- THEN the system SHALL use Room (Android's SQLite abstraction) or an equivalent local database
- AND the database schema SHALL support all fields defined in the data model (UUIDs, YouTube IDs, timestamps, metadata)

### Requirement: Sync Trigger
The mobile application SHALL trigger sync automatically when network conditions allow.

#### Scenario: Background sync
- GIVEN the device has network connectivity
- WHEN the user has configured a Master Sync Key
- THEN the application SHALL periodically attempt to discover peers and synchronise in the background
- AND SHALL respect Android battery optimisation constraints (using WorkManager or equivalent)

#### Scenario: Manual sync
- GIVEN the user is on the sync settings screen
- WHEN they tap "Sync Now"
- THEN the application SHALL immediately attempt peer discovery and synchronisation
- AND SHALL show progress indicators during the sync process

### Requirement: First-Run Experience
The mobile application SHALL guide the user through initial setup.

#### Scenario: First launch
- GIVEN the application is launched for the first time
- WHEN the setup screen is shown
- THEN the user SHALL be offered the choice to: create a new Master Sync Key, configure an existing Master Sync Key, or skip sync setup (use locally only)
- AND the application SHALL display a brief explanation of what the Master Sync Key is and how it protects privacy

#### Scenario: Skip sync
- GIVEN the user chooses to skip sync setup
- WHEN they use the application
- THEN the application SHALL function fully as a local-only playlist and subscription manager
- AND SHALL allow sync setup at any later time from settings
