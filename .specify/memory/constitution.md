<!--
SYNC IMPACT REPORT
==================
Version Change: [NONE] → 1.0.0
Change Type: Initial constitution establishment
Modified Principles: N/A (initial version)
Added Sections: All core sections established
Removed Sections: N/A

Templates Status:
✅ plan-template.md - Constitution Check section aligns with principles
✅ spec-template.md - User story structure supports Spec-First and Independent Testing
✅ tasks-template.md - Phase structure supports TDD and story-based implementation
⚠️  Command files - Generic structure, no agent-specific references found

Follow-up TODOs: None
==================
-->

# Enron Graph Constitution

## Core Principles

### I. Specification-First Development

Every feature MUST begin with a complete specification document before any implementation code is written.
Specifications MUST include:
- User scenarios with acceptance criteria in Given-When-Then format
- Functional requirements (FR-XXX identifiers)
- Success criteria with measurable outcomes
- Edge cases and boundary conditions

**Rationale**: Clear specifications prevent scope creep, enable accurate estimation, provide testable
contracts, and ensure alignment between stakeholders before costly implementation begins.

### II. Independent User Stories

User stories MUST be independently testable and deliverable. Each story represents a complete vertical
slice of functionality that:
- Can be developed independently of other stories
- Can be tested independently
- Delivers standalone user value
- Has assigned priority (P1, P2, P3, etc.)

**Rationale**: Independent stories enable parallel development, incremental delivery, risk mitigation
through MVP identification, and flexibility to reprioritize based on feedback.

### III. Test-Driven Development (NON-NEGOTIABLE)

TDD is mandatory for all implementation work. The Red-Green-Refactor cycle MUST be strictly followed:
1. Write failing tests first (contract tests, integration tests as applicable)
2. Obtain user approval on test scenarios
3. Verify tests fail for the right reasons
4. Implement minimum code to make tests pass
5. Refactor while keeping tests green

Tests are OPTIONAL only when explicitly excluded from the specification.

**Rationale**: TDD ensures code correctness, provides living documentation, enables fearless refactoring,
catches regressions early, and enforces clear thinking about requirements before implementation.

### IV. Documentation-Driven Design

Before implementation, design artifacts MUST be created and reviewed:
- `plan.md` - Technical context, constitution check, project structure
- `research.md` - Technical research findings (Phase 0)
- `data-model.md` - Entity definitions and relationships (Phase 1)
- `contracts/` - API contracts and interface definitions (Phase 1)
- `tasks.md` - Granular implementation tasks organized by user story (Phase 2)

**Rationale**: Upfront design documents reduce implementation errors, enable parallel work planning,
provide reference material for future maintainers, and create checkpoints for stakeholder alignment.

### V. Complexity Justification

Any violation of simplicity principles MUST be explicitly justified in the plan's Complexity Tracking
table. Justifications MUST include:
- What violation is being introduced (e.g., additional abstraction layer, fourth service)
- Why the complexity is necessary for the current requirement
- What simpler alternative was considered and why it was rejected

**Rationale**: Forcing explicit justification prevents premature abstraction, maintains YAGNI discipline,
creates an audit trail for architectural decisions, and enables future refactoring when complexity is
no longer justified.

### VI. Measurable Success Criteria

Every feature MUST define success criteria (SC-XXX identifiers) that are:
- Technology-agnostic (describe outcomes, not implementation)
- Measurable (quantifiable metrics or observable behaviors)
- Testable (can be verified through tests or monitoring)
- User-focused (describe user or business value)

**Rationale**: Measurable criteria prevent feature bloat, enable objective completion verification,
align technical work with business value, and provide clear targets for implementation quality.

## Development Workflow

### Feature Lifecycle

Features MUST progress through the following phases:

1. **Specification** (`/speckit.specify`) - Create spec.md with user stories, requirements, and success criteria
2. **Analysis** (`/speckit.analyze`) - Extract technical context and identify unknowns
3. **Planning** (`/speckit.plan`) - Generate plan.md, research.md, data-model.md, contracts/
4. **Task Breakdown** (`/speckit.tasks`) - Generate tasks.md organized by user story
5. **Implementation** (`/speckit.implement`) - Execute tasks following TDD cycle
6. **Verification** - Validate against success criteria and constitution compliance

### Branch and Documentation Structure

Each feature MUST:
- Use branch naming: `[###-feature-name]` where ### is the feature number
- Maintain documentation in: `specs/[###-feature-name]/`
- Reference the constitution check gates in plan.md before proceeding to implementation

### Constitution Compliance

All implementation work MUST verify compliance with constitution principles:
- Specifications complete before code
- User stories independently testable
- Tests written before implementation (when tests are included)
- Design documents reviewed before task execution
- Complexity violations explicitly justified
- Success criteria defined and measurable

## Quality Gates

### Pre-Implementation Gates

MUST pass before Phase 0 research begins (re-checked after Phase 1 design):
- ✅ Specification document complete with all required sections
- ✅ User stories have assigned priorities
- ✅ Functional requirements identified (FR-XXX)
- ✅ Success criteria defined (SC-XXX)
- ✅ Edge cases documented

### Implementation Gates

MUST pass before marking user story complete:
- ✅ Tests written first and failed appropriately (when tests included)
- ✅ Implementation makes tests pass
- ✅ Code passes linting and formatting checks
- ✅ Integration tests pass (when applicable)
- ✅ Success criteria verified

## Governance

This constitution supersedes all other development practices and conventions. Any deviation MUST be
explicitly documented in the Complexity Tracking section of the implementation plan with full
justification.

### Amendment Process

Constitution amendments require:
1. Proposal documenting the change rationale
2. Impact analysis on existing templates and workflows
3. Version update following semantic versioning:
   - MAJOR: Breaking changes to core principles or removal of sections
   - MINOR: Addition of new principles or substantial expansions
   - PATCH: Clarifications, wording improvements, non-semantic refinements
4. Update to Sync Impact Report at the top of this file
5. Propagation of changes to affected templates and command files

### Versioning Policy

This constitution follows semantic versioning (MAJOR.MINOR.PATCH):
- MAJOR version: Backward-incompatible governance changes, principle removals/redefinitions
- MINOR version: New principles added, sections materially expanded
- PATCH version: Clarifications, typo fixes, wording improvements

### Compliance Review

Constitution compliance MUST be verified:
- During plan.md generation (Constitution Check section)
- Before task execution begins
- During code review of implementation PRs
- In retrospectives when complexity violations were justified

**Version**: 1.0.0 | **Ratified**: 2026-01-24 | **Last Amended**: 2026-01-24
