# y2psync-overview Specification

## Purpose

y2psync is a cross-platform application that allows users to save, organise, and sync YouTube playlists and channel subscriptions across their own devices without requiring a centralised server, user account, or any YouTube API access. The application uses a decentralised peer-to-peer sync mechanism authenticated by a user-chosen passphrase (Master Sync Key).

## Requirements

### Requirement: Application Platform
The system SHALL target Android (mobile/tablet) and desktop Linux/Windows/macOS.

#### Scenario: Mobile target
- GIVEN the system is built for mobile
- WHEN the build target is Android
- THEN the mobile application SHALL be implemented in Kotlin
- AND the application SHALL target Android 8.0 (API 26) or later

#### Scenario: Desktop target
- GIVEN the system is built for desktop
- WHEN the build target is selected
- THEN the desktop application MAY be implemented in Go, Rust, C++, or another systems language
- AND the application SHALL run on Linux, Windows, and macOS

#### Scenario: iOS portability
- GIVEN future porting is considered
- WHEN the system is extended to iOS
- THEN the architecture SHALL support adding an iOS client without rearchitecting the sync protocol or data model

### Requirement: No YouTube API Dependency
The system SHALL NOT use or depend on any YouTube API (OAuth, YouTube Data API v3, YouTube Reporting API, or any other YouTube programmatic interface).

#### Scenario: No API calls made
- GIVEN the application is running
- WHEN the user adds any YouTube content
- THEN the application SHALL NOT make any HTTP request to `www.googleapis.com` or any YouTube API endpoint
- AND the application SHALL NOT require the user to sign in to a Google/YouTube account

#### Scenario: All content via URLs
- GIVEN the user wants to add YouTube content
- WHEN the user provides a YouTube URL
- THEN the application SHALL parse the URL to extract stable YouTube IDs (video ID, playlist ID, channel ID)
- AND SHALL NOT attempt to fetch metadata via API
- AND MAY scrape publicly visible page metadata from the YouTube watch/playlist/channel page for display purposes only

### Requirement: Google / Youtube independence
The system SHALL function without any dependency on Google services or YouTube API access.

#### Scenario: Google Play services not required
- GIVEN the application is installed on an Android device
- WHEN the application starts
- THEN the application SHALL NOT require Google Play Services
- AND SHALL be installable via F-Droid or direct APK

### Requirement: Core Data Types
The system SHALL manage two fundamental data types: YouTube Playlists and YouTube Channel Subscriptions.

#### Scenario: Playlist data model
- GIVEN the user creates a playlist
- THEN the playlist SHALL have: a unique local UUID, a user-defined name, a created-at timestamp, a list of video entries
- AND each video entry SHALL have: a YouTube video ID, a user-supplied title (optional, for display), a created-at timestamp (birth timestamp), and an ordering position within the playlist

#### Scenario: Subscription data model
- GIVEN the user creates a subscription list
- THEN the subscription list SHALL have: a unique local UUID, a user-defined name, a created-at timestamp, a list of channel entries
- AND each channel entry SHALL have: a YouTube channel ID, a channel name (optional, for display), a created-at timestamp (birth timestamp), and the channel URL
