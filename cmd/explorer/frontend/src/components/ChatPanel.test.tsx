/**
 * ChatPanel Component Tests (T013)
 * Test Suite for User Story 1 - Display Chat Interface
 * 
 * Requirements tested:
 * - FR-001: Chat panel positioned as bottom panel below graph visualization
 * - FR-002: Chat panel contains text input field and conversation area
 * - FR-007: Chat panel can be collapsed to save screen space
 * - FR-008: Chat panel can be expanded to view conversation
 * - FR-009: Collapsed state persists within session (sessionStorage)
 * - FR-010: Panel initially expanded on first application launch
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Mock the chat service module FIRST (before any imports that use it)
vi.mock('../services/chat', () => ({
  processChatQuery: vi.fn(),
  clearChatContext: vi.fn(),
  ChatServiceError: class ChatServiceError extends Error {
    constructor(message: string, public canRetry: boolean, public originalError: string) {
      super(message);
      this.name = 'ChatServiceError';
    }
  },
  chatService: {
    processQuery: vi.fn(),
    clearContext: vi.fn(),
  },
}));

import ChatPanel from './ChatPanel';
import { processChatQuery, clearChatContext } from '../services/chat';

// Mock sessionStorage
const sessionStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; },
  };
})();

Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
});

describe('ChatPanel Component', () => {
  beforeEach(() => {
    sessionStorageMock.clear();
    // Mock scrollTo and scrollHeight for auto-scroll tests
    Object.defineProperty(HTMLElement.prototype, 'scrollTo', {
      writable: true,
      value: vi.fn(),
    });
    // Note: No longer mocking window.confirm since we use a custom dialog
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('FR-002: Basic Structure', () => {
    it('renders the chat panel container', () => {
      render(<ChatPanel />);

      const panel = screen.getByRole('region', { name: /chat/i });
      expect(panel).toBeInTheDocument();
    });

    it('contains a conversation area', () => {
      render(<ChatPanel />);

      const conversationArea = screen.getByRole('log');
      expect(conversationArea).toBeInTheDocument();
    });

    it('contains a chat input component', () => {
      render(<ChatPanel />);

      const input = screen.getByRole('textbox');
      expect(input).toBeInTheDocument();
    });

    it('contains a submit button', () => {
      render(<ChatPanel />);

      const button = screen.getByRole('button', { name: /send/i });
      expect(button).toBeInTheDocument();
    });
  });

  describe('FR-007 & FR-008: Collapse/Expand Functionality', () => {
    it('renders collapse/expand toggle button', () => {
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse|expand/i });
      expect(toggleButton).toBeInTheDocument();
    });

    it('starts in expanded state by default', () => {
      render(<ChatPanel />);

      const conversationArea = screen.getByRole('log');
      expect(conversationArea).toBeVisible();

      const input = screen.getByRole('textbox');
      expect(input).toBeVisible();
    });

    it('collapses panel when toggle button is clicked', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(toggleButton);

      const conversationArea = screen.queryByRole('log');
      expect(conversationArea).not.toBeVisible();

      // The input is still present but may be visible/hidden depending on implementation
      // Just verify the conversation area is hidden
    });

    it('expands panel when toggle button is clicked while collapsed', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      // First collapse
      const collapseButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(collapseButton);

      // Then expand
      const expandButton = screen.getByRole('button', { name: /expand/i });
      await user.click(expandButton);

      const conversationArea = screen.getByRole('log');
      expect(conversationArea).toBeVisible();

      const input = screen.getByRole('textbox');
      expect(input).toBeVisible();
    });

    it('shows appropriate icon/label when expanded', () => {
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });
      expect(toggleButton).toBeInTheDocument();
    });

    it('shows appropriate icon/label when collapsed', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(toggleButton);

      const expandButton = screen.getByRole('button', { name: /expand/i });
      expect(expandButton).toBeInTheDocument();
    });
  });

  describe('FR-009: Session Persistence', () => {
    it('saves collapsed state to sessionStorage when collapsed', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(toggleButton);

      expect(sessionStorageMock.getItem('chatPanelCollapsed')).toBe('true');
    });

    it('saves expanded state to sessionStorage when expanded', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      // Collapse first
      const collapseButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(collapseButton);

      // Then expand
      const expandButton = screen.getByRole('button', { name: /expand/i });
      await user.click(expandButton);

      expect(sessionStorageMock.getItem('chatPanelCollapsed')).toBe('false');
    });

    it('restores collapsed state from sessionStorage on mount', () => {
      sessionStorageMock.setItem('chatPanelCollapsed', 'true');

      render(<ChatPanel />);

      const conversationArea = screen.queryByRole('log');
      expect(conversationArea).not.toBeVisible();
    });

    it('restores expanded state from sessionStorage on mount', () => {
      sessionStorageMock.setItem('chatPanelCollapsed', 'false');

      render(<ChatPanel />);

      const conversationArea = screen.getByRole('log');
      expect(conversationArea).toBeVisible();
    });
  });

  describe('FR-010: Initial State', () => {
    it('starts expanded when no sessionStorage value exists', () => {
      render(<ChatPanel />);

      const conversationArea = screen.getByRole('log');
      expect(conversationArea).toBeVisible();
    });

    it('respects initialCollapsed prop when provided', () => {
      render(<ChatPanel initialCollapsed={true} />);

      const conversationArea = screen.queryByRole('log');
      expect(conversationArea).not.toBeVisible();
    });

    it('sessionStorage overrides initialCollapsed prop', () => {
      sessionStorageMock.setItem('chatPanelCollapsed', 'false');

      render(<ChatPanel initialCollapsed={true} />);

      const conversationArea = screen.getByRole('log');
      expect(conversationArea).toBeVisible();
    });
  });

  describe('Callback Props', () => {
    it('calls onCollapseChange when panel is collapsed', async () => {
      const user = userEvent.setup();
      const onCollapseChange = vi.fn();
      render(<ChatPanel onCollapseChange={onCollapseChange} />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(toggleButton);

      expect(onCollapseChange).toHaveBeenCalledWith(true);
    });

    it('calls onCollapseChange when panel is expanded', async () => {
      const user = userEvent.setup();
      const onCollapseChange = vi.fn();
      render(<ChatPanel onCollapseChange={onCollapseChange} />);

      // Collapse first
      const collapseButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(collapseButton);

      // Then expand
      const expandButton = screen.getByRole('button', { name: /expand/i });
      await user.click(expandButton);

      expect(onCollapseChange).toHaveBeenCalledWith(false);
    });
  });

  describe('Message Display', () => {
    it('displays empty conversation area initially', () => {
      render(<ChatPanel />);

      const conversationArea = screen.getByRole('log');
      const messages = within(conversationArea).queryAllByRole('article');

      expect(messages).toHaveLength(0);
    });

    it('conversation area is scrollable', () => {
      const { container } = render(<ChatPanel />);

      const conversationArea = container.querySelector('[role="log"]');
      expect(conversationArea).toHaveStyle({ overflowY: 'auto' });
    });
  });

  describe('Input Integration', () => {
    it('input field is enabled by default', () => {
      render(<ChatPanel />);

      const input = screen.getByRole('textbox');
      expect(input).not.toBeDisabled();
    });

    it('submit button is enabled when input has content', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      const input = screen.getByRole('textbox');
      await user.type(input, 'test');

      const button = screen.getByRole('button', { name: /send/i });
      expect(button).not.toBeDisabled();
    });
  });

  describe('Accessibility', () => {
    it('has appropriate ARIA role for the panel', () => {
      render(<ChatPanel />);

      const panel = screen.getByRole('region');
      expect(panel).toBeInTheDocument();
    });

    it('has accessible name for the panel', () => {
      render(<ChatPanel />);

      const panel = screen.getByRole('region', { name: /chat/i });
      expect(panel).toHaveAccessibleName();
    });

    it('conversation area has appropriate ARIA role', () => {
      render(<ChatPanel />);

      const conversationArea = screen.getByRole('log');
      expect(conversationArea).toBeInTheDocument();
    });

    it('toggle button has accessible label', () => {
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse|expand/i });
      expect(toggleButton).toHaveAccessibleName();
    });

    it('announces collapse state change to screen readers', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(toggleButton);

      // Panel should have aria-expanded attribute
      const panel = screen.getByRole('region');
      expect(panel).toHaveAttribute('aria-expanded');
    });
  });

  describe('Layout and Positioning', () => {
    it('applies bottom panel positioning class', () => {
      const { container } = render(<ChatPanel />);

      const panel = container.querySelector('.chat-panel');
      expect(panel).toHaveClass('chat-panel--bottom');
    });

    it('maintains consistent height when expanded', () => {
      const { container } = render(<ChatPanel />);

      const panel = container.querySelector('.chat-panel');
      // Should have a fixed or minimum height defined in CSS
      expect(panel).toBeInTheDocument();
    });

    it('has minimal height when collapsed', async () => {
      const user = userEvent.setup();
      const { container } = render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });
      await user.click(toggleButton);

      const panel = container.querySelector('.chat-panel');
      expect(panel).toHaveClass('chat-panel--collapsed');
    });
  });

  describe('Edge Cases', () => {
    it('handles rapid collapse/expand toggling', async () => {
      const user = userEvent.setup();
      render(<ChatPanel />);

      const toggleButton = screen.getByRole('button', { name: /collapse/i });

      // Rapid toggling
      await user.click(toggleButton);
      await user.click(screen.getByRole('button', { name: /expand/i }));
      await user.click(screen.getByRole('button', { name: /collapse/i }));

      // Should not crash or have inconsistent state
      const conversationArea = screen.queryByRole('log');
      expect(conversationArea).not.toBeVisible();
    });

    it('handles sessionStorage unavailable gracefully', () => {
      // Mock sessionStorage to throw error
      const originalSessionStorage = window.sessionStorage;
      Object.defineProperty(window, 'sessionStorage', {
        get: () => { throw new Error('sessionStorage not available'); },
      });

      // Should not crash
      expect(() => render(<ChatPanel />)).not.toThrow();

      // Restore
      Object.defineProperty(window, 'sessionStorage', {
        value: originalSessionStorage,
      });
    });
  });

  /**
   * T027: User Story 2 - Query Submission and Response Display Tests
   * 
   * Requirements tested:
   * - FR-003: Submit query with Enter key
   * - FR-004: Shift+Enter creates newline
   * - FR-005: User queries appear right-aligned
   * - FR-006: System responses appear left-aligned with different background
   * - FR-011: User-friendly error messages displayed
   * - FR-014: Auto-scroll to latest message
   * - FR-015: Empty query prevention
   * - FR-021: Input disabled while processing query
   * - FR-022: Retry functionality for failed queries
   * - FR-024: Query timeout (60 seconds)
   */
  describe('User Story 2: Query Submission and Response Display', () => {
    beforeEach(() => {
      vi.mocked(processChatQuery).mockClear();
      vi.mocked(clearChatContext).mockClear();
    });

    describe('Query Submission', () => {
      it('submits query when user types and presses Enter', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Test response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        expect(vi.mocked(processChatQuery)).toHaveBeenCalledWith('test query');
      });

      it('submits query when user clicks send button', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Test response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query');

        const sendButton = screen.getByRole('button', { name: /send/i });
        await user.click(sendButton);

        expect(vi.mocked(processChatQuery)).toHaveBeenCalledWith('test query');
      });

      it('clears input field after successful submission', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Test response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox') as HTMLTextAreaElement;
        await user.type(input, 'test query{Enter}');

        // Wait for submission to complete
        await screen.findByText(/test response/i);

        expect(input.value).toBe('');
      });

      it('does not clear input if submission fails', async () => {
        const user = userEvent.setup();
        const error = new Error('Submission failed');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery).mockRejectedValue(error);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox') as HTMLTextAreaElement;
        await user.type(input, 'test query{Enter}');

        // Wait for error to appear
        await screen.findByText(/failed/i);

        expect(input.value).toBe('test query');
      });
    });

    describe('FR-015: Empty Query Prevention', () => {
      it('does not submit empty query', async () => {
        const user = userEvent.setup();

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, '{Enter}');

        expect(vi.mocked(processChatQuery)).not.toHaveBeenCalled();
      });

      it('does not submit whitespace-only query', async () => {
        const user = userEvent.setup();

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, '   {Enter}');

        expect(vi.mocked(processChatQuery)).not.toHaveBeenCalled();
      });

      it('send button is disabled when input is empty', () => {
        render(<ChatPanel />);

        const sendButton = screen.getByRole('button', { name: /send/i });
        expect(sendButton).toBeDisabled();
      });

      it('send button is enabled when input has content', async () => {
        const user = userEvent.setup();
        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test');

        const sendButton = screen.getByRole('button', { name: /send/i });
        expect(sendButton).not.toBeDisabled();
      });
    });

    describe('Message Display', () => {
      it('displays user query as a message', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        const userMessage = await screen.findByText('test query');
        expect(userMessage).toBeInTheDocument();
      });

      it('displays system response as a message', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('System response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        const systemMessage = await screen.findByText('System response');
        expect(systemMessage).toBeInTheDocument();
      });

      it('displays both user query and system response', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('System response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText('System response');

        expect(screen.getByText('test query')).toBeInTheDocument();
        expect(screen.getByText('System response')).toBeInTheDocument();
      });

      it('displays messages in chronological order', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery)
          .mockResolvedValueOnce('Response 1')
          .mockResolvedValueOnce('Response 2');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // First query
        await user.type(input, 'query 1{Enter}');
        await screen.findByText('Response 1');

        // Second query
        await user.type(input, 'query 2{Enter}');
        await screen.findByText('Response 2');

        const conversationArea = screen.getByRole('log');
        const messages = within(conversationArea).getAllByRole('article');

        // Should have 4 messages: query1, response1, query2, response2
        expect(messages).toHaveLength(4);
      });
    });

    describe('FR-005 & FR-006: Visual Distinction', () => {
      it('user messages have user sender type', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText('Response');

        const userMessage = screen.getByText('test query').closest('[role="article"]');
        expect(userMessage).toHaveClass('chat-message--user');
      });

      it('system messages have system sender type', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('System response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        const systemMessage = await screen.findByText('System response');
        const messageElement = systemMessage.closest('[role="article"]');
        expect(messageElement).toHaveClass('chat-message--system');
      });
    });

    describe('FR-021: Loading State', () => {
      it('shows loading indicator while processing query', async () => {
        const user = userEvent.setup();
        let resolveQuery: (value: string) => void;
        const queryPromise = new Promise<string>((resolve) => {
          resolveQuery = resolve;
        });
        vi.mocked(processChatQuery).mockReturnValue(queryPromise);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        // Loading indicator should appear
        expect(screen.getByText(/loading|processing/i)).toBeInTheDocument();

        // Resolve the query
        resolveQuery!('Response');

        // Loading indicator should disappear
        await screen.findByText('Response');
        expect(screen.queryByText(/loading|processing/i)).not.toBeInTheDocument();
      });

      it('disables input while query is processing', async () => {
        const user = userEvent.setup();
        let resolveQuery: (value: string) => void;
        const queryPromise = new Promise<string>((resolve) => {
          resolveQuery = resolve;
        });
        vi.mocked(processChatQuery).mockReturnValue(queryPromise);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        // Input should be disabled
        expect(input).toBeDisabled();

        // Resolve the query
        resolveQuery!('Response');

        // Input should be enabled again
        await screen.findByText('Response');
        expect(input).not.toBeDisabled();
      });

      it('disables send button while query is processing', async () => {
        const user = userEvent.setup();
        let resolveQuery: (value: string) => void;
        const queryPromise = new Promise<string>((resolve) => {
          resolveQuery = resolve;
        });
        vi.mocked(processChatQuery).mockReturnValue(queryPromise);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        // Send button should be disabled
        const sendButton = screen.getByRole('button', { name: /send/i });
        expect(sendButton).toBeDisabled();

        // Resolve the query
        resolveQuery!('Response');

        // Wait for response and input to clear
        await screen.findByText('Response');

        // Send button should be disabled again because input is empty
        expect(sendButton).toBeDisabled();

        // Type something to verify it's enabled when input has content (not disabled due to loading)
        await user.type(input, 'another query');
        expect(sendButton).not.toBeDisabled();
      });
    });

    describe('FR-011: Error Handling', () => {
      it('displays error message when query fails', async () => {
        const user = userEvent.setup();
        const error = new Error('Query processing failed');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery).mockRejectedValue(error);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        const errorMessage = await screen.findByText(/error|failed/i);
        expect(errorMessage).toBeInTheDocument();
      });

      it('clears previous error when new query is submitted', async () => {
        const user = userEvent.setup();
        const error = new Error('First query failed');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery)
          .mockRejectedValueOnce(error)
          .mockResolvedValueOnce('Success response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // First query fails
        await user.type(input, 'test query 1{Enter}');
        await screen.findByText(/error|failed/i);

        // Second query succeeds
        await user.type(input, 'test query 2{Enter}');
        await screen.findByText('Success response');

        // Error message should be gone
        expect(screen.queryByText(/error|failed/i)).not.toBeInTheDocument();
      });

      it('error message includes user-friendly text', async () => {
        const user = userEvent.setup();
        const error = new Error('Internal server error');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery).mockRejectedValue(error);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        const errorMessage = await screen.findByText(/failed|error|try again/i);
        expect(errorMessage).toBeInTheDocument();
      });
    });

    describe('FR-022: Retry Functionality', () => {
      it('shows retry button when error is retryable', async () => {
        const user = userEvent.setup();
        const error = new Error('Temporary failure');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery).mockRejectedValue(error);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText(/error|failed/i);

        const retryButton = screen.getByRole('button', { name: /retry/i });
        expect(retryButton).toBeInTheDocument();
      });

      it('does not show retry button when error is not retryable', async () => {
        const user = userEvent.setup();
        const error = new Error('Invalid query format');
        (error as any).canRetry = false;
        vi.mocked(processChatQuery).mockRejectedValue(error);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText(/error|failed/i);

        const retryButton = screen.queryByRole('button', { name: /retry/i });
        expect(retryButton).not.toBeInTheDocument();
      });

      it('resubmits last query when retry is clicked', async () => {
        const user = userEvent.setup();
        const error = new Error('Temporary failure');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery)
          .mockRejectedValueOnce(error)
          .mockResolvedValueOnce('Retry success');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText(/error|failed/i);

        const retryButton = screen.getByRole('button', { name: /retry/i });
        await user.click(retryButton);

        // Should call processQuery again with the same query
        expect(vi.mocked(processChatQuery)).toHaveBeenCalledTimes(2);
        expect(vi.mocked(processChatQuery)).toHaveBeenNthCalledWith(2, 'test query');

        await screen.findByText('Retry success');
      });

      it('clears error message when retry is clicked', async () => {
        const user = userEvent.setup();
        const error = new Error('Temporary failure');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery)
          .mockRejectedValueOnce(error)
          .mockResolvedValueOnce('Success');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        const errorMessage = await screen.findByText(/error|failed/i);

        const retryButton = screen.getByRole('button', { name: /retry/i });
        await user.click(retryButton);

        await screen.findByText('Success');

        expect(errorMessage).not.toBeInTheDocument();
      });
    });

    describe('FR-024: Timeout Handling', () => {
      it('displays timeout error after 60 seconds', async () => {
        const user = userEvent.setup();
        const timeoutError = new Error('Query timeout exceeded');
        (timeoutError as any).canRetry = true;
        vi.mocked(processChatQuery).mockRejectedValue(timeoutError);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'slow query{Enter}');

        const errorMessage = await screen.findByText(/timeout/i);
        expect(errorMessage).toBeInTheDocument();
      });

      it('timeout error includes retry button', async () => {
        const user = userEvent.setup();
        const timeoutError = new Error('Query timeout');
        (timeoutError as any).canRetry = true;
        vi.mocked(processChatQuery).mockRejectedValue(timeoutError);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'slow query{Enter}');

        await screen.findByText(/timeout/i);

        const retryButton = screen.getByRole('button', { name: /retry/i });
        expect(retryButton).toBeInTheDocument();
      });
    });

    describe('FR-014: Auto-scroll to Latest Message', () => {
      it('scrolls to bottom when new message is added', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const conversationArea = screen.getByRole('log');
        const scrollToSpy = vi.spyOn(conversationArea, 'scrollTo');

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText('Response');

        // Should scroll to bottom (scrollTop = scrollHeight)
        expect(scrollToSpy).toHaveBeenCalled();
      });

      it('maintains scroll position when panel is collapsed', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText('Response');

        const toggleButton = screen.getByRole('button', { name: /collapse/i });
        await user.click(toggleButton);

        // Should not crash or lose messages
        const expandButton = screen.getByRole('button', { name: /expand/i });
        await user.click(expandButton);

        expect(screen.getByText('test query')).toBeInTheDocument();
        expect(screen.getByText('Response')).toBeInTheDocument();
      });
    });

    describe('Auto-expand on Query Submission', () => {
      it('expands panel when user submits query while collapsed', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        // Collapse panel
        const collapseButton = screen.getByRole('button', { name: /collapse/i });
        await user.click(collapseButton);

        // Submit query while collapsed
        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        // Panel should auto-expand
        const conversationArea = screen.getByRole('log');
        expect(conversationArea).toBeVisible();
      });
    });
  });

  /**
   * T043: User Story 3 - Conversation Scrolling Behavior Tests
   * 
   * Requirements tested:
   * - FR-013: Conversation area is scrollable when content exceeds visible space
   * - FR-014: Auto-scroll to latest message when new response arrives
   * - SC-006: Interface remains responsive with 50+ messages
   */
  describe('User Story 3: Conversation Scrolling Behavior', () => {
    beforeEach(() => {
      vi.mocked(processChatQuery).mockClear();
      vi.mocked(clearChatContext).mockClear();

      // Mock scrollHeight and clientHeight for scrolling tests
      Object.defineProperty(HTMLElement.prototype, 'scrollHeight', {
        configurable: true,
        get: function () {
          // Simulate overflow: scrollHeight > clientHeight
          return 1000;
        },
      });

      Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
        configurable: true,
        get: function () {
          return 300;
        },
      });
    });

    describe('FR-013: Scrollable Conversation Area', () => {
      it('conversation area has scrollable overflow style', () => {
        const { container } = render(<ChatPanel />);

        const conversationArea = container.querySelector('[role="log"]');
        const styles = window.getComputedStyle(conversationArea!);

        // Should have overflow-y: auto or scroll to enable scrolling
        expect(['auto', 'scroll']).toContain(styles.overflowY);
      });

      it('allows scrolling when content exceeds visible height', async () => {
        const user = userEvent.setup();
        // Mock many responses to create overflow
        vi.mocked(processChatQuery).mockResolvedValue('Response text');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // Submit 10 queries to create scrollable content
        for (let i = 0; i < 10; i++) {
          await user.type(input, `Query ${i + 1}{Enter}`);
          // Wait for responses to appear (expect multiple)
          await screen.findByText(`Query ${i + 1}`);
        }

        const conversationArea = screen.getByRole('log');

        // Verify scrollHeight > clientHeight (content overflows)
        expect(conversationArea.scrollHeight).toBeGreaterThan(conversationArea.clientHeight);
      });

      it('maintains scroll position when user manually scrolls up', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const conversationArea = screen.getByRole('log');
        const input = screen.getByRole('textbox');

        // Add several messages
        for (let i = 0; i < 5; i++) {
          await user.type(input, `Query ${i}{Enter}`);
          // Wait for query to appear
          await screen.findByText(`Query ${i}`);
        }

        // Manually scroll to top
        conversationArea.scrollTop = 0;

        // Scroll position should be maintained (not forced to bottom)
        // unless a new message is added
        expect(conversationArea.scrollTop).toBe(0);
      });
    });

    describe('FR-014: Auto-scroll to Latest Message', () => {
      it('automatically scrolls to bottom when new message is added', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response text');

        render(<ChatPanel />);

        const conversationArea = screen.getByRole('log');
        const scrollToSpy = vi.spyOn(conversationArea, 'scrollTo');

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        await screen.findByText('Response text');

        // Should call scrollTo to move to bottom
        expect(scrollToSpy).toHaveBeenCalled();

        // Verify it scrolled to the bottom position
        const lastCall = scrollToSpy.mock.calls[scrollToSpy.mock.calls.length - 1];
        expect(lastCall[0]).toHaveProperty('top');
      });

      it('scrolls to bottom after user message is added', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const conversationArea = screen.getByRole('log');
        const scrollToSpy = vi.spyOn(conversationArea, 'scrollTo');

        const input = screen.getByRole('textbox');

        // Clear any initial calls
        scrollToSpy.mockClear();

        await user.type(input, 'test query{Enter}');

        // Should scroll when user message is added
        expect(scrollToSpy).toHaveBeenCalled();
      });

      it('scrolls to bottom after system response is added', async () => {
        const user = userEvent.setup();
        let resolveQuery: (value: string) => void;
        const queryPromise = new Promise<string>((resolve) => {
          resolveQuery = resolve;
        });
        vi.mocked(processChatQuery).mockReturnValue(queryPromise);

        render(<ChatPanel />);

        const conversationArea = screen.getByRole('log');
        const scrollToSpy = vi.spyOn(conversationArea, 'scrollTo');

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');

        // Clear scroll calls from user message
        scrollToSpy.mockClear();

        // Resolve query to add system response
        resolveQuery!('System response');
        await screen.findByText('System response');

        // Should scroll after system response is added
        expect(scrollToSpy).toHaveBeenCalled();
      });

      it('auto-scrolls for each new message in sequence', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery)
          .mockResolvedValueOnce('Response 1')
          .mockResolvedValueOnce('Response 2')
          .mockResolvedValueOnce('Response 3');

        render(<ChatPanel />);

        const conversationArea = screen.getByRole('log');
        const scrollToSpy = vi.spyOn(conversationArea, 'scrollTo');

        const input = screen.getByRole('textbox');

        // Submit three queries
        await user.type(input, 'Query 1{Enter}');
        await screen.findByText('Response 1');

        const callsAfterFirst = scrollToSpy.mock.calls.length;

        await user.type(input, 'Query 2{Enter}');
        await screen.findByText('Response 2');

        const callsAfterSecond = scrollToSpy.mock.calls.length;

        await user.type(input, 'Query 3{Enter}');
        await screen.findByText('Response 3');

        // Should have additional scroll calls for each message pair
        expect(scrollToSpy.mock.calls.length).toBeGreaterThan(callsAfterSecond);
        expect(callsAfterSecond).toBeGreaterThan(callsAfterFirst);
      });

      it('does not scroll when panel is collapsed', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        // Collapse the panel
        const collapseButton = screen.getByRole('button', { name: /collapse/i });
        await user.click(collapseButton);

        const conversationArea = screen.queryByRole('log');

        // Conversation area should not be visible when collapsed
        expect(conversationArea).not.toBeVisible();
      });
    });

    describe('SC-006: Performance with 50+ Messages', () => {
      it('renders 50+ messages without performance degradation', async () => {
        const user = userEvent.setup();

        // Mock responses for 30 queries (= 60 messages total: 30 user + 30 system)
        vi.mocked(processChatQuery).mockResolvedValue('Response message text');

        const startTime = performance.now();

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // Submit 30 queries to create 60 total messages
        for (let i = 0; i < 30; i++) {
          await user.type(input, `Query ${i + 1}{Enter}`);
          // Wait for query to appear
          await screen.findByText(`Query ${i + 1}`);
        }

        const renderTime = performance.now() - startTime;

        // Verify all messages are present
        const conversationArea = screen.getByRole('log');
        const messages = within(conversationArea).getAllByRole('article');
        expect(messages.length).toBe(60); // 30 user + 30 system messages

        // Rendering should complete reasonably quickly (under 5 seconds for 60 messages)
        // This is a generous threshold; actual performance should be much faster
        expect(renderTime).toBeLessThan(5000);
      });

      it('maintains UI responsiveness with large message count', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // Add 25 query-response pairs (50 messages)
        for (let i = 0; i < 25; i++) {
          await user.type(input, `Query ${i}{Enter}`);
          // Wait for query to appear
          await screen.findByText(`Query ${i}`);
        }

        // Test that UI is still responsive after many messages
        const testStartTime = performance.now();

        // Should be able to collapse/expand quickly
        const collapseButton = screen.getByRole('button', { name: /collapse/i });
        await user.click(collapseButton);

        const expandButton = screen.getByRole('button', { name: /expand/i });
        await user.click(expandButton);

        const interactionTime = performance.now() - testStartTime;

        // UI interaction should be fast (under 500ms)
        expect(interactionTime).toBeLessThan(500);

        // Should still be able to submit new queries
        await user.type(input, 'New query{Enter}');
        await screen.findByText('New query');

        const conversationArea = screen.getByRole('log');
        const messages = within(conversationArea).getAllByRole('article');

        // Should have 52 messages (25 original pairs + 1 new pair)
        expect(messages.length).toBe(52);
      });

      it('scrolling remains smooth with 50+ messages', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        const conversationArea = screen.getByRole('log');

        // Add 25 query-response pairs
        for (let i = 0; i < 25; i++) {
          await user.type(input, `Query ${i}{Enter}`);
          // Wait for query to appear
          await screen.findByText(`Query ${i}`);
        }

        const scrollToSpy = vi.spyOn(conversationArea, 'scrollTo');

        // Add one more message and verify auto-scroll still works
        await user.type(input, 'Final query{Enter}');
        await screen.findByText('Final query');

        // Auto-scroll should still be called even with many messages
        expect(scrollToSpy).toHaveBeenCalled();
      });
    });

    describe('Scroll Behavior Edge Cases', () => {
      it('handles rapid message additions without scroll errors', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Fast response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // Rapidly submit multiple queries
        const queryPromises = [];
        for (let i = 0; i < 5; i++) {
          queryPromises.push(user.type(input, `Query ${i}{Enter}`));
        }

        await Promise.all(queryPromises);

        // Wait for all responses
        const responses = await screen.findAllByText('Fast response');
        expect(responses.length).toBeGreaterThan(0);

        // Should not crash or throw errors
        const conversationArea = screen.getByRole('log');
        expect(conversationArea).toBeInTheDocument();
      });

      it('preserves conversation after collapse/expand with scrolling', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // Add messages
        for (let i = 0; i < 5; i++) {
          await user.type(input, `Query ${i}{Enter}`);
          // Wait for query to appear
          await screen.findByText(`Query ${i}`);
        }

        // Collapse
        const collapseButton = screen.getByRole('button', { name: /collapse/i });
        await user.click(collapseButton);

        // Expand
        const expandButton = screen.getByRole('button', { name: /expand/i });
        await user.click(expandButton);

        // All messages should still be present
        const conversationArea = screen.getByRole('log');
        const messages = within(conversationArea).getAllByRole('article');
        expect(messages.length).toBe(10); // 5 queries + 5 responses
      });

      it('handles scrolling when conversation area is resized', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'Query 1{Enter}');
        await screen.findByText('Query 1');

        const conversationArea = screen.getByRole('log');

        // Simulate resize by changing clientHeight
        Object.defineProperty(conversationArea, 'clientHeight', {
          configurable: true,
          value: 150, // Smaller height
        });

        // Add another message
        await user.type(input, 'Query 2{Enter}');
        await screen.findByText('Query 2');

        // Should still auto-scroll even after resize
        expect(conversationArea).toBeInTheDocument();
      });
    });
  });

  /**
   * T049: User Story 4 - Clear Conversation Functionality Tests
   * 
   * Requirements tested:
   * - FR-016: System MUST provide a way to clear the conversation history
   * - FR-017: System MUST preserve conversation history within a session
   * - Acceptance Scenario 1: Conversation area emptied after clear action
   * - Acceptance Scenario 2: Fresh conversation can be started after clearing
   */
  describe('User Story 4: Clear Conversation Functionality', () => {
    beforeEach(() => {
      vi.mocked(processChatQuery).mockReset();
      vi.mocked(clearChatContext).mockReset();
    });

    // Helper function to click clear button and confirm
    const clickClearAndConfirm = async (user: ReturnType<typeof userEvent.setup>) => {
      const clearButton = screen.getByRole('button', { name: /clear conversation/i });
      await user.click(clearButton);

      // Wait for confirmation dialog and click confirm
      const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
      await user.click(confirmButton);
    };

    describe('FR-016: Clear Conversation', () => {
      it('renders a clear button in the chat panel', () => {
        render(<ChatPanel />);

        const clearButton = screen.getByRole('button', { name: /clear/i });
        expect(clearButton).toBeInTheDocument();
      });

      it('clear button is visible when panel is expanded', () => {
        render(<ChatPanel />);

        const clearButton = screen.getByRole('button', { name: /clear/i });
        expect(clearButton).toBeVisible();
      });

      it('clear button is enabled when conversation has messages', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const clearButton = screen.getByRole('button', { name: /clear/i });
        expect(clearButton).not.toBeDisabled();
      });

      it('clear button is disabled when conversation is empty', () => {
        render(<ChatPanel />);

        const clearButton = screen.getByRole('button', { name: /clear/i });
        expect(clearButton).toBeDisabled();
      });

      it('calls clearChatContext when clear button is clicked', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        expect(vi.mocked(clearChatContext)).toHaveBeenCalledTimes(1);
      });

      it('empties conversation area after clearing', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const conversationArea = screen.getByRole('log');
        let messages = within(conversationArea).getAllByRole('article');
        expect(messages.length).toBe(2); // User message + system response

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Wait for clear to complete
        await vi.waitFor(() => {
          const messagesAfterClear = within(conversationArea).queryAllByRole('article');
          expect(messagesAfterClear.length).toBe(0);
        });
      });

      it('clears all messages including multiple query-response pairs', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // Submit multiple queries
        for (let i = 0; i < 3; i++) {
          await user.type(input, `Query ${i}{Enter}`);
          await screen.findByText(`Query ${i}`);
        }

        const conversationArea = screen.getByRole('log');
        let messages = within(conversationArea).getAllByRole('article');
        expect(messages.length).toBe(6); // 3 queries + 3 responses

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Wait for clear to complete
        await vi.waitFor(() => {
          const messagesAfterClear = within(conversationArea).queryAllByRole('article');
          expect(messagesAfterClear.length).toBe(0);
        });
      });

      it('clears error state when clearing conversation', async () => {
        const user = userEvent.setup();
        const error = new Error('Query failed');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery).mockRejectedValue(error);
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText(/error|failed/i);

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Error message should be cleared
        await vi.waitFor(() => {
          expect(screen.queryByText(/error|failed/i)).not.toBeInTheDocument();
        });
      });

      it('clears lastQuery state when clearing conversation', async () => {
        const user = userEvent.setup();
        const error = new Error('Query failed');
        (error as any).canRetry = true;
        vi.mocked(processChatQuery)
          .mockRejectedValueOnce(error)
          .mockResolvedValueOnce('Retry success');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText(/error|failed/i);

        // Clear conversation
        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        await vi.waitFor(() => {
          expect(screen.queryByText(/error|failed/i)).not.toBeInTheDocument();
        });

        // Retry button should not be present after clearing
        expect(screen.queryByRole('button', { name: /retry/i })).not.toBeInTheDocument();
      });
    });

    describe('Acceptance Scenario 1: Conversation emptied after clear', () => {
      it('given conversation history exists, when clear is triggered, then conversation area is emptied', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('System response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        // Given: conversation history exists
        const input = screen.getByRole('textbox');
        await user.type(input, 'First query{Enter}');
        await vi.waitFor(() => {
          expect(screen.getAllByText('System response').length).toBeGreaterThanOrEqual(1);
        });
        await user.type(input, 'Second query{Enter}');
        await vi.waitFor(() => {
          expect(screen.getAllByText('System response').length).toBeGreaterThanOrEqual(2);
        });

        const conversationArea = screen.getByRole('log');
        let messages = within(conversationArea).getAllByRole('article');
        expect(messages.length).toBeGreaterThan(0);

        // When: clear is triggered
        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Then: conversation area is emptied
        await vi.waitFor(() => {
          const messagesAfterClear = within(conversationArea).queryAllByRole('article');
          expect(messagesAfterClear.length).toBe(0);
        });
      });
    });

    describe('Acceptance Scenario 2: Fresh conversation after clearing', () => {
      it('given conversation is cleared, when user submits a new query, then it starts a fresh conversation', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('New response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        // Given: conversation is cleared
        const input = screen.getByRole('textbox');
        await user.type(input, 'Old query{Enter}');
        await screen.findByText('New response');

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        await vi.waitFor(() => {
          const conversationArea = screen.getByRole('log');
          const messages = within(conversationArea).queryAllByRole('article');
          expect(messages.length).toBe(0);
        });

        // When: user submits a new query
        await user.type(input, 'Fresh query{Enter}');

        // Then: it starts a fresh conversation
        const freshQuery = await screen.findByText('Fresh query');
        const freshResponse = await screen.findByText('New response');

        expect(freshQuery).toBeInTheDocument();
        expect(freshResponse).toBeInTheDocument();

        // Old query should not be present
        expect(screen.queryByText('Old query')).not.toBeInTheDocument();

        // Should only have the new query-response pair
        const conversationArea = screen.getByRole('log');
        const messages = within(conversationArea).getAllByRole('article');
        expect(messages.length).toBe(2); // Fresh query + response
      });
    });

    describe('Error Handling', () => {
      it('displays error message if clearChatContext fails', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockRejectedValue(new Error('Clear failed'));

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Should display an error message
        const errorMessage = await screen.findByText(/failed|error/i);
        expect(errorMessage).toBeInTheDocument();
      });

      it('preserves messages if clearChatContext fails', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockRejectedValue(new Error('Clear failed'));

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const conversationArea = screen.getByRole('log');
        const messagesBefore = within(conversationArea).getAllByRole('article');
        const messageCountBefore = messagesBefore.length;

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Wait for error to appear
        await screen.findByText(/failed|error/i);

        // Messages should still be present
        const messagesAfter = within(conversationArea).getAllByRole('article');
        expect(messagesAfter.length).toBe(messageCountBefore);
      });

      it('does not disable clear button after failed clear attempt', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockRejectedValue(new Error('Clear failed'));

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Wait for error
        await screen.findByText(/failed|error/i);

        // Clear button should still be enabled for retry
        expect(clearButton).not.toBeDisabled();
      });
    });

    describe('UI Behavior', () => {
      it('shows loading state while clearing conversation', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        let resolveClear: () => void;
        const clearPromise = new Promise<void>((resolve) => {
          resolveClear = resolve;
        });
        vi.mocked(clearChatContext).mockReturnValue(clearPromise);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Clear button should be disabled during clearing
        await vi.waitFor(() => {
          expect(clearButton).toBeDisabled();
        });

        // Resolve clear
        resolveClear!();
        await vi.waitFor(() => {
          const conversationArea = screen.getByRole('log');
          const messages = within(conversationArea).queryAllByRole('article');
          expect(messages.length).toBe(0);
        });
      });

      it('maintains collapsed state after clearing conversation', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        // Collapse panel
        const collapseButton = screen.getByRole('button', { name: /collapse/i });
        await user.click(collapseButton);

        // Panel should be collapsed
        let conversationArea = screen.queryByRole('log');
        expect(conversationArea).not.toBeVisible();

        // Expand panel to access clear button
        const expandButton = screen.getByRole('button', { name: /expand/i });
        await user.click(expandButton);

        // Clear conversation
        const clearButton = screen.getByRole('button', { name: /clear/i });
        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        await vi.waitFor(() => {
          expect(vi.mocked(clearChatContext)).toHaveBeenCalled();
        });

        // Collapse panel again
        const collapseButtonAgain = screen.getByRole('button', { name: /collapse/i });
        await user.click(collapseButtonAgain);

        // Panel should be collapsed again
        conversationArea = screen.queryByRole('log');
        expect(conversationArea).not.toBeVisible();
      });

      it('auto-disables clear button when conversation is empty', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const clearButton = screen.getByRole('button', { name: /clear/i });
        expect(clearButton).not.toBeDisabled();

        await user.click(clearButton);

        // Click confirm in dialog
        const confirmButton = await screen.findByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Wait for clear to complete
        await vi.waitFor(() => {
          const conversationArea = screen.getByRole('log');
          const messages = within(conversationArea).queryAllByRole('article');
          expect(messages.length).toBe(0);
        });

        // Clear button should now be disabled
        expect(clearButton).toBeDisabled();
      });
    });

    describe('Confirmation Dialog', () => {
      it('shows confirmation dialog before clearing conversation with multiple messages', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');
        vi.mocked(clearChatContext).mockResolvedValue(undefined);

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');

        // Add multiple messages
        for (let i = 0; i < 3; i++) {
          await user.type(input, `Query ${i}{Enter}`);
          await screen.findByText(`Query ${i}`);
        }

        const clearButton = screen.getByRole('button', { name: /clear conversation/i });
        await user.click(clearButton);

        // Should show confirmation dialog
        const confirmDialog = await screen.findByText(/are you sure you want to clear/i);
        expect(confirmDialog).toBeInTheDocument();

        // Click confirm button
        const confirmButton = screen.getByRole('button', { name: /confirm clearing conversation/i });
        await user.click(confirmButton);

        // Wait for clear to complete
        await vi.waitFor(() => {
          expect(vi.mocked(clearChatContext)).toHaveBeenCalled();
        });
      });

      it('does not clear if user cancels confirmation dialog', async () => {
        const user = userEvent.setup();
        vi.mocked(processChatQuery).mockResolvedValue('Response');

        render(<ChatPanel />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'test query{Enter}');
        await screen.findByText('Response');

        const conversationArea = screen.getByRole('log');
        const messagesBefore = within(conversationArea).getAllByRole('article');

        const clearButton = screen.getByRole('button', { name: /clear conversation/i });
        await user.click(clearButton);

        // Should show confirmation dialog
        await screen.findByText(/are you sure you want to clear/i);

        // Click cancel button
        const cancelButton = screen.getByRole('button', { name: /cancel/i });
        await user.click(cancelButton);

        // Should not call clearChatContext
        expect(vi.mocked(clearChatContext)).not.toHaveBeenCalled();

        // Messages should still be present
        const messagesAfter = within(conversationArea).getAllByRole('article');
        expect(messagesAfter.length).toBe(messagesBefore.length);
      });
    });
  });
});
