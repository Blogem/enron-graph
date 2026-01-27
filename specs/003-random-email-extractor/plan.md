# Implementation Plan: Random Email Extractor

**Branch**: `003-random-email-extractor` | **Date**: 2026-01-27 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-random-email-extractor/spec.md`

## Summary

Create a command-line utility that extracts random email samples from the Enron emails CSV file (`assets/enron-emails/emails.csv`) with configurable sample sizes, tracking previously extracted emails to prevent duplicates across multiple runs. The tool will integrate with the existing Go codebase structure under `cmd/sampler/` and reuse the proven CSV parsing logic from `internal/loader/parser.go`.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: 
- `encoding/csv` (stdlib) - CSV reading/writing, already used in `internal/loader/parser.go`
- `flag` (stdlib) - Command-line argument parsing (consistent with other cmd/ tools)
- `math/rand` - Randomization for email selection
- `time` - Timestamp generation for output filenames

**Storage**: 
- Input: `assets/enron-emails/emails.csv` (source data, ~500k emails, 2 columns: file, message)
- Output: Timestamped CSV files in `assets/enron-emails/` (format: `sampled-emails-YYYYMMDD-HHMMSS.csv`)
- Tracking: One text file per extraction in `assets/enron-emails/` (format: `extracted-YYYYMMDD-HHMMSS.txt`, one identifier per line)

**Testing**: 
- Go testing framework (`testing` package)
- Test fixtures in `tests/fixtures/`
- Integration tests in `tests/integration/`
- Table-driven tests for edge cases

**Target Platform**: macOS/Linux command-line (development environments)  

**Project Type**: Single project (Go monorepo with cmd/ structure)

**Performance Goals**: 
- Extract up to 10,000 emails in under 10 seconds (SC-001)
- Handle CSV streaming for large source files (500k+ emails) without loading entire file into memory
- Tracking file lookups in reasonable time (<1s for 100k tracked entries)

**Constraints**: 
- Must preserve exact CSV format from source (headers and quoting)
- Must not corrupt multi-line email content in CSV output
- Must handle concurrent-safe file operations for tracking file
- 100% duplicate prevention accuracy when tracking file is intact (SC-002)

**Scale/Scope**: 
- Source file: ~500,000 emails
- Expected tracking file growth: up to 500k identifiers over time
- Typical sample sizes: 10-10,000 emails per extraction
- No database dependency - file-based operation only

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Pre-Phase 0)

#### Principle I: Specification-First Development
✅ **PASS** - Complete specification exists at `specs/003-random-email-extractor/spec.md` with:
- 3 prioritized user stories (P1, P2) with Given-When-Then acceptance scenarios
- 17 functional requirements (FR-001 through FR-017)
- 5 success criteria with measurable outcomes (SC-001 through SC-005)
- 7 edge cases with explicit handling strategies
- Clarifications section documenting 5 key decisions

#### Principle II: Independent User Stories
✅ **PASS** - Each user story is independently testable:
- US1 (Extract Random Sample): Can verify by running tool and checking output CSV count
- US2 (Prevent Duplicate Extraction): Can verify by running twice and comparing outputs
- US3 (Configure Sample Size): Can verify by testing with different --count values
All stories deliver standalone value and can be developed/tested independently.

#### Principle III: Test-Driven Development
✅ **PASS** - TDD will be applied:
- Unit tests for CSV parser, sampler logic, tracking file operations
- Integration tests for end-to-end extraction workflow
- Tests will be written before implementation per Phase 3 task structure
- Each functional requirement has corresponding test scenarios in spec

#### Principle IV: Documentation-Driven Design
✅ **PASS** - Design artifacts will be created in correct phases:
- `plan.md` - This file (current phase)
- `research.md` - Phase 0 (design patterns for randomization, tracking file format)
- `data-model.md` - Phase 1 (EmailRecord, TrackingRegistry, ExtractionSession)
- `quickstart.md` - Phase 1 (how to run the sampler utility)
- `tasks.md` - Phase 2 (implementation tasks by user story)

#### Principle V: Complexity Justification
✅ **PASS** - No complexity violations anticipated:
- Single command-line tool following existing cmd/ pattern
- Reuses existing CSV parsing from internal/loader
- Simple file-based tracking (no database, no distributed systems)
- Standard library dependencies only
- No abstraction layers beyond single package

#### Principle VI: Measurable Success Criteria
✅ **PASS** - All success criteria are measurable and technology-agnostic:
- SC-001: Performance target of <10s for 10k emails (measurable via time command)
- SC-002: 100% uniqueness across extractions (measurable by comparing file identifiers)
- SC-003: Output format compatibility (measurable by using output as loader input)
- SC-004: No manual intervention needed (measurable by sequential executions)
- SC-005: 100% error handling coverage (measurable via test scenarios)

---

### Re-Check After Phase 1 Design

#### Principle I: Specification-First Development
✅ **PASS** - Design documents created without code:
- `research.md` completed with 8 research decisions documented
- `data-model.md` completed with 3 entities defined (EmailRecord, TrackingRegistry, ExtractionSession)
- `quickstart.md` completed with usage examples and troubleshooting
- No implementation code written yet

#### Principle II: Independent User Stories
✅ **PASS** - Data model supports independent implementation:
- EmailRecord entity enables US1 (basic extraction)
- TrackingRegistry entity enables US2 (duplicate prevention) independently
- ExtractionSession ties them together but each can be developed separately
- No hidden dependencies between user stories in design

#### Principle III: Test-Driven Development
✅ **PASS** - Design enables TDD approach:
- Each entity has clear validation rules for test assertions
- Data model defines state transitions testable at each step
- Error conditions specified for test coverage
- Ready for test-first implementation in Phase 3

#### Principle IV: Documentation-Driven Design
✅ **PASS** - All Phase 1 artifacts complete:
- ✅ `research.md` - 8 research decisions with alternatives documented
- ✅ `data-model.md` - 3 entities with relationships and constraints
- ✅ `quickstart.md` - User guide with examples and troubleshooting
- ✅ Agent context updated (copilot-instructions.md)
- Ready for Phase 2 (tasks.md creation)

#### Principle V: Complexity Justification
✅ **PASS** - Design maintains simplicity:
- Three simple entities (EmailRecord, TrackingRegistry, ExtractionSession)
- File-based persistence (no ORM, no schema migrations)
- Reuses existing loader.ParseCSV() (no duplicate logic)
- Standard library only (confirmed in research.md)
- No complexity violations introduced

#### Principle VI: Measurable Success Criteria
✅ **PASS** - Design supports all success criteria:
- Two-pass approach enables SC-001 performance target (<10s)
- Set-based tracking ensures SC-002 (100% uniqueness)
- CSV writer with same config ensures SC-003 (format compatibility)
- Automated tracking append ensures SC-004 (no manual intervention)
- Error handling patterns documented for SC-005 (100% coverage)

**Phase 1 Gate Status**: ✅ **APPROVED** - Ready to proceed to Phase 2 (Task Breakdown)

## Project Structure

### Documentation (this feature)

```text
specs/003-random-email-extractor/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file (/speckit.plan output)
├── research.md          # Phase 0 output (randomization patterns, tracking approaches)
├── data-model.md        # Phase 1 output (EmailRecord, TrackingRegistry entities)
├── quickstart.md        # Phase 1 output (usage examples and CLI reference)
└── checklists/
    └── requirements.md  # Specification quality validation (complete)
