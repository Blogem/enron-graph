# Tasks: Random Email Extractor

**Feature**: 003-random-email-extractor  
**Input**: Design documents from `/specs/003-random-email-extractor/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic Go package structure

- [X] T001 Create cmd/sampler/main.go with basic CLI flag parsing (--count flag)
- [X] T002 Create internal/sampler/types.go with EmailRecord and ExtractionSession structs
- [X] T003 [P] Create tests/fixtures/sample-small.csv with 10 test emails for unit tests
- [X] T004 [P] Add entry to README.md documenting new sampler utility

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core CSV parsing and tracking file infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

**TDD Workflow**: Write tests first, then implement to pass tests

- [X] T005 [P] Create internal/sampler/tracker_test.go with unit tests for LoadTracking()
- [X] T006 Create internal/sampler/tracker.go with LoadTracking() function to read all extracted-*.txt files
- [X] T007 [P] Create internal/sampler/writer_test.go with unit tests for CSV output format preservation
- [X] T008 Create internal/sampler/writer.go with WriteCSV() function using encoding/csv

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Extract Random Sample (Priority: P1) 🎯 MVP

**Goal**: Enable developers to extract a specified number of random emails from the source CSV into a new timestamped output file

**Independent Test**: Run `go run cmd/sampler/main.go --count 10` and verify that a new CSV file is created with exactly 10 emails

**TDD Workflow**: Write failing tests first, then implement minimum code to pass tests

### Tests for User Story 1 (Write These First)

- [X] T009 [P] [US1] Create internal/sampler/parser_test.go with tests for ParseCSV() wrapper
- [X] T010 [P] [US1] Create internal/sampler/sampler_test.go with tests for CountAvailable() function
- [X] T011 [P] [US1] Add tests to internal/sampler/sampler_test.go for GenerateIndices() randomization logic
- [X] T012 [P] [US1] Add tests to internal/sampler/sampler_test.go for ExtractEmails() function
- [X] T013 [US1] Create tests/integration/sampler_basic_test.go for end-to-end extraction workflow

### Implementation for User Story 1 (Make Tests Pass)

- [X] T014 [US1] Implement ParseCSV() wrapper in internal/sampler/parser.go that calls loader.ParseCSV()
- [X] T015 [US1] Implement CountAvailable() function in internal/sampler/sampler.go for first-pass counting
- [X] T016 [US1] Implement GenerateIndices() function in internal/sampler/sampler.go for random selection using math/rand
- [X] T017 [US1] Implement ExtractEmails() function in internal/sampler/sampler.go for second-pass extraction
- [X] T018 [US1] Wire up main.go to call counting, selection, and extraction functions
- [X] T019 [US1] Add progress logging to main.go showing "Processing N/M emails..." every 100 records
- [X] T020 [US1] Implement timestamp-based output filename generation (sampled-emails-YYYYMMDD-HHMMSS.csv)
- [X] T021 [US1] Add error handling for missing source CSV file with clear error message
- [X] T022 [US1] Create tests/test_us1_extraction.sh that tests basic extraction flow: runs sampler with --count 10, verifies output file exists, verifies exactly 10 emails extracted, verifies CSV format preserved

**Checkpoint**: At this point, User Story 1 should be fully functional - can extract random samples and create output CSV

---

## Phase 4: User Story 2 - Prevent Duplicate Extraction (Priority: P2)

**Goal**: Ensure that previously extracted emails are never selected again across multiple runs

**Independent Test**: Run sampler twice with `--count 5`, then verify all 10 emails are unique and tracking files are created

**TDD Workflow**: Write tests for duplicate prevention, then implement tracking functionality

### Tests for User Story 2 (Write These First)

- [X] T023 [P] [US2] Add tests to internal/sampler/tracker_test.go for CreateTrackingFile() function
- [X] T024 [P] [US2] Add tests to internal/sampler/sampler_test.go for CountAvailable() with exclusion logic
- [X] T025 [P] [US2] Add tests to internal/sampler/sampler_test.go for ExtractEmails() with duplicate filtering
- [X] T026 [US2] Create tests/integration/sampler_duplicate_test.go for multi-run uniqueness verification
- [X] T027 [US2] Add tests to internal/sampler/tracker_test.go for corrupted tracking file handling

### Implementation for User Story 2 (Make Tests Pass)

- [X] T028 [US2] Implement CreateTrackingFile() function in internal/sampler/tracker.go to write extracted-YYYYMMDD-HHMMSS.txt
- [X] T029 [US2] Update CountAvailable() to exclude emails that exist in loaded tracking set
- [X] T030 [US2] Update ExtractEmails() to skip emails found in tracking set during extraction
- [X] T031 [US2] Wire up main.go to call CreateTrackingFile() after successful extraction
- [X] T032 [US2] Add logging to show "Found N previously extracted emails (from M tracking files)"
- [X] T033 [US2] Add handling for corrupted tracking files (skip file, log warning, continue)
- [X] T034 [US2] Create tests/test_us2_duplicates.sh that tests duplicate prevention: runs sampler twice with --count 5, verifies 10 unique emails across both files, verifies tracking files created, verifies no duplicates between runs

---

## Phase 5: User Story 3 - Configure Sample Size (Priority: P1)

**Goal**: Allow developers to specify different sample sizes for different testing scenarios

**Independent Test**: Run with various counts (10, 100, 1000) and verify each produces exact count requested

**TDD Workflow**: Write tests for edge cases and validation, then implement configuration logic

### Tests for User Story 3 (Write These First)

- [X] T035 [P] [US3] Create cmd/sampler/main_test.go with tests for --count flag validation
- [X] T036 [P] [US3] Add tests to internal/sampler/sampler_test.go for count exceeding available emails
- [X] T037 [P] [US3] Add tests for edge case: zero emails available
- [X] T038 [US3] Create tests/integration/sampler_config_test.go for various count scenarios (10, 100, 1000)

### Implementation for User Story 3 (Make Tests Pass)

- [X] T039 [US3] Add validation in main.go for --count flag (must be positive integer)
- [X] T040 [US3] Implement edge case handling when requested count exceeds available emails
- [X] T041 [US3] Add warning message "Only N emails available, extracting all remaining"
- [X] T042 [US3] Add --help flag with usage documentation
- [X] T043 [US3] Add final summary logging: "Successfully extracted N emails to [path]"
- [X] T044 [US3] Create tests/test_us3_configuration.sh that tests various count scenarios: runs sampler with --count 10, 100, 1000, verifies exact counts, tests --help flag, tests invalid count values, verifies appropriate error messages

**Checkpoint**: At this point, User Story 3 should be fully functional - can handle various sample sizes and edge cases

---

## Phase 6: Polish & Edge Cases

**Purpose**: Comprehensive edge case handling and performance optimization

**TDD Workflow**: Add comprehensive tests for edge cases, then handle them in implementation

### Additional Tests (Write These First)

- [X] T046 [P] Add test to internal/sampler/parser_test.go for edge case: source file missing (should fail gracefully)
- [X] T047 [P] Add test to internal/sampler/parser_test.go for edge case: corrupted CSV row (should skip and continue)
- [X] T048 [P] Add test to internal/sampler/sampler_test.go for edge case: all emails already extracted (should extract 0)
- [X] T049 [P] Add test to internal/sampler/writer_test.go for disk space issues when writing output
- [X] T050 [P] Create tests/integration/sampler_performance_test.go: verify 10k emails extracted in <10 seconds (SC-001)
- [X] T051 [P] Add test to internal/sampler/parser_test.go for CSV format compatibility with loader.ParseCSV() (SC-003)

### Implementation & Refinements (Make Tests Pass)

- [X] T052 Add error handling for source file missing with clear error message
- [X] T053 Add error handling for corrupted CSV rows (skip and continue with warning)
- [X] T054 Add handling for all emails already extracted scenario
- [X] T055 Add error handling for disk space issues when writing output
- [X] T056 Verify quickstart.md examples work as documented
- [X] T057 Optimize extraction performance to meet <10s requirement for 10k emails

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-5)**: All depend on Foundational phase completion
  - US1 (Phase 3): Can start after Phase 2 - No dependencies on other stories
  - US2 (Phase 4): Can start after Phase 2 - Independent from US1
  - US3 (Phase 5): Can start after Phase 2 - Independent from US1 and US2
- **Polish (Phase 6)**: Depends on all user stories for comprehensive edge case testing

### TDD Order

- **US1**: Tests First (T009-T013) → Implementation (T014-T022)
- **US2**: Tests First (T023-T027) → Implementation (T028-T034)
- **US3**: Tests First (T035-T038) → Implementation (T039-T044)
- **Polish**: Tests First (T046-T051) → Implementation (T052-T057)

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (T003, T004)
- All Foundational test tasks marked [P] can run in parallel (T005, T007)
- Within each user story, all test creation tasks marked [P] can run in parallel
- All Polish test tasks marked [P] can run in parallel (T046-T051)

### Sequential Dependencies Within User Stories

- **US1**: Parser → Counting → Random Selection → Extraction → Output → Error Handling
- **US2**: Tracking File Write → Filter During Count → Filter During Extract → Error Handling
- **US3**: Validation → Edge Cases → Documentation → Summary

---

## Parallel Example: Phase 2 (Foundational - TDD)

```bash
# Write tests in parallel first:
Task: "Create internal/sampler/tracker_test.go with unit tests for LoadTracking()"
Task: "Create internal/sampler/writer_test.go with unit tests for CSV output format"
# Then implement in parallel:
Task: "Create internal/sampler/tracker.go with LoadTracking() function"
Task: "Create internal/sampler/writer.go with WriteCSV() function"
```

## Parallel Example: Phase 3 (User Story 1 - TDD)

```bash
# Write all tests in parallel first:
Task: "Create internal/sampler/parser_test.go with tests for ParseCSV()"
Task: "Create internal/sampler/sampler_test.go with tests for CountAvailable()"
Task: "Add tests to sampler_test.go for GenerateIndices()"
Task: "Add tests to sampler_test.go for ExtractEmails()"
# Then implement sequentially (dependencies exist):
Task: "Implement ParseCSV() → CountAvailable() → GenerateIndices() → ExtractEmails()"
```

## Parallel Example: Phase 6 (Polish - TDD)

```bash
# Write all edge case tests in parallel:
Task: "Add test for source file missing"
Task: "Add test for corrupted CSV row"
Task: "Add test for all emails already extracted"
Task: "Create performance test for 10k emails"
Task: "Add test for CSV format compatibility"
# Then fix issues found by tests
```

---

## Implementation Strategy
TDD Strategy (Recommended)

Following Test-Driven Development principles:

1. **Red-Green-Refactor Cycle**: 
   - Write failing test first (Red)
   - Write minimum code to pass test (Green)
   - Refactor if needed
2. **Within Each Phase**:
   - Complete all test tasks BEFORE implementation tasks
   - Run tests frequently to verify they fail initially
   - Implement to make tests pass
3. **Complete phases in order**: 1 → 2 → 3 → 4 → 5 → 6
4. **Verify after each phase**: All tests should pass before moving to next phase
5. **Commit working increments**: After each user story phase completionon with `go run cmd/sampler/main.go --count 100`
5. Verify output CSV is valid and contains 100 unique emails

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Basic sampler works! (MVP!)
3. Add User Story 2 → Test independently → Now prevents duplicates!
4. Add User Story 3 → Test independently → Now handles all edge cases!
5. Each story adds value without breaking previous stories

### Sequential Strategy (Recommended)

Since this is a single-developer utility tool:

1. Complete all phases in order (1 → 2 → 3 → 4 → 5 → 6)
2. Test after each user story phase
3. Commit working increments frequently

---

## Notes

- File paths follow existing project structure (cmd/, internal/, tests/)
- Reuses existing loader.ParseCSV() from internal/loader/parser.go
- Standard library only (no external dependencies)
- [P] tasks can run in parallel (different files, no dependencies)
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Tests use existing test framework patterns from internal/loader/parser_test.go
