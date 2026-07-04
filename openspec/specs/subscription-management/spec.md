# subscription-management Specification

## Purpose

Define how y2psync manages YouTube channel subscriptions, including adding channels via shared links or manual URL entry, organising subscriptions into named lists, and delta import without duplicates.

## Requirements

### Requirement: Subscription List CRUD
Users SHALL be able to create, read, update, and delete named subscription lists locally.

#### Scenario: Create subscription list
- GIVEN the user chooses to create a new subscription list
- WHEN they enter a name
- THEN the system SHALL create a subscription list with: a unique UUID identifier, the user-provided name, the current system time as the list created-at timestamp, and an empty list of channel entries

#### Scenario: Rename subscription list
- GIVEN a subscription list exists
- WHEN the user edits its name
- THEN the system SHALL update the list name
- AND SHALL NOT affect the channel entries

#### Scenario: Delete subscription list
- GIVEN a subscription list exists
- WHEN the user deletes it
- THEN the system SHALL remove the list and all its channel entries
- AND SHALL propagate the deletion via tombstone on next sync

### Requirement: Add Channel by URL
Users SHALL be able to add YouTube channels to subscription lists by providing channel URLs.

#### Scenario: Add channel via share intent (mobile)
- GIVEN the user shares a YouTube channel URL or channel video URL from the official YouTube app
- WHEN the share target is y2psync
- THEN the system SHALL extract the channel ID from the URL
- AND SHALL present a list of available subscription lists to add to
- AND SHALL add the channel entry to the chosen list
- AND SHALL NOT make any YouTube API call

#### Scenario: Add channel via manual URL entry (desktop)
- GIVEN the user has a YouTube channel URL (e.g., `https://www.youtube.com/@ChannelName` or `https://www.youtube.com/channel/UC...`)
- WHEN they add it to a subscription list
- THEN the system SHALL parse the URL to extract the channel ID
- AND SHALL create a channel entry with: the channel ID, the current system time as the entry's created-at timestamp, and the original URL for reference

#### Scenario: Duplicate prevention within list
- GIVEN a subscription list already contains channel ID "UCabc123"
- WHEN the user tries to add the same channel ID again
- THEN the system SHALL NOT add a duplicate entry
- AND SHALL notify the user

### Requirement: Delta Import of Multiple Subscriptions
Users SHALL be able to import multiple channels into a subscription list in bulk.

#### Scenario: Import from shared playlist (mobile)
- GIVEN the user shares a YouTube playlist URL
- WHEN they choose to import the channel of each video as a subscription
- THEN the system SHALL scrape the playlist page to extract video IDs
- AND SHALL (optionally) attempt to discover the channel ID for each video from the publicly visible page metadata
- AND SHALL add each discovered channel to the chosen subscription list using delta semantics

#### Scenario: Import subscription list from URL (desktop)
- GIVEN the user provides a YouTube channel or playlist URL
- WHEN they import into a subscription list
- THEN the system SHALL add the channel as a new entry in the target list
- AND SHALL NOT remove any existing entries
- AND SHALL NOT add duplicate channel IDs

#### Scenario: Bulk import from YouTube subscriptions page (user-mediated)
- GIVEN the user has a list of YouTube channel URLs they manually collected or exported
- WHEN they provide multiple URLs (e.g., one per line, or from a file)
- THEN the system SHALL process each URL, extract the channel ID, and add it to the chosen subscription list
- AND SHALL report which URLs were successfully added and which could not be parsed

### Requirement: Per-Entry Timestamps
Every channel entry SHALL carry its own birth timestamp for sync conflict resolution.

#### Scenario: Channel entry timestamp
- GIVEN a new channel is added to a subscription list
- WHEN the entry is created
- THEN the entry SHALL store the current system time (UTC) as its `created_at` timestamp

#### Scenario: Conflict resolution
- GIVEN two devices add the same channel ID to the same subscription list at different times
- WHEN they sync
- THEN the entry with the newer `created_at` timestamp SHALL determine the entry's metadata in the merged list