```

### Source Code (repository root)

```text
cmd/
├── sampler/                # NEW: Random email extraction utility
│   └── main.go            # CLI entry point, flag parsing, orchestration
│
├── loader/                # EXISTING: Reference for CSV handling patterns
│   └── main.go
└── [other commands...]

internal/
├── loader/                # EXISTING: Reuse CSV parser
│   ├── parser.go          # REUSE: ParseCSV() for reading source emails.csv
│   └── parser_test.go     # REFERENCE: CSV parsing test patterns
│
├── sampler/               # NEW: Core sampling logic
│   ├── sampler.go         # Random selection, filtering by tracking
│   ├── sampler_test.go    # Unit tests for sampling logic
│   ├── tracker.go         # Tracking file read/write operations
│   ├── tracker_test.go    # Unit tests for tracking operations
│   ├── writer.go          # CSV output with timestamp naming
│   └── writer_test.go     # Unit tests for CSV writing
│
└── [other packages...]

assets/
└── enron-emails/
    ├── emails.csv                        # EXISTING: Source data (~500k emails)
    ├── emails-test-10.csv                # EXISTING: Test dataset
    ├── sampled-emails-*.csv              # NEW: Generated sample outputs
    └── extracted-*.txt                   # NEW: Tracking files (one per extraction, one ID per line)

tests/
├── fixtures/
│   ├── sample-small.csv                 # NEW: 100 emails for unit tests
│   └── sample-tracking.csv              # NEW: Test tracking file operations
│
└── integration/
    └── sampler_test.go                  # NEW: End-to-end sampler tests
```

**Structure Decision**: Following the existing single-project Go monorepo structure with clear separation between command entry point (`cmd/sampler/`) and reusable business logic (`internal/sampler/`). This aligns with the established patterns in `cmd/loader/` and `internal/loader/` while maintaining independence for the new sampling functionality.

## Complexity Tracking

*No violations - this section intentionally empty.*

All design decisions follow simplicity principles:
- Single command-line tool (no new services or architecture layers)
- Reuses existing CSV parser from `internal/loader/parser.go`
- Standard library dependencies only (no external packages)
- File-based tracking (no database or distributed coordination)
- Direct implementation without abstraction layers beyond single package

No complexity justifications required.
