# Feature Specification: Random Email Extractor

**Feature Branch**: `003-random-email-extractor`  
**Created**: January 27, 2026  
**Status**: Draft  
**Input**: User description: "I want to add a new feature to solution. I want to have a small utility that can extract random emails from the big emails csv in the assets folder. The number I want to extract should be configurable (e.g. 1000) and put it in a new csv with the same format. Next to that csv we should have a file that tracks which emails have already been extracted (use some unique identifier), so that the next time the tool runs, these emails are not extracted again"

## Clarifications

### Session 2026-01-27

- Q: When the utility extracts emails and creates an output CSV file, what naming convention should be used? → A: Timestamp-based naming (e.g., "sampled-emails-20260127-143022.csv")
- Q: What format should the tracking file use to store extracted email identifiers? → A: Simple text file with one identifier per line, timestamped per extraction (e.g., "extracted-20260127-143022.txt")
- Q: What level of logging/output should the utility provide during normal operation? → A: Verbose logging showing progress (e.g., "Processed 500/1000 emails...")
- Q: How should the sample size be specified when running the utility? → A: Command-line flag (e.g., --count 1000 or -n 1000)
- Q: Where should the tracking file and output CSV files be stored? → A: Fixed location in assets/enron-emails/ directory

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Extract Random Sample (Priority: P1)

A developer needs to create a smaller test dataset from the large email corpus for development and testing purposes. They run the utility tool with a specified sample size to extract random emails into a new CSV file.

**Why this priority**: This is the core functionality that delivers immediate value - enabling developers to work with manageable subsets of data for testing and development without loading the entire dataset.

**Independent Test**: Can be fully tested by running the tool with a sample size parameter and verifying that a new CSV file is created with the correct number of randomly selected emails in the same format as the source.

**Acceptance Scenarios**:

1. **Given** a source CSV file with 500,000 emails, **When** a developer runs the utility with sample size 1000, **Then** a new CSV file is created containing exactly 1000 randomly selected emails with the same column structure (file, message)
2. **Given** the utility has not been run before, **When** a developer requests 500 emails, **Then** the output CSV contains 500 unique emails and a tracking file is created recording these extracted emails
3. **Given** a valid sample size parameter, **When** the extraction completes, **Then** the user receives confirmation of the number of emails extracted and the output file location

---

### User Story 2 - Prevent Duplicate Extraction (Priority: P2)

A developer runs the extraction utility multiple times to build up a larger test dataset. The utility ensures that previously extracted emails are never selected again, maintaining dataset uniqueness across multiple runs.

**Why this priority**: This enables incremental dataset building and ensures test data diversity, which is critical for comprehensive testing scenarios.

**Independent Test**: Can be tested by running the utility twice and verifying that the second run produces completely different emails than the first run, with the tracking file correctly recording all extracted email identifiers.

**Acceptance Scenarios**:

1. **Given** 500 emails were previously extracted and tracked, **When** the utility runs again requesting 300 emails, **Then** all 300 new emails are different from the previously extracted 500
2. **Given** a tracking file exists with extracted email identifiers, **When** the utility starts, **Then** it reads the tracking file and excludes those emails from the selection pool
3. **Given** multiple extraction runs over time, **When** examining all output CSV files, **Then** no email appears in more than one output file

---

### User Story 3 - Configure Sample Size (Priority: P1)

A developer needs different dataset sizes for different testing scenarios (unit tests vs integration tests). They specify the desired number of emails as a parameter when running the utility.

**Why this priority**: Flexibility in sample size is essential for the tool to be useful across different testing contexts and workflows.

**Independent Test**: Can be tested by running the utility with various sample size values (10, 100, 1000) and verifying that each run produces exactly the requested number of emails.

**Acceptance Scenarios**:

1. **Given** the utility accepts a sample size parameter, **When** a developer specifies 1000, **Then** exactly 1000 emails are extracted
2. **Given** different testing needs, **When** a developer runs the utility with size 50 for unit tests and size 5000 for performance tests, **Then** each run produces the correct number of emails
3. **Given** a sample size parameter is provided, **When** the available pool (after excluding tracked emails) is larger than the requested size, **Then** the extraction succeeds with the exact count requested

---

### Edge Cases

