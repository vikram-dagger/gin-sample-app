# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-05-15
### Changed
- Maintenance update: No user-facing changes. Internal configuration and CI files (.dagger directory) were cleaned up. No Go source code or API changes were made.

## [1.0.0] - 2025-05-14
### Added
- Initial release of the Book API
- Create book endpoint (`POST /api/books`)
- List all books endpoint (`GET /api/books`)
- Get book by ID endpoint (`GET /api/books/{id}`)
- OpenAPI specification describing all endpoints and models
- Basic book model with title and author fields
- GORM integration for database operations
- RESTful API using Gin framework

### Changed
- N/A (initial release)

### Removed
- N/A (initial release)
