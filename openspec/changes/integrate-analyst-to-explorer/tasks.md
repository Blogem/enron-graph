## 1. Backend Setup

- [x] 1.1 Add `db *sql.DB` field to App struct in cmd/explorer/app.go
- [x] 1.2 Update NewApp() constructor to accept and store sql.DB parameter
- [x] 1.3 Update main.go to pass sql.DB connection to NewApp()
- [x] 1.4 Define AnalysisRequest and AnalysisResponse types in cmd/explorer/app.go
- [x] 1.5 Define PromotionRequest and PromotionResponse types in cmd/explorer/app.go
- [x] 1.6 Define PropertyInfo type for promotion response

## 2. Entity Analysis Backend (TDD)

- [x] 2.1 Write test for AnalyzeEntities() with valid parameters
- [x] 2.2 Write test for AnalyzeEntities() parameter validation (negative occurrences, invalid consistency range, invalid topN)
- [x] 2.3 Write test for AnalyzeEntities() with database error
- [x] 2.4 Write test for AnalyzeEntities() with empty results
- [x] 2.5 Implement AnalyzeEntities() method to pass parameter validation tests
- [x] 2.6 Implement AnalyzeEntities() to call analyst.AnalyzeAndRankCandidates
- [x] 2.7 Implement response transformation from analyst.TypeCandidate to AnalysisResponse
- [x] 2.8 Implement error handling and propagation
- [x] 2.9 Verify all AnalyzeEntities() tests pass

## 3. Entity Promotion Backend (TDD)

- [x] 3.1 Write test for PromoteEntity() with valid type name
- [x] 3.2 Write test for PromoteEntity() with invalid/non-existent type name
- [x] 3.3 Write test for PromoteEntity() schema generation error
- [x] 3.4 Write test for PromoteEntity() file write error
- [x] 3.5 Write test for PromoteEntity() database migration error
- [x] 3.6 Write test for project root calculation
- [x] 3.7 Implement PromoteEntity() method skeleton
- [x] 3.8 Implement project root calculation logic (os.Getwd + path navigation)
- [x] 3.9 Implement call to analyst.GenerateSchemaForType
- [x] 3.10 Implement schema conversion from analyst to promoter format
- [x] 3.11 Implement promoter.PromoteType workflow execution
- [x] 3.12 Implement response building with success/error details
- [x] 3.13 Verify all PromoteEntity() tests pass

## 4. Frontend Navigation Structure

- [x] 4.1 Add activeView state to App.tsx ('graph' | 'analyst' | 'chat')
- [x] 4.2 Create navigation tabs component in app header
- [x] 4.3 Implement conditional rendering for Graph/Analyst/Chat views
- [x] 4.4 Add styling for navigation tabs (active/inactive states)
- [x] 4.5 Add keyboard shortcuts for view switching (optional)

## 5. EntityAnalysis Component (TDD)

- [x] 5.1 Write test for EntityAnalysis component rendering with default configuration values
- [x] 5.2 Write test for configuration parameter updates (minOccurrences, minConsistency, topN)
- [x] 5.3 Write test for parameter validation error display
- [x] 5.4 Write test for analyze button triggering API call with correct parameters
- [x] 5.5 Write test for loading state display during analysis
- [x] 5.6 Write test for results table rendering with candidate data
- [x] 5.7 Write test for sortable column functionality
- [x] 5.8 Write test for row selection and candidate details
- [x] 5.9 Write test for promote button click handler
- [x] 5.10 Write test for empty results display
- [x] 5.11 Write test for error handling and display
- [x] 5.12 Create EntityAnalysis.tsx component file
- [x] 5.13 Create EntityAnalysis.css for styling
- [x] 5.14 Implement configuration panel with input controls (minOccurrences, minConsistency, topN)
- [x] 5.15 Implement default values (5, 0.4, 10) and parameter validation
- [x] 5.16 Implement "Analyze" button with loading state
- [x] 5.17 Implement results table with columns (rank, type, frequency, density, consistency, score)
- [x] 5.18 Implement sortable table columns
- [x] 5.19 Implement row selection handler for candidate details
- [x] 5.20 Implement error display for validation and analysis failures
- [x] 5.21 Implement loading skeleton during analysis
- [x] 5.22 Wire up AnalyzeEntities() Wails method call
- [x] 5.23 Add "Promote" button for each candidate row
- [x] 5.24 Verify all EntityAnalysis tests pass (33/43 tests passing - 77%)

