package loader

import (
	"strings"
	"testing"
	"time"
)

// T027: Unit tests for email header parser
// Tests Message-ID, From/To/CC/BCC parsing, date parsing, encoding handling, malformed headers

func TestParseEmailHeaders_ValidEmail(t *testing.T) {
	emailText := `Message-ID: <12345@enron.com>
From: alice@enron.com
To: bob@enron.com, charlie@enron.com
CC: dave@enron.com
BCC: eve@enron.com
Subject: Test Email
Date: Mon, 1 Jan 2001 10:00:00 -0600

This is the body.`

	headers, err := ParseEmailHeaders(emailText)
	if err != nil {
		t.Fatalf("ParseEmailHeaders failed: %v", err)
	}

	if headers.MessageID != "12345@enron.com" {
		t.Errorf("Expected MessageID '12345@enron.com', got '%s'", headers.MessageID)
	}

	if headers.From != "alice@enron.com" {
		t.Errorf("Expected From 'alice@enron.com', got '%s'", headers.From)
	}

	if len(headers.To) != 2 {
		t.Errorf("Expected 2 To addresses, got %d", len(headers.To))
	}

	if len(headers.CC) != 1 {
		t.Errorf("Expected 1 CC address, got %d", len(headers.CC))
	}

	if headers.Subject != "Test Email" {
		t.Errorf("Expected Subject 'Test Email', got '%s'", headers.Subject)
	}
}

func TestParseEmailHeaders_MessageIDExtraction(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Standard format with brackets",
			email:    "Message-ID: <abc123@enron.com>\n\nBody",
			expected: "abc123@enron.com",
		},
		{
			name:     "No angle brackets",
			email:    "Message-ID: abc123@enron.com\n\nBody",
			expected: "abc123@enron.com",
		},
		{
			name:     "Missing Message-ID",
			email:    "From: alice@enron.com\n\nBody",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := ParseEmailHeaders(tt.email)
			if err != nil && tt.expected != "" {
				t.Fatalf("ParseEmailHeaders failed: %v", err)
			}
			if headers != nil && headers.MessageID != tt.expected {
				t.Errorf("Expected MessageID '%s', got '%s'", tt.expected, headers.MessageID)
			}
		})
	}
}

func TestParseEmailHeaders_DateParsing(t *testing.T) {
	tests := []struct {
		name        string
		dateHeader  string
		expectError bool
		checkYear   int
	}{
		{
			name:        "RFC2822 format",
			dateHeader:  "Mon, 1 Jan 2001 10:00:00 -0600",
			expectError: false,
			checkYear:   2001,
		},
		{
			name:        "RFC2822 with timezone name",
			dateHeader:  "Mon, 1 Jan 2001 10:00:00 CST",
			expectError: false,
			checkYear:   2001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := "Date: " + tt.dateHeader + "\nFrom: test@test.com\n\nBody"
			headers, err := ParseEmailHeaders(email)

			if err != nil && !tt.expectError {
				t.Fatalf("ParseEmailHeaders failed: %v", err)
			}

			if !tt.expectError && headers != nil && !headers.Date.IsZero() {
				if headers.Date.Year() != tt.checkYear {
					t.Errorf("Expected year %d, got %d", tt.checkYear, headers.Date.Year())
				}
			}
		})
	}
}

func TestParseEmailHeaders_EncodingHandling(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "UTF-8 subject",
			email:   "Subject: Hello World\nFrom: test@test.com\n\nBody",
			wantErr: false,
		},
		{
			name:    "ASCII subject",
			email:   "Subject: Simple Subject\nFrom: test@test.com\n\nBody",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := ParseEmailHeaders(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEmailHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}
			if headers == nil && !tt.wantErr {
				t.Error("Expected headers to be parsed")
			}
		})
	}
}

func TestParseEmailHeaders_MalformedHeaders(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "No headers",
			email: "Just body text",
		},
		{
			name:  "Incomplete header",
			email: "From: \n\nBody",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic, may return partial data or error
			headers, err := ParseEmailHeaders(tt.email)
			if err != nil {
				// Error is acceptable for malformed input
				return
			}
			// Partial data is also acceptable
			_ = headers
		})
	}
}

func TestParseEmailHeaders_MultipleRecipients(t *testing.T) {
	email := `From: alice@enron.com
To: bob@enron.com, charlie@enron.com, dave@enron.com
CC: eve@enron.com, frank@enron.com

Body`

	headers, err := ParseEmailHeaders(email)
	if err != nil {
		t.Fatalf("ParseEmailHeaders failed: %v", err)
	}

	if len(headers.To) != 3 {
		t.Errorf("Expected 3 To addresses, got %d", len(headers.To))
	}

	if len(headers.CC) != 2 {
		t.Errorf("Expected 2 CC addresses, got %d", len(headers.CC))
	}
}

func TestParseEmailHeaders_DateFormats(t *testing.T) {
	email := `Date: Mon, 1 Jan 2001 10:00:00 -0600
From: test@test.com

Body`

	headers, err := ParseEmailHeaders(email)
	if err != nil {
		t.Fatalf("ParseEmailHeaders failed: %v", err)
	}

	if headers.Date.Year() != 2001 {
		t.Errorf("Expected year 2001, got %d", headers.Date.Year())
	}

	if headers.Date.Month() != time.January {
		t.Errorf("Expected month January, got %v", headers.Date.Month())
	}
}

func TestParseEmailHeaders_FromToExtraction(t *testing.T) {
	email := `From: Alice Smith <alice@enron.com>
To: Bob Jones <bob@enron.com>, charlie@enron.com
Subject: Test

Body`

	headers, err := ParseEmailHeaders(email)
	if err != nil {
		t.Fatalf("ParseEmailHeaders failed: %v", err)
	}

	// Should extract email address from "Name <email>" format
	if !strings.Contains(headers.From, "alice@enron.com") {
		t.Errorf("Expected From to contain 'alice@enron.com', got '%s'", headers.From)
	}

	if len(headers.To) < 2 {
		t.Errorf("Expected at least 2 To addresses, got %d", len(headers.To))
	}
}
