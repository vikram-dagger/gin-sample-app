# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-05-14
### Added
- Added `DELETE /api/books/{id}` endpoint for deleting books by ID.
- Registered the new route in Gin router.
- Added tests for DeleteBook handler covering deletion and non-existent books.

### Changed
- Updated codebase to support book deletion functionality in controller and routes.

### Note
- The OpenAPI specification (`openapi.yaml`) has NOT yet been updated to include the new DELETE endpoint.

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
