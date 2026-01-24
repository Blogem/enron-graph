# Clarifications for Cognitive Backbone POC

**Date**: 2026-01-24  
**Status**: ✅ Resolved - All clarifications integrated into spec.md and analysis.md

> **Note**: This document has been superseded. All technology-agnostic decisions have been integrated into `spec.md`, and all technical decisions have been integrated into `analysis.md`. This file is retained for historical reference only.

---

## Dataset Format & Schema

### Question 1: Enron Dataset Format
**Answer**: Inspected `emails.csv` - Contains 2 columns:
- `file`: Email identifier (e.g., "allen-p/_sent_mail/1.")
- `message`: Full email content including headers (Message-ID, Date, From, To, CC, BCC, Subject) and body as raw text

**Implications**:
- Need to parse email headers from the `message` field
- Message-ID is available for duplicate detection (FR-010)
- CC and BCC are within the message headers (need parsing)
- Body content is plain text embedded after headers
- Date format appears to be RFC 2822 format (e.g., "Mon, 14 May 2001 16:39:00 -0700 (PDT)")

---

## Entity Extraction

### Question 2: Confidence Thresholds
**Answer**: 
- Minimum confidence score: **0.7**
- Keep implementation simple (single threshold for all entity types)
- No special flagging for low-confidence entities in POC

**Implications**:
- Entities with confidence < 0.7 are discarded
- Simplifies storage and query logic
- May reduce total entity count but improves data quality

---

## Schema Evolution

### Question 3: Schema Promotion Approval
**Answer**:
- **Manual approval** workflow
- CLI/TUI interface is sufficient (no web UI required)
- Analyst presents candidates, user confirms promotion via terminal prompt

**Implications**:
- Simpler implementation (no approval UI needed)
- Requires interactive terminal session during promotion
- User has control over what gets promoted

---

## Visualization

### Question 4: Interface Type
**Answer**:
- **Prioritize TUI** (Bubble Tea/tview)
- Only pivot to web if certain features are impossible in TUI
- ASCII graph rendering is acceptable
- **P4 (visualization) is required** - cannot be discarded, but can be very basic

**Implications**:
- Evaluate TUI graph rendering libraries (e.g., tview with ASCII art)
- If graph rendering in TUI proves too limited, build minimal web interface
- Focus on functionality over aesthetics for POC

---

## Data Model

### Question 5: Relationship Properties
**Answer**:
- **COMMUNICATES_WITH**: Bidirectional
- **MENTIONS**: Both keyword matching AND LLM extraction
- **Add** timestamps (from email date) to all relationships
- **Add** confidence scores to all relationships
- **No** weight/strength properties needed

**Implications**:
- Relationship schema: `type`, `timestamp`, `confidence_score`
- MENTIONS extraction requires both simple pattern matching and LLM calls
- Queries can filter by timestamp and confidence

---

## Performance & Concurrency

### Question 6: Concurrent Processing
**Answer**:
- **Yes**, process emails concurrently
- Concurrency level: **10 to 100** emails in parallel
- Conflict resolution: **Last write wins** (keep simple)

**Implications**:
- Use goroutine pool with 10-100 workers
- No complex locking or conflict detection needed
- Database must handle concurrent writes (most graph DBs do)

---

## Vector Embeddings

### Question 7: Embedding Scope
**Answer**:
- Generate embeddings for **all entities**
- **Not** for email bodies
- Keep implementation simple

**Implications**:
- Storage overhead limited to entities only (not full emails)
- LLM API costs apply to entity descriptions, not full email corpus
- Enables semantic search on entities

---

## Error Handling

### Question 8: Error Tolerance
**Answer**:
- Acceptable failure rate: **2%** of emails
- **Halt** on critical errors (e.g., database connection lost)
- Logging: **Error messages only** (no stack traces)
- Keep implementation simple for POC

**Implications**:
- Need error counter/tracker during ingestion
- Fatal errors should stop processing gracefully
- Minimal logging overhead

---

## Performance Testing

### Question 9: Hardware Baseline
**Answer**:
- Development environment: **MacBook Air M4 with 24GB RAM**
- Docker available
- Use this as "standard hardware" for performance targets

**Implications**:
- SC-001 target: 10k emails in <10 minutes on M4 MacBook Air
- Can use Docker for database containerization
- Adequate resources for POC scale (100k+ nodes)

---

## Natural Language Queries

### Question 10: Query Scope
**Answer**: **Recommendation requested**

**Recommended Queries for POC** (targeting SC-012: 80% accuracy):

1. **Entity Lookup**:
   - "Who is [person name]?"
   - "Show me information about [organization name]"
   - "Find emails from [person]"

2. **Relationship Discovery**:
   - "Who did [person] email?"
   - "How are [person A] and [person B] connected?"
   - "Show me [person]'s contacts"

3. **Concept Search**:
   - "Find emails about [topic]"
   - "What topics did [person] discuss?"

4. **Simple Aggregations**:
   - "How many emails did [person] send?"
   - "Who sent the most emails?"

5. **Temporal Queries** (optional for POC):
   - "Emails from [date/month/year]"
   - "When did [person] last email [person]?"

**Rationale**:
- Covers core query patterns (lookup, relationship, concept)
- Simple enough for LLM to translate to graph queries
- Avoids complex aggregations or analytics
- Can achieve 80% accuracy with basic prompt engineering
- Total test set: 10 queries (5 basic types × 2 examples each)

**Out of Scope**:
- Complex analytics ("What was the communication pattern in Q3 2001?")
- Multi-hop reasoning ("Who knows someone who knows X?")
- Comparative queries ("Did X email more than Y?")

---

## Summary of Key Decisions

| Area | Decision | Impact |
|------|----------|--------|
| **Dataset** | 2-column CSV (file, message); parse headers from message field | Need email parser |
| **Confidence** | 0.7 threshold, single value for all types | Simple filtering |
| **Approval** | Manual CLI/TUI workflow | Interactive promotion |
| **Visualization** | TUI preferred, web if needed, basic ASCII acceptable | Evaluate TUI libs |
| **Relationships** | Bidirectional COMMUNICATES_WITH, timestamps + confidence | Richer data model |
| **Concurrency** | 10-100 goroutines, last-write-wins | Parallel processing |
| **Embeddings** | Entities only (not emails) | Moderate API costs |
| **Error Rate** | 2% acceptable, halt on critical | Graceful degradation |
| **Hardware** | M4 MacBook Air 24GB | Performance baseline |
| **NL Queries** | 10 test queries (5 basic patterns) | Focused LLM testing |

---

**Next Steps**: Proceed to planning phase with these clarifications integrated.
