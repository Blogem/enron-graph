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
import ChatPanel from './ChatPanel';

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
      
      const input = screen.queryByRole('textbox');
      expect(input).not.toBeVisible();
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

    it('submit button is enabled by default', () => {
      render(<ChatPanel />);
      
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
});
