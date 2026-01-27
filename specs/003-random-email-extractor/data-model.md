# Data Model: Random Email Extractor

**Feature**: 003-random-email-extractor  
**Phase**: 1 - Design  
**Date**: 2026-01-27

## Overview

The Random Email Extractor operates on three key entities: EmailRecord (source data), TrackingRegistry (state persistence), and ExtractionSession (runtime state). This document defines these entities and their relationships without implementation details.

## Entities

### EmailRecord

**Purpose**: Represents a single email entry from the source CSV file

**Attributes**:
| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| File | string | Unique identifier from "file" column | Required, unique in static source |
| Message | string | Full email content from "message" column | Required, may span multiple lines |

**Relationships**:
- One EmailRecord may appear in zero or one ExtractionSession (extracted or not)
- One EmailRecord corresponds to zero or one TrackingRegistry entry (tracked or available)

**Invariants**:
- File attribute is unique across all EmailRecords in source CSV (verified in static dataset)
- Message preserves exact formatting including headers and multi-line content

**Validation Rules**:
- File must not be empty
- Message must not be empty
- Source CSV must contain header row with columns: "file", "message"

**State Transitions**:
```
[Available] → [Selected] → [Extracted] → [Tracked]
     ↑                                       ↓
     └────────── [Persisted] ←──────────────┘
```

**Notes**:
- EmailRecord is read-only during extraction
- Multi-line email content is properly quoted in CSV format
- File identifier serves as foreign key for tracking

---

### TrackingRegistry

**Purpose**: Persistent record of all previously extracted email identifiers across all extraction sessions

**Attributes**:
| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| ExtractedIDs | collection of strings | Set of email file identifiers | Unique strings, no duplicates |

**Storage Format**: One plain text file per extraction session (`assets/enron-emails/extracted-YYYYMMDD-HHMMSS.txt`)
- One identifier per line
- Filename matches corresponding output CSV timestamp
- All tracking files are read at session start to build complete extraction history
- Individual files can be deleted to allow re-extraction of specific samples

**Relationships**:
- TrackingRegistry contains zero or more references to EmailRecord.File values
- One TrackingRegistry entry prevents one EmailRecord from being selected

**Operations**:
- **Load**: Read all tracking files matching pattern `extracted-*.txt` and aggregate IDs into memory set (for O(1) lookup)
- **Check**: Test if given EmailRecord.File exists in aggregated registry
- **Create**: Write newly extracted IDs to new timestamped tracking file
- **Verify**: Ensure no duplicate IDs across all tracking files

**Invariants**:
- ExtractedIDs contains only unique values (no duplicates)
- All IDs in ExtractedIDs must have existed in source CSV at some point (though source may change)
- Registry persists across multiple ExtractionSessions

**State Transitions**:
```
[Empty Registry] → [Load all files from disk] → [Aggregated in-memory set]
                                                      ↓
[New IDs extracted] → [Create new tracking file] → [Persisted to disk]
```

**Error Handling**:
- **Corrupted tracking file**: Skip corrupted file, log warning, continue with other files
- **Missing tracking files**: Create first tracking file (first run case)
- **Duplicate entries**: Log warning but continue (tolerate manual edits across files)

**Notes**:
- Registry grows monotonically (IDs only added via new files, never removed automatically)
- User can delete specific tracking files to "reset" those email batches
- File-per-extraction format enables selective re-extraction and auditing
- File format chosen for human readability and Unix tool compatibility

---

### ExtractionSession

**Purpose**: Represents a single execution of the sampler utility with runtime state

**Attributes**:
| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| RequestedCount | integer | Number of emails requested via --count flag | Must be > 0 |
| AvailableCount | integer | Count of emails not in TrackingRegistry | Computed during first pass |
| ExtractedCount | integer | Actual number of emails extracted | 0 ≤ ExtractedCount ≤ min(RequestedCount, AvailableCount) |
| SelectedIndices | collection of integers | Random indices to extract | Sorted for sequential access |
| OutputPath | string | Full path to generated CSV file | Format: assets/enron-emails/sampled-emails-YYYYMMDD-HHMMSS.csv |
| Timestamp | datetime | When extraction started | Used for output filename |

**Relationships**:
- One ExtractionSession produces one output CSV file
- One ExtractionSession selects zero or more EmailRecords
- One ExtractionSession appends zero or more IDs to TrackingRegistry

**Lifecycle**:
```
1. Initialize: Parse CLI flags, set RequestedCount, Timestamp
2. Load Registry: Read all tracking files matching extracted-*.txt pattern
3. Count Available: First CSV pass, compute AvailableCount
4. Generate Indices: Randomly select SelectedIndices
5. Extract Emails: Second CSV pass, write to OutputPath
6. Create Tracking File: Write extracted IDs to new extracted-YYYYMMDD-HHMMSS.txt
7. Report: Log ExtractedCount and OutputPath
```

