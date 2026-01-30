/**
 * EntityAnalysis Component Tests
 * Test Suite for Entity Analysis and Promotion Features
 * 
 * Requirements tested (tasks 5.1-5.11):
 * - 5.1: Component renders with configuration controls
 * - 5.2: Configuration values can be updated
 * - 5.3: Validation of configuration parameters
 * - 5.4: Analyze button triggers API call with correct parameters
 * - 5.5: Loading state displays while analyzing
 * - 5.6: Results display in table with all columns
 * - 5.7: Table can be sorted by different columns
 * - 5.8: Clicking a row selects and displays details
 * - 5.9: Promote button calls onPromote callback
 * - 5.10: Empty state when no candidates found
 * - 5.11: Error handling and display
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import EntityAnalysis from './EntityAnalysis';
import { main } from '../wailsjs/go/models';

// Mock the wails API service
vi.mock('../services/wails', () => ({
    wailsAPI: {
        analyzeEntities: vi.fn(),
    },
}));

// Mock LoadingSkeleton component
vi.mock('./LoadingSkeleton', () => ({
    default: () => <div data-testid="loading-skeleton">Loading...</div>,
}));

import { wailsAPI } from '../services/wails';

describe('EntityAnalysis Component', () => {
    const mockOnPromote = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    describe('5.1: Component Rendering', () => {
        it('renders the entity analysis container', () => {
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const heading = screen.getByRole('heading', { name: /entity analysis/i });
            expect(heading).toBeInTheDocument();
        });

        it('renders configuration panel with all controls', () => {
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            expect(screen.getByLabelText(/min occurrences/i)).toBeInTheDocument();
            expect(screen.getByLabelText(/min consistency/i)).toBeInTheDocument();
            expect(screen.getByLabelText(/top n/i)).toBeInTheDocument();
            expect(screen.getByRole('button', { name: /analyze/i })).toBeInTheDocument();
        });

        it('renders with default configuration values', () => {
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const minOccurrences = screen.getByLabelText(/min occurrences/i);
            const minConsistency = screen.getByLabelText(/min consistency/i);
            const topN = screen.getByLabelText(/top n/i);

            expect(minOccurrences).toHaveValue(5);
            expect(minConsistency).toHaveValue(0.4);
            expect(topN).toHaveValue(10);
        });
    });

    describe('5.2: Configuration Updates', () => {
        it('updates minOccurrences when input changes', async () => {
            const user = userEvent.setup();
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const input = screen.getByLabelText(/min occurrences/i);
            await user.tripleClick(input);
            await user.keyboard('10');

            expect(input).toHaveValue(10);
        });

        it('updates minConsistency when input changes', async () => {
            const user = userEvent.setup();
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const input = screen.getByLabelText(/min consistency/i);
            await user.clear(input);
            await user.type(input, '0.7');

            expect(input).toHaveValue(0.7);
        });

        it('updates topN when input changes', async () => {
            const user = userEvent.setup();
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const input = screen.getByLabelText(/top n/i);
            await user.tripleClick(input);
            await user.keyboard('20');

            expect(input).toHaveValue(20);
        });
    });

    describe('5.3: Parameter Validation', () => {
        it('prevents minOccurrences from being set below 1', async () => {
            const user = userEvent.setup();
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const input = screen.getByLabelText(/min occurrences/i);

            await user.tripleClick(input);
            await user.keyboard('0');

            // Component prevents invalid value, defaults to 1
            expect(input).toHaveValue(1);
        });

        it('prevents minConsistency from being set below 0', async () => {
            const user = userEvent.setup();
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const input = screen.getByLabelText(/min consistency/i) as HTMLInputElement;

            await user.tripleClick(input);
            await user.keyboard('-0.1');

            // Component prevents invalid value - negative sign ignored due to min="0"
            // Result is 0.1, not 0 or -0.1
            expect(input.valueAsNumber).toBeGreaterThanOrEqual(0);
        });

        it('allows valid minConsistency values within range', async () => {
            const user = userEvent.setup();
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const input = screen.getByLabelText(/min consistency/i);

            await user.clear(input);
            await user.type(input, '0.75');

            expect(input).toHaveValue(0.75);
        });

        it('prevents topN from being set below 1', async () => {
            const user = userEvent.setup();
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const input = screen.getByLabelText(/top n/i);

            await user.tripleClick(input);
            await user.keyboard('0');

            // Component prevents invalid value, defaults to 1
            expect(input).toHaveValue(1);
        });
    });

    describe('5.4: API Integration - Analysis Request', () => {
        it('calls analyzeEntities API with correct parameters', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [],
                totalTypes: 0,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(wailsAPI.analyzeEntities).toHaveBeenCalledTimes(1);
            });

            const callArg = vi.mocked(wailsAPI.analyzeEntities).mock.calls[0][0];
            expect(callArg.minOccurrences).toBe(5);
            expect(callArg.minConsistency).toBe(0.4);
            expect(callArg.topN).toBe(10);
        });

        it('calls API with updated configuration values', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [],
                totalTypes: 0,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const minOccurrences = screen.getByLabelText(/min occurrences/i);
            const minConsistency = screen.getByLabelText(/min consistency/i);
            const topN = screen.getByLabelText(/top n/i);
            const analyzeButton = screen.getByRole('button', { name: /analyze/i });

            await user.tripleClick(minOccurrences);
            await user.keyboard('15');
            await user.tripleClick(minConsistency);
            await user.keyboard('0.8');
            await user.tripleClick(topN);
            await user.keyboard('25');
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(wailsAPI.analyzeEntities).toHaveBeenCalledTimes(1);
            });

            const callArg = vi.mocked(wailsAPI.analyzeEntities).mock.calls[0][0];
            expect(callArg.minOccurrences).toBe(15);
            expect(callArg.minConsistency).toBe(0.8);
            expect(callArg.topN).toBe(25);
        });
    });

    describe('5.5: Loading State', () => {
        it('shows loading state while analyzing', async () => {
            const user = userEvent.setup();
            let resolvePromise: (value: main.AnalysisResponse) => void;
            const promise = new Promise<main.AnalysisResponse>((resolve) => {
                resolvePromise = resolve;
            });

            vi.mocked(wailsAPI.analyzeEntities).mockReturnValue(promise);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByTestId('loading-skeleton')).toBeInTheDocument();
            });

            resolvePromise!(new main.AnalysisResponse({ candidates: [], totalTypes: 0 }));

            await waitFor(() => {
                expect(screen.queryByTestId('loading-skeleton')).not.toBeInTheDocument();
            });
        });

        it('disables analyze button while loading', async () => {
            const user = userEvent.setup();
            let resolvePromise: (value: main.AnalysisResponse) => void;
            const promise = new Promise<main.AnalysisResponse>((resolve) => {
                resolvePromise = resolve;
            });

            vi.mocked(wailsAPI.analyzeEntities).mockReturnValue(promise);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /analyzing/i })).toBeDisabled();
            });

            resolvePromise!(new main.AnalysisResponse({ candidates: [], totalTypes: 0 }));

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /analyze/i })).not.toBeDisabled();
            });
        });

        it('disables configuration inputs while loading', async () => {
            const user = userEvent.setup();
            let resolvePromise: (value: main.AnalysisResponse) => void;
            const promise = new Promise<main.AnalysisResponse>((resolve) => {
                resolvePromise = resolve;
            });

            vi.mocked(wailsAPI.analyzeEntities).mockReturnValue(promise);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByLabelText(/min occurrences/i)).toBeDisabled();
                expect(screen.getByLabelText(/min consistency/i)).toBeDisabled();
                expect(screen.getByLabelText(/top n/i)).toBeDisabled();
            });

            resolvePromise!(new main.AnalysisResponse({ candidates: [], totalTypes: 0 }));
        });
    });

    describe('5.6: Results Display', () => {
        it('displays results table with candidates', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                    new main.TypeCandidate({
                        rank: 2,
                        typeName: 'Organization',
                        frequency: 75,
                        density: 0.78,
                        consistency: 0.88,
                        score: 0.81,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Person')).toBeInTheDocument();
                expect(screen.getByText('Organization')).toBeInTheDocument();
            });
        });

        it('displays all required columns', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByRole('columnheader', { name: /rank/i })).toBeInTheDocument();
                expect(screen.getByRole('columnheader', { name: /type name/i })).toBeInTheDocument();
                expect(screen.getByRole('columnheader', { name: /frequency/i })).toBeInTheDocument();
                expect(screen.getByRole('columnheader', { name: /density/i })).toBeInTheDocument();
                expect(screen.getByRole('columnheader', { name: /consistency/i })).toBeInTheDocument();
                expect(screen.getByRole('columnheader', { name: /score/i })).toBeInTheDocument();
                expect(screen.getByRole('columnheader', { name: /actions/i })).toBeInTheDocument();
            });
        });

        it('displays results summary', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText(/found 1 candidates out of 50 total types/i)).toBeInTheDocument();
            });
        });

        it('formats decimal values correctly', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.856,
                        consistency: 0.923,
                        score: 0.894,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('0.86')).toBeInTheDocument();
                expect(screen.getByText('0.92')).toBeInTheDocument();
                expect(screen.getByText('0.89')).toBeInTheDocument();
            });
        });
    });

    describe('5.7: Table Sorting', () => {
        const createMockCandidates = () => [
            new main.TypeCandidate({
                rank: 1,
                typeName: 'Alpha',
                frequency: 50,
                density: 0.5,
                consistency: 0.8,
                score: 0.7,
            }),
            new main.TypeCandidate({
                rank: 2,
                typeName: 'Beta',
                frequency: 100,
                density: 0.9,
                consistency: 0.6,
                score: 0.9,
            }),
            new main.TypeCandidate({
                rank: 3,
                typeName: 'Gamma',
                frequency: 75,
                density: 0.7,
                consistency: 0.9,
                score: 0.8,
            }),
        ];

        it('sorts by rank in ascending order by default', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: createMockCandidates(),
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                const rows = screen.getAllByRole('row');
                expect(rows[1]).toHaveTextContent('Alpha');
                expect(rows[2]).toHaveTextContent('Beta');
                expect(rows[3]).toHaveTextContent('Gamma');
            });
        });

        it('toggles sort direction when clicking same column header', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: createMockCandidates(),
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Alpha')).toBeInTheDocument();
            });

            const rankHeader = screen.getByText(/rank/i, { selector: 'th' });
            await user.click(rankHeader);

            await waitFor(() => {
                const rows = screen.getAllByRole('row');
                expect(rows[1]).toHaveTextContent('Gamma');
                expect(rows[2]).toHaveTextContent('Beta');
                expect(rows[3]).toHaveTextContent('Alpha');
            });
        });

        it('sorts by type name alphabetically', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: createMockCandidates(),
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Alpha')).toBeInTheDocument();
            });

            const typeNameHeader = screen.getByText(/type name/i, { selector: 'th' });
            await user.click(typeNameHeader);

            await waitFor(() => {
                const rows = screen.getAllByRole('row');
                expect(rows[1]).toHaveTextContent('Alpha');
                expect(rows[2]).toHaveTextContent('Beta');
                expect(rows[3]).toHaveTextContent('Gamma');
            });
        });

        it('sorts by frequency numerically', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: createMockCandidates(),
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Alpha')).toBeInTheDocument();
            });

            const frequencyHeader = screen.getByText(/frequency/i, { selector: 'th' });
            await user.click(frequencyHeader);

            await waitFor(() => {
                const rows = screen.getAllByRole('row');
                expect(rows[1]).toHaveTextContent('Alpha');
                expect(rows[2]).toHaveTextContent('Gamma');
                expect(rows[3]).toHaveTextContent('Beta');
            });
        });

        it('sorts by score numerically', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: createMockCandidates(),
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Alpha')).toBeInTheDocument();
            });

            const scoreHeader = screen.getByText(/score/i, { selector: 'th' });
            await user.click(scoreHeader);

            await waitFor(() => {
                const rows = screen.getAllByRole('row');
                expect(rows[1]).toHaveTextContent('Alpha');
                expect(rows[2]).toHaveTextContent('Gamma');
                expect(rows[3]).toHaveTextContent('Beta');
            });
        });

        it('displays sort direction indicator', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: createMockCandidates(),
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                const rankHeader = screen.getByText(/rank/i, { selector: 'th' });
                expect(rankHeader).toHaveTextContent('↑');
            });

            const rankHeader = screen.getByText(/rank/i, { selector: 'th' });
            await user.click(rankHeader);

            await waitFor(() => {
                expect(rankHeader).toHaveTextContent('↓');
            });
        });
    });

    describe('5.8: Row Selection and Details', () => {
        it('selects row when clicked', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Person')).toBeInTheDocument();
            });

            const row = screen.getByText('Person').closest('tr');
            await user.click(row!);

            await waitFor(() => {
                expect(row).toHaveClass('selected');
            });
        });

        it('displays candidate details when row is selected', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.856,
                        consistency: 0.923,
                        score: 0.894,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Person')).toBeInTheDocument();
            });

            const row = screen.getByText('Person').closest('tr');
            await user.click(row!);

            await waitFor(() => {
                expect(screen.getByText(/selected: person/i)).toBeInTheDocument();
                expect(screen.getByText('0.856')).toBeInTheDocument();
                expect(screen.getByText('0.923')).toBeInTheDocument();
                expect(screen.getByText('0.894')).toBeInTheDocument();
            });
        });

        it('changes selection when different row is clicked', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                    new main.TypeCandidate({
                        rank: 2,
                        typeName: 'Organization',
                        frequency: 75,
                        density: 0.78,
                        consistency: 0.88,
                        score: 0.81,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText('Person')).toBeInTheDocument();
            });

            const firstRow = screen.getByText('Person').closest('tr');
            await user.click(firstRow!);

            await waitFor(() => {
                expect(screen.getByText(/selected: person/i)).toBeInTheDocument();
            });

            const secondRow = screen.getByText('Organization').closest('tr');
            await user.click(secondRow!);

            await waitFor(() => {
                expect(screen.getByText(/selected: organization/i)).toBeInTheDocument();
                expect(screen.queryByText(/selected: person/i)).not.toBeInTheDocument();
            });
        });
    });

    describe('5.9: Promote Button', () => {
        it('displays promote button for each candidate', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /promote/i })).toBeInTheDocument();
            });
        });

        it('calls onPromote with type name when promote button is clicked', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /promote/i })).toBeInTheDocument();
            });

            const promoteButton = screen.getByRole('button', { name: /promote/i });
            await user.click(promoteButton);

            expect(mockOnPromote).toHaveBeenCalledTimes(1);
            expect(mockOnPromote).toHaveBeenCalledWith('Person');
        });

        it('does not select row when promote button is clicked', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [
                    new main.TypeCandidate({
                        rank: 1,
                        typeName: 'Person',
                        frequency: 100,
                        density: 0.85,
                        consistency: 0.92,
                        score: 0.89,
                    }),
                ],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /promote/i })).toBeInTheDocument();
            });

            const promoteButton = screen.getByRole('button', { name: /promote/i });
            await user.click(promoteButton);

            const row = screen.getByText('Person').closest('tr');
            expect(row).not.toHaveClass('selected');
        });
    });

    describe('5.10: Empty State', () => {
        it('displays empty state when no candidates are found', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText(/no candidates found matching the specified criteria/i)).toBeInTheDocument();
            });
        });

        it('displays help text in empty state', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText(/try adjusting the configuration parameters/i)).toBeInTheDocument();
            });
        });

        it('does not show empty state before analysis', () => {
            render(<EntityAnalysis onPromote={mockOnPromote} />);

            expect(screen.queryByText(/no candidates found/i)).not.toBeInTheDocument();
        });

        it('does not show results table when empty', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.AnalysisResponse({
                candidates: [],
                totalTypes: 50,
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValue(mockResponse);

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText(/no candidates found/i)).toBeInTheDocument();
            });

            expect(screen.queryByRole('table')).not.toBeInTheDocument();
        });
    });

    describe('5.11: Error Handling', () => {
        it('displays error message when API call fails', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.analyzeEntities).mockRejectedValue(new Error('Database connection failed'));

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText(/database connection failed/i)).toBeInTheDocument();
            });
        });

        it('displays retry button when error occurs', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.analyzeEntities).mockRejectedValue(new Error('Network error'));

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
            });
        });

        it('retries analysis when retry button is clicked', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.analyzeEntities).mockRejectedValueOnce(new Error('Network error'));
            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValueOnce(
                new main.AnalysisResponse({ candidates: [], totalTypes: 0 })
            );

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
            });

            const retryButton = screen.getByRole('button', { name: /retry/i });
            await user.click(retryButton);

            await waitFor(() => {
                expect(wailsAPI.analyzeEntities).toHaveBeenCalledTimes(2);
            });
        });

        it('clears previous error on new analysis', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.analyzeEntities).mockRejectedValueOnce(new Error('Network error'));

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText(/network error/i)).toBeInTheDocument();
            });

            vi.mocked(wailsAPI.analyzeEntities).mockResolvedValueOnce(
                new main.AnalysisResponse({ candidates: [], totalTypes: 0 })
            );

            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.queryByText(/network error/i)).not.toBeInTheDocument();
            });
        });

        it('handles non-Error exceptions gracefully', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.analyzeEntities).mockRejectedValue('String error');

            render(<EntityAnalysis onPromote={mockOnPromote} />);

            const analyzeButton = screen.getByRole('button', { name: /analyze/i });
            await user.click(analyzeButton);

            await waitFor(() => {
                expect(screen.getByText(/failed to analyze entities/i)).toBeInTheDocument();
            });
        });
    });
});
