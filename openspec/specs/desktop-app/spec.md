# desktop-app Specification

## Purpose

Define desktop-specific behaviours of the y2psync application, including URL-based import, full playlist and subscription management, and P2P sync.

## Requirements

### Requirement: Playlist Import from URL
The desktop application SHALL support importing YouTube playlists via URL.

#### Scenario: Import YouTube playlist by URL
- GIVEN the user has a YouTube playlist URL
- WHEN they paste it into the import dialog
- THEN the system SHALL extract the playlist ID from the URL
- AND SHALL scrape the publicly accessible YouTube playlist page to obtain the list of video IDs
- AND SHALL optionally fetch publicly visible video titles from the page metadata (for display purposes only)
- AND SHALL NOT use any YouTube API

#### Scenario: Target selection
- GIVEN the playlist URL has been processed and video IDs discovered
- WHEN the user proceeds
- THEN the system SHALL present options: create a new playlist with the discovered videos, or add the discovered videos to an existing playlist (delta merge)

#### Scenario: Import report
- GIVEN the import process completes
- WHEN the operation finishes
- THEN the system SHALL display a summary: total videos discovered, new videos added, duplicates skipped, any videos that could not be processed

### Requirement: Channel Subscription Import
The desktop application SHALL support adding channel subscriptions via URL and bulk import.

#### Scenario: Subscribe to channel by URL
- GIVEN the user has a YouTube channel URL
- WHEN they paste it into the subscription dialog
- THEN the system SHALL extract the channel ID from the URL
- AND SHALL present options: add to existing subscription list, or create a new subscription list

#### Scenario: Bulk import from file
- GIVEN the user has a text file or CSV with YouTube channel URLs
- WHEN they import the file
- THEN the system SHALL parse each line as a URL
- AND SHALL attempt to extract the channel ID from each
- AND SHALL add valid channels to the chosen subscription list (delta merge)
- AND SHALL report invalid/unparseable URLs

### Requirement: Desktop Sync Management
The desktop application SHALL provide sync status visibility and control.

#### Scenario: Sync status display
- GIVEN the application is running
- WHEN the user views the sync panel
- THEN the system SHALL display: current sync status (idle, discovering peers, syncing, error), number of known peer devices, last successful sync time, list of pending changes to sync

#### Scenario: Sync key management
- GIVEN the user wants to configure or change sync
- WHEN they open sync settings
- THEN the system SHALL support: setting a Master Sync Key for the first time, changing the Master Sync Key (requires re-authenticating all devices), clearing the Master Sync Key (disables sync, keeps local data)

#### Scenario: Sync key change propagation
- GIVEN the user changes their Master Sync Key on the desktop
- WHEN the key is changed
- THEN the system SHALL generate a new Sync Group Key and Rendezvous Tag
- AND the device SHALL be unable to sync with devices still using the old key
- AND the user SHALL be prompted to update the key on all other devices

### Requirement: Clipboard URL Detection
The desktop application MAY detect YouTube URLs in the clipboard.

#### Scenario: Clipboard monitoring
- GIVEN the application is running
- WHEN the user copies a YouTube URL to their clipboard
- THEN the system MAY show a notification suggesting an import
- AND the user MAY dismiss the notification without action
