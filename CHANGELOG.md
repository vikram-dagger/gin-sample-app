# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-05-15
### Added
- Delete book endpoint (`DELETE /api/books/{id}`) to remove a book by ID.
- New route for deleting books added to the routes configuration.
- Tests for book deletion, covering deletion of existing and non-existent books.

### Changed
- RegisterRoutes in `routes/routes.go` updated to include DELETE method.
