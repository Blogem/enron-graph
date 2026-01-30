## ADDED Requirements

### Requirement: Configure analysis parameters
The system SHALL allow users to configure analysis parameters before running entity analysis.

#### Scenario: User sets minimum occurrences threshold
- **WHEN** user sets minimum occurrences to 5
- **THEN** only entity types appearing at least 5 times SHALL be included in results

#### Scenario: User sets consistency threshold
- **WHEN** user sets minimum consistency to 0.4 (40%)
- **THEN** only entity types with property consistency >= 40% SHALL be included in results

#### Scenario: User sets top N limit
- **WHEN** user sets top N to 10
- **THEN** system SHALL return at most 10 candidates ranked by score

#### Scenario: Default parameter values
- **WHEN** user does not specify parameters
- **THEN** system SHALL use defaults: minOccurrences=5, minConsistency=0.4, topN=10

### Requirement: Analyze discovered entities
The system SHALL analyze discovered entities to identify and rank candidates for type promotion.

#### Scenario: Successful analysis with results
- **WHEN** user triggers analysis with valid parameters
- **THEN** system SHALL return ranked candidates with frequency, density, consistency, and score metrics

#### Scenario: Analysis with no qualifying candidates
- **WHEN** no entity types meet the specified thresholds
- **THEN** system SHALL return empty results with totalTypes=0

#### Scenario: Analysis with database error
- **WHEN** database connection fails during analysis
- **THEN** system SHALL return an error indicating database failure

### Requirement: Calculate ranking metrics
The system SHALL calculate frequency, density, consistency, and score for each candidate.

#### Scenario: Calculate frequency metric
- **WHEN** analyzing an entity type
- **THEN** frequency SHALL equal the count of entities of that type

#### Scenario: Calculate density metric
- **WHEN** analyzing an entity type
- **THEN** density SHALL equal the average number of relationships per entity of that type

#### Scenario: Calculate consistency metric
- **WHEN** analyzing an entity type
- **THEN** consistency SHALL equal the average property consistency across all properties of that type

#### Scenario: Calculate composite score
- **WHEN** calculating candidate score
- **THEN** score SHALL be computed as: 0.4 × normalized_frequency + 0.3 × normalized_density + 0.3 × consistency

### Requirement: Display ranked candidates
The system SHALL display candidates in a sortable table with ranking information.

#### Scenario: Display candidates sorted by rank
- **WHEN** analysis completes successfully
- **THEN** candidates SHALL be displayed sorted by score in descending order

#### Scenario: Display all candidate metrics
- **WHEN** viewing candidate list
- **THEN** each candidate SHALL show: rank, type name, frequency, density, consistency percentage, and score

#### Scenario: Sort candidates by column
- **WHEN** user clicks a column header
- **THEN** candidates SHALL be re-sorted by that column

#### Scenario: Select candidate for details
- **WHEN** user clicks a candidate row
- **THEN** system SHALL display detailed properties and promotion option for that candidate

### Requirement: Handle analysis performance
The system SHALL complete analysis within acceptable time limits and provide feedback.

#### Scenario: Analysis completes quickly
- **WHEN** analyzing a dataset with <10,000 discovered entities
- **THEN** analysis SHALL complete within 5 seconds

#### Scenario: Show loading state during analysis
- **WHEN** analysis is in progress
- **THEN** system SHALL display a loading indicator and disable the analyze button

#### Scenario: Analysis timeout handling
- **WHEN** analysis takes longer than 60 seconds
- **THEN** system SHALL cancel the operation and display timeout error

### Requirement: Validate analysis parameters
The system SHALL validate user-provided parameters before executing analysis.

#### Scenario: Reject negative occurrences
- **WHEN** user provides minOccurrences < 1
- **THEN** system SHALL return validation error "Minimum occurrences must be at least 1"

#### Scenario: Reject invalid consistency range
- **WHEN** user provides minConsistency outside range 0.0-1.0
- **THEN** system SHALL return validation error "Consistency must be between 0.0 and 1.0"

#### Scenario: Reject invalid top N
- **WHEN** user provides topN < 1
- **THEN** system SHALL return validation error "Top N must be at least 1"

#### Scenario: Accept valid parameters
- **WHEN** user provides minOccurrences=5, minConsistency=0.4, topN=10
- **THEN** system SHALL execute analysis without validation errors

### Requirement: Reuse analyst package logic
The system SHALL delegate analysis logic to the existing internal/analyst package.

#### Scenario: Call analyst package for analysis
- **WHEN** executing entity analysis
- **THEN** system SHALL call analyst.AnalyzeAndRankCandidates with provided parameters

#### Scenario: Transform analyst results for UI
- **WHEN** analyst package returns results
- **THEN** system SHALL transform TypeCandidate objects to AnalysisResponse format

#### Scenario: Propagate analyst package errors
- **WHEN** analyst package returns an error
- **THEN** system SHALL propagate the error to the frontend with appropriate context
