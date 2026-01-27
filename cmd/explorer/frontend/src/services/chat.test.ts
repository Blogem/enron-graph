/**
 * Chat Service Tests (T026)
 * Test Suite for User Story 2 - Send Queries and Display Responses
 * 
 * Requirements tested:
 * - Chat service wraps Wails API calls (ProcessChatQuery, ClearChatContext)
 * - FR-024: Query processing timeout of 60 seconds
 * - FR-022: Retry functionality for failed queries
 * - FR-011: User-friendly error messages
 * - Error handling and error type classification (can retry vs cannot retry)
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { chatService, ChatServiceError } from './chat';

// Mock the Wails runtime bindings
vi.mock('../wailsjs/go/main/App', () => ({
    ProcessChatQuery: vi.fn(),
    ClearChatContext: vi.fn(),
}));

import { ProcessChatQuery, ClearChatContext } from '../wailsjs/go/main/App';

describe('Chat Service', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.clearAllTimers();
    });

    describe('processQuery', () => {
        it('calls ProcessChatQuery with the provided query', async () => {
            const mockResponse = 'This is a test response';
            vi.mocked(ProcessChatQuery).mockResolvedValue(mockResponse);

            const result = await chatService.processQuery('test query');

            expect(ProcessChatQuery).toHaveBeenCalledWith('test query');
            expect(result).toBe(mockResponse);
        });

        it('returns the response text from ProcessChatQuery', async () => {
            const mockResponse = 'Graph has 42 nodes of type Person';
            vi.mocked(ProcessChatQuery).mockResolvedValue(mockResponse);

            const result = await chatService.processQuery('How many Person nodes?');

            expect(result).toBe(mockResponse);
        });

        it('handles empty query string', async () => {
            const mockResponse = 'Please provide a query';
            vi.mocked(ProcessChatQuery).mockResolvedValue(mockResponse);

            const result = await chatService.processQuery('');

            expect(ProcessChatQuery).toHaveBeenCalledWith('');
            expect(result).toBe(mockResponse);
        });

        it('handles query with special characters', async () => {
            const mockResponse = 'Found results';
            vi.mocked(ProcessChatQuery).mockResolvedValue(mockResponse);

            const query = 'Find nodes where name contains "O\'Brien" & category = "Test"';
            const result = await chatService.processQuery(query);

            expect(ProcessChatQuery).toHaveBeenCalledWith(query);
            expect(result).toBe(mockResponse);
        });

        describe('Error Handling', () => {
            it('throws ChatServiceError when ProcessChatQuery fails', async () => {
                const errorMessage = 'Database connection failed';
                vi.mocked(ProcessChatQuery).mockRejectedValue(new Error(errorMessage));

                await expect(chatService.processQuery('test')).rejects.toThrow(ChatServiceError);
            });

            it('wraps error with user-friendly message', async () => {
                vi.mocked(ProcessChatQuery).mockRejectedValue(new Error('Internal error'));

                try {
                    await chatService.processQuery('test');
                    expect.fail('Should have thrown error');
                } catch (error) {
                    expect(error).toBeInstanceOf(ChatServiceError);
                    expect((error as ChatServiceError).message).toContain('Failed to process query');
                }
            });

            it('preserves original error message in ChatServiceError', async () => {
                const originalError = 'LLM service unavailable';
                vi.mocked(ProcessChatQuery).mockRejectedValue(new Error(originalError));

                try {
                    await chatService.processQuery('test');
                    expect.fail('Should have thrown error');
                } catch (error) {
                    expect(error).toBeInstanceOf(ChatServiceError);
                    expect((error as ChatServiceError).originalError).toBe(originalError);
                }
            });

            it('marks network errors as retryable', async () => {
                const networkError = new Error('Network request failed');
                vi.mocked(ProcessChatQuery).mockRejectedValue(networkError);

                try {
                    await chatService.processQuery('test');
                    expect.fail('Should have thrown error');
                } catch (error) {
                    expect(error).toBeInstanceOf(ChatServiceError);
                    expect((error as ChatServiceError).canRetry).toBe(true);
                }
            });

            it('marks validation errors as non-retryable', async () => {
                const validationError = new Error('Query exceeds maximum length');
                vi.mocked(ProcessChatQuery).mockRejectedValue(validationError);

                try {
                    await chatService.processQuery('test');
                    expect.fail('Should have thrown error');
                } catch (error) {
                    expect(error).toBeInstanceOf(ChatServiceError);
                    expect((error as ChatServiceError).canRetry).toBe(false);
                }
            });

            it('marks timeout errors as retryable', async () => {
                const timeoutError = new Error('Query timeout');
                vi.mocked(ProcessChatQuery).mockRejectedValue(timeoutError);

                try {
                    await chatService.processQuery('test');
                    expect.fail('Should have thrown error');
                } catch (error) {
                    expect(error).toBeInstanceOf(ChatServiceError);
                    expect((error as ChatServiceError).canRetry).toBe(true);
                }
            });
        });

        describe('FR-024: Timeout Handling', () => {
            beforeEach(() => {
                vi.useFakeTimers();
            });

            afterEach(() => {
                vi.useRealTimers();
            });

            it('applies 60 second timeout to queries', async () => {
                // Mock a query that never resolves
                vi.mocked(ProcessChatQuery).mockImplementation(
                    () => new Promise(() => { }) // Never resolves
                );

                const queryPromise = chatService.processQuery('slow query');

                // Advance time by 60 seconds
                vi.advanceTimersByTime(60000);

                await expect(queryPromise).rejects.toThrow(ChatServiceError);
            });

            it('timeout error is marked as retryable', async () => {
                vi.mocked(ProcessChatQuery).mockImplementation(
                    () => new Promise(() => { }) // Never resolves
                );

                const queryPromise = chatService.processQuery('slow query');

                // Advance time past timeout
                vi.advanceTimersByTime(60000);

                try {
                    await queryPromise;
                    expect.fail('Should have thrown timeout error');
                } catch (error) {
                    expect(error).toBeInstanceOf(ChatServiceError);
                    expect((error as ChatServiceError).message).toContain('timeout');
                    expect((error as ChatServiceError).canRetry).toBe(true);
                }
            });

            it('does not timeout if query completes before 60 seconds', async () => {
                const mockResponse = 'Quick response';
                vi.mocked(ProcessChatQuery).mockImplementation(
                    () => new Promise((resolve) => {
                        setTimeout(() => resolve(mockResponse), 30000);
                    })
                );

                const queryPromise = chatService.processQuery('quick query');

                // Advance time by 30 seconds (less than timeout)
                vi.advanceTimersByTime(30000);

                const result = await queryPromise;
                expect(result).toBe(mockResponse);
            });

            it('cleans up timeout timer when query completes', async () => {
                const mockResponse = 'Response';
                vi.mocked(ProcessChatQuery).mockResolvedValue(mockResponse);

                await chatService.processQuery('test');

                // Verify no timers are pending
                expect(vi.getTimerCount()).toBe(0);
            });
        });

        describe('Concurrent Queries', () => {
            it('handles multiple concurrent queries independently', async () => {
                vi.mocked(ProcessChatQuery)
                    .mockResolvedValueOnce('Response 1')
                    .mockResolvedValueOnce('Response 2')
                    .mockResolvedValueOnce('Response 3');

                const [result1, result2, result3] = await Promise.all([
                    chatService.processQuery('query 1'),
                    chatService.processQuery('query 2'),
                    chatService.processQuery('query 3'),
                ]);

                expect(result1).toBe('Response 1');
                expect(result2).toBe('Response 2');
                expect(result3).toBe('Response 3');
            });

            it('one query failure does not affect others', async () => {
                vi.mocked(ProcessChatQuery)
                    .mockResolvedValueOnce('Response 1')
                    .mockRejectedValueOnce(new Error('Query 2 failed'))
                    .mockResolvedValueOnce('Response 3');

                const results = await Promise.allSettled([
                    chatService.processQuery('query 1'),
                    chatService.processQuery('query 2'),
                    chatService.processQuery('query 3'),
                ]);

                expect(results[0].status).toBe('fulfilled');
                expect(results[1].status).toBe('rejected');
                expect(results[2].status).toBe('fulfilled');
            });
        });
    });

    describe('clearContext', () => {
        it('calls ClearChatContext', async () => {
            vi.mocked(ClearChatContext).mockResolvedValue();

            await chatService.clearContext();

            expect(ClearChatContext).toHaveBeenCalledTimes(1);
        });

        it('resolves successfully when ClearChatContext succeeds', async () => {
            vi.mocked(ClearChatContext).mockResolvedValue();

            await expect(chatService.clearContext()).resolves.toBeUndefined();
        });

        it('throws ChatServiceError when ClearChatContext fails', async () => {
            const errorMessage = 'Failed to clear context';
            vi.mocked(ClearChatContext).mockRejectedValue(new Error(errorMessage));

            await expect(chatService.clearContext()).rejects.toThrow(ChatServiceError);
        });

        it('error from clearContext is marked as retryable', async () => {
            vi.mocked(ClearChatContext).mockRejectedValue(new Error('Clear failed'));

            try {
                await chatService.clearContext();
                expect.fail('Should have thrown error');
            } catch (error) {
                expect(error).toBeInstanceOf(ChatServiceError);
                expect((error as ChatServiceError).canRetry).toBe(true);
            }
        });

        it('provides user-friendly error message', async () => {
            vi.mocked(ClearChatContext).mockRejectedValue(new Error('Internal error'));

            try {
                await chatService.clearContext();
                expect.fail('Should have thrown error');
            } catch (error) {
                expect(error).toBeInstanceOf(ChatServiceError);
                expect((error as ChatServiceError).message).toContain('Failed to clear conversation');
            }
        });
    });

    describe('ChatServiceError', () => {
        it('is an instance of Error', () => {
            const error = new ChatServiceError('Test error', true, 'Original');
            expect(error).toBeInstanceOf(Error);
        });

        it('has correct message property', () => {
            const error = new ChatServiceError('Test message', true, 'Original');
            expect(error.message).toBe('Test message');
        });

        it('has canRetry property', () => {
            const retryable = new ChatServiceError('Test', true, 'Original');
            const nonRetryable = new ChatServiceError('Test', false, 'Original');

            expect(retryable.canRetry).toBe(true);
            expect(nonRetryable.canRetry).toBe(false);
        });

        it('has originalError property', () => {
            const error = new ChatServiceError('Test', true, 'Original error message');
            expect(error.originalError).toBe('Original error message');
        });

        it('name property is ChatServiceError', () => {
            const error = new ChatServiceError('Test', true, 'Original');
            expect(error.name).toBe('ChatServiceError');
        });
    });

    describe('Error Classification', () => {
        it('classifies connection errors as retryable', async () => {
            vi.mocked(ProcessChatQuery).mockRejectedValue(new Error('Connection refused'));

            try {
                await chatService.processQuery('test');
                expect.fail('Should have thrown error');
            } catch (error) {
                expect((error as ChatServiceError).canRetry).toBe(true);
            }
        });

        it('classifies service unavailable as retryable', async () => {
            vi.mocked(ProcessChatQuery).mockRejectedValue(new Error('Service unavailable'));

            try {
                await chatService.processQuery('test');
                expect.fail('Should have thrown error');
            } catch (error) {
                expect((error as ChatServiceError).canRetry).toBe(true);
            }
        });

        it('classifies invalid input as non-retryable', async () => {
            vi.mocked(ProcessChatQuery).mockRejectedValue(new Error('Invalid query format'));

            try {
                await chatService.processQuery('test');
                expect.fail('Should have thrown error');
            } catch (error) {
                expect((error as ChatServiceError).canRetry).toBe(false);
            }
        });

        it('classifies max length exceeded as non-retryable', async () => {
            vi.mocked(ProcessChatQuery).mockRejectedValue(new Error('exceeds maximum length'));

            try {
                await chatService.processQuery('test');
                expect.fail('Should have thrown error');
            } catch (error) {
                expect((error as ChatServiceError).canRetry).toBe(false);
            }
        });

        it('defaults unknown errors to retryable', async () => {
            vi.mocked(ProcessChatQuery).mockRejectedValue(new Error('Unknown error XYZ'));

            try {
                await chatService.processQuery('test');
                expect.fail('Should have thrown error');
            } catch (error) {
                expect((error as ChatServiceError).canRetry).toBe(true);
            }
        });
    });
});