**State Transitions**:
```
[Created] → [Registry Loaded] → [Counted] → [Indices Generated] → 
[Extracting] → [Complete] → [Registry Updated]
```

**Constraints**:
- ExtractedCount ≤ RequestedCount (can't extract more than requested)
- ExtractedCount ≤ AvailableCount (can't extract more than available)
- SelectedIndices.length == ExtractedCount (exact match)
- All SelectedIndices < AvailableCount (valid indices)

**Validation Rules**:
- RequestedCount must be positive integer
- Output directory must be writable
- Source CSV must exist and be readable
- Sufficient disk space for output file

**Error Conditions**:
- AvailableCount == 0 → Log warning, extract 0 emails, exit cleanly
- AvailableCount < RequestedCount → Log notification, extract AvailableCount emails
- Source file missing → Fatal error, do not create session
- Corrupted source CSV row → Skip row, log warning, continue

**Notes**:
- Session is ephemeral (not persisted)
- Verbose logging tracks progress through lifecycle states
- Exit code indicates success/failure for scripting

---

## Entity Relationships

```
EmailRecord (source CSV)
    ↓ 1:0..1 (may be tracked)
TrackingRegistry (persistent)
    ↓ 0..* (prevents selection)
ExtractionSession (runtime)
    ↓ produces
Output CSV File
```

**Key Relationships**:

1. **EmailRecord → TrackingRegistry**:
   - One EmailRecord.File may appear in TrackingRegistry.ExtractedIDs
   - Relationship determines if email is available for selection

2. **TrackingRegistry → ExtractionSession**:
   - ExtractionSession queries TrackingRegistry to filter available emails
   - ExtractionSession appends to TrackingRegistry after extraction

3. **EmailRecord → ExtractionSession**:
   - ExtractionSession selects subset of EmailRecords not in TrackingRegistry
   - Selected EmailRecords are written to output CSV

## Data Flow

### Input Flow
```
Source CSV (emails.csv)
    ↓ stream
EmailRecord instances
    ↓ filter by
TrackingRegistry (extracted-*.txt)
    ↓ random selection
ExtractionSession.SelectedIndices
```

### Output Flow
```
ExtractionSession.SelectedIndices
    ↓ extract corresponding
EmailRecord instances
    ↓ write to
Output CSV (sampled-emails-YYYYMMDD-HHMMSS.csv)
    ↓ update
TrackingRegistry (append IDs)
```

### State Persistence

**Persistent State** (survives sessions):
- TrackingRegistry → `assets/enron-emails/extracted-*.txt` (one file per extraction)
- Output CSVs → `assets/enron-emails/sampled-emails-*.csv`

**Ephemeral State** (runtime only):
- ExtractionSession (entire entity)
- In-memory EmailRecord instances
- In-memory TrackingRegistry set (loaded from disk)

## Validation and Integrity

### Source Data Validation (FR-007)
- Assume File attribute uniqueness (verified in static source dataset)
- Fail fast if duplicates detected in source CSV
- Log duplicate identifiers for investigation

### Tracking Registry Integrity
- Prevent duplicate IDs within tracking file
- Handle corrupted tracking file gracefully (start fresh)
- Tolerate manual edits (sorted order not required)

### Output Format Validation (FR-010)
- Preserve exact CSV structure from source
- Maintain header row: "file,message"
- Properly quote multi-line Message content
- Output must be parsable by loader.ParseCSV()

## Scalability Considerations

### Memory Footprint
- TrackingRegistry: O(n) where n = number of tracked emails (~100 bytes/entry)
- ExtractionSession.SelectedIndices: O(m) where m = requested count (~8 bytes/index)
- EmailRecord streaming: O(1) constant memory (not loaded all at once)

### Performance Characteristics
- Source CSV reading: Linear time O(n) for n emails (streaming)
- Tracking lookup: Constant time O(1) per email (set-based)
- Random selection: O(m log m) for sorting m indices
- Output writing: Linear time O(m) for m selected emails

### Growth Patterns
- TrackingRegistry grows monotonically up to source CSV size
- Each ExtractionSession adds at most RequestedCount entries to registry
- After all emails extracted once, AvailableCount → 0 (steady state)

## Summary

The data model defines three clean entities with clear responsibilities:

1. **EmailRecord**: Immutable source data representation
2. **TrackingRegistry**: Persistent extraction history
3. **ExtractionSession**: Ephemeral runtime state

All entities are file-based (no database), maintain simple relationships, and support the required operations for random sampling with duplicate prevention. The model satisfies all functional requirements (FR-001 through FR-017) and success criteria (SC-001 through SC-005).