- **Requested sample size exceeds available emails**: Extract all remaining unextracted emails and notify the user of the actual count extracted
- **Corrupted or incomplete CSV rows**: Skip invalid rows, log warnings, and continue processing valid rows
- **Corrupted or deleted tracking file**: Ignore the corrupted tracking file and start fresh (user can manually preserve old tracking data if needed)
- **Empty or missing source CSV file**: Return clear error message and exit gracefully
- **Source file smaller than tracking registry**: Check file contents on each run rather than relying on cached counts; ignore tracked identifiers that don't exist in current source
- **Concurrent executions**: Not supported; utility assumes single-user development environment with sequential runs
- **Insufficient disk space**: Return clear error message when unable to write output files

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Utility MUST read emails from the source CSV file located in the assets folder
- **FR-002**: Utility MUST accept a command-line flag parameter (e.g., --count or -n) specifying the number of emails to extract
- **FR-003**: Utility MUST randomly select emails from the source file that have not been previously extracted
- **FR-004**: Utility MUST write extracted emails to a new CSV file in the assets/enron-emails/ directory with timestamp-based naming (format: sampled-emails-YYYYMMDD-HHMMSS.csv) maintaining the same format as the source (file, message columns)
- **FR-005**: Utility MUST create a timestamped tracking file per extraction in the assets/enron-emails/ directory (format: extracted-YYYYMMDD-HHMMSS.txt, simple text format with one identifier per line) that records unique identifiers of all extracted emails in that session
- **FR-006**: Utility MUST read all tracking files (extracted-*.txt) before each extraction to determine which emails to exclude from selection
- **FR-007**: Utility MUST use the "file" column value as the unique identifier for tracking extracted emails (uniqueness verified: static dataset has unique file identifiers)
- **FR-008**: Utility MUST update the tracking file after each successful extraction to include newly extracted email identifiers
- **FR-009**: Utility MUST handle the case where requested sample size exceeds available unextracted emails by extracting all remaining available emails and notifying the user of the actual count extracted
- **FR-010**: Utility MUST preserve the CSV format exactly, including headers and field delimiters from the source file
- **FR-011**: Utility MUST ensure randomization across the entire pool of unextracted emails, not just sequential selection
- **FR-012**: Utility MUST create the tracking file if it does not exist on first run
- **FR-013**: Utility MUST validate that the source CSV file exists and is readable before attempting extraction
- **FR-014**: Utility MUST provide verbose progress logging during operation including: number requested, progress updates, number extracted, and output file location
- **FR-015**: Utility MUST skip corrupted or incomplete CSV rows, log warnings for each skipped row, and continue processing
- **FR-016**: Utility MUST handle corrupted tracking files by ignoring them and starting fresh extraction tracking
- **FR-017**: Utility MUST check source file contents on each run rather than caching email counts, ensuring tracking works even if source file changes

### Key Entities

- **Email Record**: Represents a single email entry with a unique file identifier and message content; note that email content may span multiple lines within a single CSV row (properly quoted in CSV format)
- **Extraction Session**: Represents a single execution of the utility with a specific sample size parameter; produces one output CSV file
- **Tracking Registry**: Maintains the persistent record of all extracted email identifiers across all extraction sessions; prevents duplicate selection

### Assumptions

- The source CSV file format is consistent and follows the structure: header row with "file" and "message" columns
- The "file" column contains unique values that can serve as identifiers (as evidenced by the existing data structure like "allen-p/_sent_mail/1.")
- The source CSV file remains in the same location (assets/enron-emails/emails.csv) between utility runs
- Users will run this utility from a command-line environment where parameters can be passed
- Output CSV files and tracking file will be stored in the assets/enron-emails/ directory alongside the source file
- The utility is intended for development/testing purposes, not production data processing
- Sufficient disk space is available for output files and tracking data

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can extract any specified number of emails (e.g., 1000) from the source CSV in under 10 seconds for samples up to 10,000 emails
- **SC-002**: 100% of extracted emails across all output files are unique (no duplicates) when the tracking system is functioning correctly
- **SC-003**: The output CSV file is immediately usable as a drop-in replacement for the source file with no format modifications required
- **SC-004**: Developers can successfully build test datasets incrementally over multiple runs without manual intervention to prevent duplicates
- **SC-005**: The utility correctly handles edge cases (insufficient available emails, missing files) with clear error messages in 100% of scenarios
