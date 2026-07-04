# playlist-management Specification

## Purpose

Define how y2psync manages YouTube playlists and their video entries, including creation, editing, delta import from YouTube, and deduplication.

## Requirements

### Requirement: Playlist CRUD
Users SHALL be able to create, read, update, and delete playlists locally.

#### Scenario: Create playlist
- GIVEN the user chooses to create a new playlist
- WHEN they enter a name
- THEN the system SHALL create a playlist with: a unique UUID identifier, the user-provided name, the current system time as the playlist created-at timestamp, and an empty list of video entries
- AND the playlist SHALL appear in the user's playlist list immediately

#### Scenario: Rename playlist
- GIVEN a playlist exists
- WHEN the user edits its name
- THEN the system SHALL update the playlist name to the new value
- AND SHALL NOT affect the playlist's entries or their timestamps

#### Scenario: Delete playlist
- GIVEN a playlist exists
- WHEN the user deletes it
- THEN the system SHALL remove the playlist and all its entries from the local data store
- AND SHALL propagate the deletion on next sync via a tombstone record

### Requirement: Video Entry Management
Users SHALL be able to add and remove individual video entries from playlists.

#### Scenario: Add video by URL
- GIVEN the user has a YouTube video URL
- WHEN they choose to add it to a playlist
- THEN the system SHALL parse the URL to extract the YouTube video ID
- AND SHALL create a new video entry with: the extracted video ID, the current system time as the entry's created-at timestamp, and an auto-assigned ordering position at the end of the playlist
- AND SHALL NOT make any API call to YouTube

#### Scenario: Remove video from playlist
- GIVEN a playlist has a video entry
- WHEN the user removes it
- THEN the system SHALL remove the entry from the playlist
- AND the remaining entries SHALL retain their ordering positions

#### Scenario: Duplicate prevention within playlist
- GIVEN a playlist already contains a video with ID "abc123"
- WHEN the user tries to add the same video ID "abc123" again
- THEN the system SHALL NOT add a duplicate entry
- AND SHALL notify the user that the video is already in the playlist

### Requirement: Delta Import from YouTube URL (Desktop)
The desktop application SHALL import video entries from a YouTube playlist URL without overwriting existing entries or creating duplicates.

#### Scenario: Import playlist URL
- GIVEN the user provides a YouTube playlist URL (e.g., `https://www.youtube.com/playlist?list=PL...`)
- WHEN they choose to import into a new or existing playlist
- THEN the system SHALL extract the playlist ID from the URL
- AND SHALL scrape the publicly accessible playlist page to discover video IDs
- AND SHALL add each discovered video ID as a new entry in the target playlist
- AND SHALL NOT remove any existing entries in the target playlist
- AND SHALL NOT add a video ID that already exists in the target playlist

#### Scenario: Existing import target has prior entries
- GIVEN a target playlist already contains video IDs A, B, C
- WHEN the user imports a YouTube playlist containing video IDs B, D, E
- THEN the resulting playlist SHALL contain A, B, C, D, E
- AND SHALL NOT contain duplicate B
- AND entry order SHALL preserve existing order for A, B, C and append D, E at the end

#### Scenario: Import into new playlist
- GIVEN the user provides a YouTube playlist URL
- WHEN they choose to import into a new playlist
- THEN the system SHALL create a new playlist with a user-specified name
- AND SHALL add all discovered video IDs as entries
- AND the playlist SHALL be available for sync immediately

### Requirement: Delta Import from Share Intent (Mobile)
The mobile application SHALL import video entries from a shared YouTube link via the Android share intent system.

#### Scenario: Share video to y2psync
- GIVEN the user shares a YouTube video URL from the official YouTube app
- WHEN the share target is y2psync
- THEN the system SHALL receive the URL via Android's `ACTION_SEND` intent
- AND SHALL present a list of available playlists to add to
- AND SHALL add the video ID to the chosen playlist (following the same deduplication rules above)

#### Scenario: Share playlist URL to y2psync
- GIVEN the user shares a YouTube playlist URL from the official YouTube app
- WHEN the share target is y2psync
- THEN the system SHALL receive the URL via Android's `ACTION_SEND` intent
- AND SHALL offer the user the choice to import all videos into a new or existing playlist
- AND SHALL perform the same delta import as the desktop application

### Requirement: Timestamp Metadata Per Entry
Every video entry in a playlist SHALL carry its own birth timestamp, distinct from the playlist's creation timestamp.

#### Scenario: Entry timestamp on creation
- GIVEN a new video entry is added to a playlist
- WHEN the entry is created
- THEN the entry SHALL store the current system time (UTC) as its `created_at` timestamp
- AND this timestamp SHALL NOT change on subsequent sync operations or reordering

#### Scenario: Timestamp used for sync conflict resolution
- GIVEN two devices add the same video ID to the same playlist at different times
- WHEN they sync
- THEN the entry with the newer `created_at` timestamp SHALL determine the entry's position in the merged playlist
- AND if timestamps are equal, the higher Peer ID (lexicographic) SHALL win deterministically
