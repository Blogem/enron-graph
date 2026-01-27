/**
 * ChatMessage Component Tests (T012)
 * Test Suite for User Story 1 - Display Chat Interface
 * 
 * Requirements tested:
 * - FR-005: Chat conversation area displays both user queries and system responses
 * - FR-006: User queries appear right-aligned with a background color
 * - FR-006: System responses appear left-aligned with a different background color
 * - FR-023: Special characters in queries and responses display correctly
 */

import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ChatMessage from './ChatMessage';
import type { ChatMessage as ChatMessageType } from '../types/chat';

describe('ChatMessage Component', () => {
  describe('Basic Rendering', () => {
    it('renders user message with correct text', () => {
      const message: ChatMessageType = {
        id: '1',
        text: 'Hello, world!',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      render(<ChatMessage message={message} />);
      
      expect(screen.getByText('Hello, world!')).toBeInTheDocument();
    });

    it('renders system message with correct text', () => {
      const message: ChatMessageType = {
        id: '2',
        text: 'This is a system response',
        sender: 'system',
        timestamp: new Date('2026-01-27T10:00:01Z'),
      };

      render(<ChatMessage message={message} />);
      
      expect(screen.getByText('This is a system response')).toBeInTheDocument();
    });

    it('renders timestamp', () => {
      const message: ChatMessageType = {
        id: '3',
        text: 'Test message',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      render(<ChatMessage message={message} />);
      
      // Check that timestamp is rendered (format may vary)
      const container = screen.getByText('Test message').closest('div');
      expect(container).toHaveTextContent(/10:00|AM|PM/i);
    });
  });

  describe('FR-006: Visual Distinction - User Messages', () => {
    it('applies user message class for styling', () => {
      const message: ChatMessageType = {
        id: '4',
        text: 'User query',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      // Check for user-specific class
      const messageElement = container.querySelector('.chat-message--user');
      expect(messageElement).toBeInTheDocument();
    });

    it('has right-aligned styling for user messages', () => {
      const message: ChatMessageType = {
        id: '5',
        text: 'User query',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      const messageElement = container.querySelector('.chat-message--user');
      // The actual alignment styling will be in CSS, but we verify the class exists
      expect(messageElement).toBeInTheDocument();
    });
  });

  describe('FR-006: Visual Distinction - System Messages', () => {
    it('applies system message class for styling', () => {
      const message: ChatMessageType = {
        id: '6',
        text: 'System response',
        sender: 'system',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      // Check for system-specific class
      const messageElement = container.querySelector('.chat-message--system');
      expect(messageElement).toBeInTheDocument();
    });

    it('has left-aligned styling for system messages', () => {
      const message: ChatMessageType = {
        id: '7',
        text: 'System response',
        sender: 'system',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      const messageElement = container.querySelector('.chat-message--system');
      // The actual alignment styling will be in CSS, but we verify the class exists
      expect(messageElement).toBeInTheDocument();
    });

    it('applies different background styling than user messages', () => {
      const userMessage: ChatMessageType = {
        id: '8',
        text: 'User query',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const systemMessage: ChatMessageType = {
        id: '9',
        text: 'System response',
        sender: 'system',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container: userContainer } = render(<ChatMessage message={userMessage} />);
      const { container: systemContainer } = render(<ChatMessage message={systemMessage} />);
      
      const userElement = userContainer.querySelector('.chat-message--user');
      const systemElement = systemContainer.querySelector('.chat-message--system');
      
      // User and system should have different classes
      expect(userElement).toBeInTheDocument();
      expect(systemElement).toBeInTheDocument();
      expect(userElement?.className).not.toBe(systemElement?.className);
    });
  });

  describe('FR-023: Special Characters Display', () => {
    it('renders special characters in user messages', () => {
      const message: ChatMessageType = {
        id: '10',
        text: 'Query with <special> & "characters" \'test\'',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      render(<ChatMessage message={message} />);
      
      expect(screen.getByText(/Query with <special> & "characters" 'test'/)).toBeInTheDocument();
    });

    it('renders special characters in system messages', () => {
      const message: ChatMessageType = {
        id: '11',
        text: 'Response with <html> tags & symbols: @#$%^&*()',
        sender: 'system',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      render(<ChatMessage message={message} />);
      
      expect(screen.getByText(/Response with <html> tags & symbols: @#\$%\^&\*\(\)/)).toBeInTheDocument();
    });

    it('renders Unicode characters correctly', () => {
      const message: ChatMessageType = {
        id: '12',
        text: 'Unicode test: 你好 🎉 émojis ñ',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      render(<ChatMessage message={message} />);
      
      expect(screen.getByText(/Unicode test: 你好 🎉 émojis ñ/)).toBeInTheDocument();
    });

    it('renders multiline text correctly', () => {
      const message: ChatMessageType = {
        id: '13',
        text: 'Line 1\nLine 2\nLine 3',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      render(<ChatMessage message={message} />);
      
      // Verify all lines are present
      expect(screen.getByText(/Line 1/)).toBeInTheDocument();
      expect(screen.getByText(/Line 2/)).toBeInTheDocument();
      expect(screen.getByText(/Line 3/)).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('renders empty message text', () => {
      const message: ChatMessageType = {
        id: '14',
        text: '',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      // Component should render even with empty text
      expect(container.querySelector('.chat-message')).toBeInTheDocument();
    });

    it('renders very long message text', () => {
      const longText = 'A'.repeat(1000);
      const message: ChatMessageType = {
        id: '15',
        text: longText,
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      render(<ChatMessage message={message} />);
      
      expect(screen.getByText(longText)).toBeInTheDocument();
    });

    it('handles invalid timestamp gracefully', () => {
      const message: ChatMessageType = {
        id: '16',
        text: 'Test message',
        sender: 'user',
        timestamp: new Date('invalid'),
      };

      // Should not throw error
      expect(() => render(<ChatMessage message={message} />)).not.toThrow();
    });
  });

  describe('Accessibility', () => {
    it('has appropriate ARIA attributes for user messages', () => {
      const message: ChatMessageType = {
        id: '17',
        text: 'User query',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      const messageElement = container.querySelector('.chat-message');
      expect(messageElement).toHaveAttribute('role', 'article');
    });

    it('has appropriate ARIA attributes for system messages', () => {
      const message: ChatMessageType = {
        id: '18',
        text: 'System response',
        sender: 'system',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      const messageElement = container.querySelector('.chat-message');
      expect(messageElement).toHaveAttribute('role', 'article');
    });

    it('includes sender information for screen readers', () => {
      const message: ChatMessageType = {
        id: '19',
        text: 'Test message',
        sender: 'user',
        timestamp: new Date('2026-01-27T10:00:00Z'),
      };

      const { container } = render(<ChatMessage message={message} />);
      
      const messageElement = container.querySelector('.chat-message');
      // Should have aria-label or similar indicating sender
      expect(messageElement).toHaveAttribute('aria-label');
    });
  });
});