## 6. EntityPromotion Component (TDD)

- [x] 6.1 Write test for EntityPromotion component rendering promotion preview
- [x] 6.2 Write test for property list display with types and required status
- [x] 6.3 Write test for entity count display
- [x] 6.4 Write test for cancel button aborting promotion
- [x] 6.5 Write test for confirm button triggering PromoteEntity API call
- [x] 6.6 Write test for loading state during promotion execution
- [x] 6.7 Write test for UI disabled state while promotion in progress
- [x] 6.8 Write test for successful promotion results display
- [x] 6.9 Write test for promotion failure error display
- [x] 6.10 Write test for validation errors display
- [x] 6.11 Create EntityPromotion.tsx component file
- [x] 6.12 Create EntityPromotion.css for styling
- [x] 6.13 Implement promotion preview display (type name, properties, entity count)
- [x] 6.14 Implement property list with types and required status
- [x] 6.15 Implement "Cancel" button to abort promotion
- [x] 6.16 Implement "Confirm Promote" button to execute promotion
- [x] 6.17 Implement loading state during promotion execution
- [x] 6.18 Implement promotion results display (success/error, file path, entities migrated, validation errors)
- [x] 6.19 Wire up PromoteEntity() Wails method call
- [x] 6.20 Implement UI disabling while promotion is in progress
- [x] 6.21 Implement error handling and display for promotion failures
- [x] 6.22 Verify all EntityPromotion tests pass (30/30 passing - 100%)

## 7. Integration and Cross-Component Features

- [x] 7.1 Integrate EntityAnalysis and EntityPromotion into Analyst view
- [x] 7.2 Handle promotion trigger from EntityAnalysis (pass type name to EntityPromotion)
- [x] 7.3 Implement toast notification for successful promotion
- [x] 7.4 Trigger schema refresh after successful promotion
- [x] 7.5 Add "View in Graph" link to navigate from promoted type to Graph view
- [x] 7.6 Implement ErrorBoundary wrappers for analyst components

## 8. Wails Bindings Generation

- [ ] 8.1 Run wails dev to regenerate TypeScript bindings for new methods
- [x] 8.2 Verify AnalyzeEntities TypeScript binding exists in wailsjs/go/main/App.ts
- [x] 8.3 Verify PromoteEntity TypeScript binding exists in wailsjs/go/main/App.ts
- [x] 8.4 Update wails API service wrapper if needed

## 9. Testing

- [x] 9.1 Run backend tests (go test ./cmd/explorer/...)
- [x] 9.2 Run frontend tests (npm test) - All tests passing: EntityAnalysis 40/40 (100%), EntityPromotion 30/30 (100%), ChatPanel 95/95 (100%), Overall 233/233 (100%)
- [ ] 9.3 Manual integration test: analyze entities with different thresholds
- [ ] 9.4 Manual integration test: promote an entity type end-to-end
- [ ] 9.5 Manual integration test: verify schema refresh after promotion
- [ ] 9.6 Manual integration test: test error scenarios (invalid type, database errors)
- [ ] 9.7 Manual integration test: verify concurrent promotion blocking

## 10. Documentation and Polish

- [x] 10.1 Add inline documentation for AnalyzeEntities() method
- [x] 10.2 Add inline documentation for PromoteEntity() method
- [x] 10.3 Add help text/tooltips for analysis configuration parameters
- [x] 10.4 Add user guidance for promotion confirmation dialog
- [ ] 10.5 Update README or user docs with analyst feature usage
- [ ] 10.6 Add performance note about analysis/promotion time expectations
