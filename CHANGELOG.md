# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.2.0] - 2021-02-22
### Added
* Add a configuration script to check for Bazel and update PATH
* Add equivalent script for Windows execute/test

### Changed
* Change how the root path is discovered.
* Reorganize the warning and info messages displayed for projectIDs
* Convert "Production" to "Prod" in release-channels list
* Update versions command description
* Update descriptions of commands
* Update build instructions

### Fixed
* Escape version-id to prevent URLs being sent in it
* Fix test when token file already exists on system
* Fix Mac and Windows presubmit tasks

## [3.1.0] - 2020-10-16
### Added
* Add CLI command to list release-channels
* Add CLI support for DeviceFulfillment and EntitySet ConfigFiles
* Add CLI command to list release-channels

### Changed
* Switch v2alpha to v2 endpoint
* Create presubmit script to build and run tests

### Removed
* Remove examples from repo

### Fixed
* Add golang_x_sys to workspace
* Add Mousetrap dep to allow Windows build target

## [3.0.0] - 2020-06-17
### Added
* Initial commit
