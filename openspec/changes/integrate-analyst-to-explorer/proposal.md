## Why

The analyst CLI provides critical functionality for schema evolution (analyzing entity type candidates and promoting them to formal schemas), but it requires users to switch between the CLI and the explorer Wails app. Integrating this into the explorer creates a unified workflow where users can analyze, visualize, and promote entity types in a single application, improving productivity and user experience.

## What Changes

- Add analyst functionality to the explorer Wails app's backend
- Expose two new capabilities through the Wails API:
  - Entity type analysis and ranking with configurable thresholds
  - Interactive entity type promotion workflow
- Create UI components for displaying ranked candidates and managing promotions
- Reuse existing analyst package logic (`internal/analyst`) and promote package logic (`internal/promoter`) rather than reimplementing

## Capabilities

### New Capabilities
- `entity-analysis`: Analyze discovered entities, detect patterns, cluster, and rank candidates for type promotion with configurable thresholds (min occurrences, consistency, top N)
- `entity-promotion`: Promote a discovered entity type to a formal Ent schema with property generation, validation, file creation, and data migration

### Modified Capabilities

## Impact

**Affected Code:**
- `cmd/explorer/app.go`: Add new Wails-bound methods for analysis and promotion
- `cmd/explorer/frontend/`: New UI components for analyst features
- `internal/analyst/`: No changes needed, will be consumed by explorer

**Dependencies:**
- Reuses existing `internal/analyst` package (detector, ranker, schema_gen)
- Reuses existing `internal/promoter` package for promotion workflow
- Requires database connection (already available in explorer app)

**Systems:**
- Explorer Wails application (backend and frontend)
- Database (reading discovered entities, executing promotions)
