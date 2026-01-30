/**
 * EntityPromotion Component Tests
 * Test Suite for Entity Promotion Feature
 * 
 * Requirements tested (6.1-6.10):
 * - Preview rendering with type name
 * - Property list display in success state
 * - Entity count display
 * - Cancel button functionality
 * - Confirm button triggers API
 * - Loading state during promotion
 * - UI disabled during promotion
 * - Success display with results
 * - Failure display with error
 * - Validation errors display
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import EntityPromotion from './EntityPromotion';
import { main } from '../wailsjs/go/models';

// Mock the wailsAPI module
vi.mock('../services/wails', () => ({
    wailsAPI: {
        promoteEntity: vi.fn(),
    },
}));

import { wailsAPI } from '../services/wails';

describe('EntityPromotion Component', () => {
    const mockOnCancel = vi.fn();
    const mockOnSuccess = vi.fn();
    const mockOnViewInGraph = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    describe('6.1: Preview Rendering', () => {
        it('renders nothing when typeName is null', () => {
            const { container } = render(
                <EntityPromotion
                    typeName={null}
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            expect(container.firstChild).toBeNull();
        });

        it('renders promotion preview when typeName is provided', () => {
            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            expect(screen.getByText('Promote Entity Type')).toBeInTheDocument();
            expect(screen.getByText('Type to Promote')).toBeInTheDocument();
            expect(screen.getByText('Person')).toBeInTheDocument();
        });

        it('displays preview description with promotion steps', () => {
            render(
                <EntityPromotion
                    typeName="Organization"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            expect(screen.getByText(/This will:/i)).toBeInTheDocument();
            expect(screen.getByText(/Analyze entity properties and generate schema/i)).toBeInTheDocument();
            expect(screen.getByText(/Create an Ent schema file/i)).toBeInTheDocument();
            expect(screen.getByText(/Run database migration/i)).toBeInTheDocument();
            expect(screen.getByText(/Migrate existing discovered entities/i)).toBeInTheDocument();
        });

        it('resets state when typeName changes', () => {
            const { rerender } = render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            // Check initial state
            expect(screen.getByText('Person')).toBeInTheDocument();

            // Change typeName
            rerender(
                <EntityPromotion
                    typeName="Organization"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            // Verify new typeName is displayed
            expect(screen.getByText('Organization')).toBeInTheDocument();
            expect(screen.queryByText('Person')).not.toBeInTheDocument();
        });
    });

    describe('6.4: Cancel Button', () => {
        it('renders cancel button', () => {
            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        });

        it('calls onCancel when cancel button is clicked', async () => {
            const user = userEvent.setup();
            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const cancelButton = screen.getByRole('button', { name: /cancel/i });
            await user.click(cancelButton);

            expect(mockOnCancel).toHaveBeenCalledTimes(1);
        });

        it('calls onCancel when close button is clicked', async () => {
            const user = userEvent.setup();
            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const closeButton = screen.getByRole('button', { name: /close/i });
            await user.click(closeButton);

            expect(mockOnCancel).toHaveBeenCalledTimes(1);
        });

        it('does not call onCancel when disabled during loading', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.promoteEntity).mockImplementation(
                () => new Promise(() => { }) // Never resolves
            );

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            // Start promotion
            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            // Try to click cancel while loading
            const cancelButton = screen.getByRole('button', { name: /cancel/i });
            expect(cancelButton).toBeDisabled();
        });
    });

    describe('6.5: Confirm Button and API Call', () => {
        it('renders confirm button', () => {
            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            expect(screen.getByRole('button', { name: /confirm promote/i })).toBeInTheDocument();
        });

        it('calls promoteEntity API when confirm button is clicked', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(wailsAPI.promoteEntity).toHaveBeenCalledTimes(1);
            });

            const callArg = vi.mocked(wailsAPI.promoteEntity).mock.calls[0][0];
            expect(callArg).toBeInstanceOf(main.PromotionRequest);
            expect(callArg.typeName).toBe('Person');
        });
    });

    describe('6.6 & 6.7: Loading State and UI Disabled', () => {
        it('shows loading overlay during promotion', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.promoteEntity).mockImplementation(
                () => new Promise(() => { }) // Never resolves
            );

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promoting entity type/i)).toBeInTheDocument();
            });

            expect(screen.getByText(/this may take a few moments/i)).toBeInTheDocument();
        });

        it('disables buttons during loading', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.promoteEntity).mockImplementation(
                () => new Promise(() => { }) // Never resolves
            );

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled();
            });

            expect(screen.getByRole('button', { name: /close/i })).toBeDisabled();
            expect(screen.getByRole('button', { name: /promoting/i })).toBeDisabled();
        });

        it('changes button text during loading', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.promoteEntity).mockImplementation(
                () => new Promise(() => { }) // Never resolves
            );

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            expect(screen.getByRole('button', { name: /confirm promote/i })).toBeInTheDocument();

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /promoting/i })).toBeInTheDocument();
            });

            expect(screen.queryByRole('button', { name: /confirm promote/i })).not.toBeInTheDocument();
        });
    });

    describe('6.8 & 6.2 & 6.3: Success Display with Properties and Entity Count', () => {
        it('displays success results after successful promotion', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 42,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion successful/i)).toBeInTheDocument();
            });

            expect(screen.getByText(/schema file:/i)).toBeInTheDocument();
            expect(screen.getByText('ent/schema/person.go')).toBeInTheDocument();
        });

        it('displays entity count in success results', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 123,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/entities migrated:/i)).toBeInTheDocument();
            });

            expect(screen.getByText('123')).toBeInTheDocument();
        });

        it('displays property list in success results', async () => {
            const user = userEvent.setup();
            const mockProperties = [
                { name: 'firstName', type: 'string', required: true },
                { name: 'lastName', type: 'string', required: true },
                { name: 'email', type: 'string', required: false },
                { name: 'age', type: 'int', required: false },
            ];

            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 42,
                validationErrors: 0,
                properties: mockProperties,
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/generated properties \(4\)/i)).toBeInTheDocument();
            });

            expect(screen.getByText('firstName')).toBeInTheDocument();
            expect(screen.getByText('lastName')).toBeInTheDocument();
            expect(screen.getByText('email')).toBeInTheDocument();
            expect(screen.getByText('age')).toBeInTheDocument();

            // Check types
            const stringTypes = screen.getAllByText('string');
            expect(stringTypes).toHaveLength(3);
            expect(screen.getByText('int')).toBeInTheDocument();

            // Check required labels (should be 2 required properties)
            const requiredLabels = screen.getAllByText('Required');
            expect(requiredLabels).toHaveLength(2);
        });

        it('calls onSuccess after 2 seconds on successful promotion', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            // Wait for API call to complete and success message to appear
            await screen.findByText(/promotion successful/i);

            // onSuccess should NOT be called yet (waiting for 2-second timeout)
            expect(mockOnSuccess).not.toHaveBeenCalled();

            // Wait for the 2-second timeout to complete
            await waitFor(() => {
                expect(mockOnSuccess).toHaveBeenCalledTimes(1);
            }, { timeout: 3000 });
        }, 10000);

        it('displays Done button in success state', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /done/i })).toBeInTheDocument();
            });
        });

        it('calls onCancel when Done button is clicked', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /done/i })).toBeInTheDocument();
            });

            const doneButton = screen.getByRole('button', { name: /done/i });
            await user.click(doneButton);

            expect(mockOnCancel).toHaveBeenCalledTimes(1);
        });
    });

    describe('6.10: Validation Errors Display', () => {
        it('displays validation errors count when present', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 100,
                validationErrors: 15,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/validation errors:/i)).toBeInTheDocument();
            });

            expect(screen.getByText('15')).toBeInTheDocument();
            expect(screen.getByText(/some entities could not be migrated/i)).toBeInTheDocument();
        });

        it('does not display validation errors section when count is zero', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 100,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion successful/i)).toBeInTheDocument();
            });

            expect(screen.queryByText(/validation errors:/i)).not.toBeInTheDocument();
        });
    });

    describe('6.9: Failure Display', () => {
        it('displays failure results when promotion fails', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: false,
                typeName: 'Person',
                error: 'Database connection failed',
                schemaFilePath: '',
                entitiesMigrated: 0,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion failed/i)).toBeInTheDocument();
            });

            expect(screen.getByText('Database connection failed')).toBeInTheDocument();
        });

        it('displays error message when API call throws', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.promoteEntity).mockRejectedValue(
                new Error('Network error occurred')
            );

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/Network error occurred/i)).toBeInTheDocument();
            });
        });

        it('displays generic error message for non-Error exceptions', async () => {
            const user = userEvent.setup();
            vi.mocked(wailsAPI.promoteEntity).mockRejectedValue('Unknown error');

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/Failed to promote entity/i)).toBeInTheDocument();
            });
        });

        it('displays retry and cancel buttons in failure state', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: false,
                typeName: 'Person',
                error: 'Database error',
                schemaFilePath: '',
                entitiesMigrated: 0,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion failed/i)).toBeInTheDocument();
            });

            expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
            expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        });

        it('retries promotion when retry button is clicked', async () => {
            const user = userEvent.setup();
            const mockFailureResponse = new main.PromotionResponse({
                success: false,
                typeName: 'Person',
                error: 'Temporary error',
                schemaFilePath: '',
                entitiesMigrated: 0,
                validationErrors: 0,
                properties: [],
            });

            const mockSuccessResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity)
                .mockResolvedValueOnce(mockFailureResponse)
                .mockResolvedValueOnce(mockSuccessResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion failed/i)).toBeInTheDocument();
            });

            const retryButton = await screen.findByRole('button', { name: /retry/i });
            await user.click(retryButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion successful/i)).toBeInTheDocument();
            });

            expect(wailsAPI.promoteEntity).toHaveBeenCalledTimes(2);
        });

        it('does not call onSuccess when promotion fails', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: false,
                typeName: 'Person',
                error: 'Database error',
                schemaFilePath: '',
                entitiesMigrated: 0,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion failed/i)).toBeInTheDocument();
            });

            // onSuccess should not be called
            expect(mockOnSuccess).not.toHaveBeenCalled();
        });
    });

    describe('Optional onViewInGraph Callback', () => {
        it('displays View in Graph button when onViewInGraph is provided', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                    onViewInGraph={mockOnViewInGraph}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /view in graph/i })).toBeInTheDocument();
            });
        });

        it('does not display View in Graph button when onViewInGraph is not provided', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByText(/promotion successful/i)).toBeInTheDocument();
            });

            expect(screen.queryByRole('button', { name: /view in graph/i })).not.toBeInTheDocument();
        });

        it('calls onViewInGraph with typeName when button is clicked', async () => {
            const user = userEvent.setup();
            const mockResponse = new main.PromotionResponse({
                success: true,
                typeName: 'Person',
                schemaFilePath: 'ent/schema/person.go',
                entitiesMigrated: 5,
                validationErrors: 0,
                properties: [],
            });

            vi.mocked(wailsAPI.promoteEntity).mockResolvedValue(mockResponse);

            render(
                <EntityPromotion
                    typeName="Person"
                    onCancel={mockOnCancel}
                    onSuccess={mockOnSuccess}
                    onViewInGraph={mockOnViewInGraph}
                />
            );

            const confirmButton = screen.getByRole('button', { name: /confirm promote/i });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(screen.getByRole('button', { name: /view in graph/i })).toBeInTheDocument();
            });

            const viewGraphButton = screen.getByRole('button', { name: /view in graph/i });
            await user.click(viewGraphButton);

            expect(mockOnViewInGraph).toHaveBeenCalledWith('Person');
            expect(mockOnViewInGraph).toHaveBeenCalledTimes(1);
        });
    });
});
