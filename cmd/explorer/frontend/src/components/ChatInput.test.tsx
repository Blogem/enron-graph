/**
 * ChatInput Component Tests (T011)
 * Test Suite for User Story 1 - Display Chat Interface
 * 
 * Requirements tested:
 * - FR-003: Pressing Enter key in the input field should submit the query
 * - FR-004: Pressing Shift+Enter in the input field should create a new line
 * - FR-015: Empty queries (whitespace only) should not be submitted
 * - FR-021: Query submission should be disabled while a query is being processed
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ChatInput from './ChatInput';

describe('ChatInput Component', () => {
  describe('Basic Rendering', () => {
    it('renders input field with default placeholder', () => {
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      expect(input).toBeInTheDocument();
      expect(input).toHaveAttribute('placeholder');
    });

    it('renders input field with custom placeholder', () => {
      const mockSubmit = vi.fn();
      const customPlaceholder = 'Ask a question...';
      render(<ChatInput onSubmit={mockSubmit} placeholder={customPlaceholder} />);
      
      const input = screen.getByPlaceholderText(customPlaceholder);
      expect(input).toBeInTheDocument();
    });

    it('renders submit button', () => {
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const button = screen.getByRole('button', { name: /send/i });
      expect(button).toBeInTheDocument();
    });
  });

  describe('FR-003: Enter Key Submission', () => {
    it('submits query when Enter key is pressed', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      await user.type(input, 'test query{Enter}');
      
      expect(mockSubmit).toHaveBeenCalledWith('test query');
      expect(mockSubmit).toHaveBeenCalledTimes(1);
    });

    it('clears input field after submission via Enter', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox') as HTMLInputElement;
      await user.type(input, 'test query{Enter}');
      
      expect(input.value).toBe('');
    });
  });

  describe('FR-004: Shift+Enter Newline', () => {
    it('creates newline when Shift+Enter is pressed', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox') as HTMLInputElement;
      await user.type(input, 'line 1{Shift>}{Enter}{/Shift}line 2');
      
      expect(mockSubmit).not.toHaveBeenCalled();
      expect(input.value).toContain('\n');
    });

    it('does not submit when Shift+Enter is pressed', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      await user.type(input, 'line 1{Shift>}{Enter}{/Shift}line 2');
      
      expect(mockSubmit).not.toHaveBeenCalled();
    });
  });

  describe('FR-015: Empty Query Prevention', () => {
    it('does not submit empty query', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      await user.type(input, '{Enter}');
      
      expect(mockSubmit).not.toHaveBeenCalled();
    });

    it('does not submit whitespace-only query', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      await user.type(input, '   {Enter}');
      
      expect(mockSubmit).not.toHaveBeenCalled();
    });

    it('trims whitespace from submitted query', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      await user.type(input, '  test query  {Enter}');
      
      expect(mockSubmit).toHaveBeenCalledWith('test query');
    });
  });

  describe('Button Submission', () => {
    it('submits query when button is clicked', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      const button = screen.getByRole('button', { name: /send/i });
      
      await user.type(input, 'test query');
      await user.click(button);
      
      expect(mockSubmit).toHaveBeenCalledWith('test query');
    });

    it('clears input field after button submission', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox') as HTMLInputElement;
      const button = screen.getByRole('button', { name: /send/i });
      
      await user.type(input, 'test query');
      await user.click(button);
      
      expect(input.value).toBe('');
    });

    it('does not submit empty query via button', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const button = screen.getByRole('button', { name: /send/i });
      await user.click(button);
      
      expect(mockSubmit).not.toHaveBeenCalled();
    });
  });

  describe('FR-021: Disabled State', () => {
    it('disables input when disabled prop is true', () => {
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} disabled={true} />);
      
      const input = screen.getByRole('textbox');
      expect(input).toBeDisabled();
    });

    it('disables button when disabled prop is true', () => {
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} disabled={true} />);
      
      const button = screen.getByRole('button', { name: /send/i });
      expect(button).toBeDisabled();
    });

    it('does not submit when disabled via Enter key', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} disabled={true} />);
      
      const input = screen.getByRole('textbox');
      await user.type(input, 'test query{Enter}');
      
      expect(mockSubmit).not.toHaveBeenCalled();
    });

    it('does not submit when disabled via button click', async () => {
      const user = userEvent.setup();
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} disabled={true} />);
      
      const button = screen.getByRole('button', { name: /send/i });
      await user.click(button);
      
      expect(mockSubmit).not.toHaveBeenCalled();
    });
  });

  describe('Accessibility', () => {
    it('input has accessible label', () => {
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const input = screen.getByRole('textbox');
      expect(input).toHaveAccessibleName();
    });

    it('button has accessible label', () => {
      const mockSubmit = vi.fn();
      render(<ChatInput onSubmit={mockSubmit} />);
      
      const button = screen.getByRole('button', { name: /send/i });
      expect(button).toHaveAccessibleName();
    });
  });
});
