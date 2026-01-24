package loader

import (
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// EmailMetadata represents parsed email headers and content
type EmailMetadata struct {
	MessageID string
	Date      time.Time
	From      string
	To        []string
	CC        []string
	BCC       []string
	Subject   string
	Body      string
}

// ParseEmailHeaders extracts metadata from a raw email message
func ParseEmailHeaders(message string) (*EmailMetadata, error) {
	// Try to detect and convert encoding
	message = normalizeEncoding(message)

	// Parse email using net/mail
	reader := strings.NewReader(message)
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	metadata := &EmailMetadata{}

	// Extract Message-ID
	metadata.MessageID = msg.Header.Get("Message-ID")
	if metadata.MessageID == "" {
		metadata.MessageID = msg.Header.Get("Message-Id")
	}
	// Clean Message-ID (remove < >)
	metadata.MessageID = strings.Trim(metadata.MessageID, "<> ")

	// Extract Date
	dateStr := msg.Header.Get("Date")
	if dateStr != "" {
		// Parse RFC 2822 date format
		date, err := mail.ParseDate(dateStr)
		if err != nil {
			// Try alternative formats if RFC 2822 fails
			date, err = parseAlternativeDate(dateStr)
			if err != nil {
				// Use current time as fallback
				metadata.Date = time.Now()
			} else {
				metadata.Date = date
			}
		} else {
			metadata.Date = date
		}
	} else {
		metadata.Date = time.Now()
	}

	// Extract From
	metadata.From = extractEmail(msg.Header.Get("From"))

	// Extract To
	metadata.To = extractEmailList(msg.Header.Get("To"))

	// Extract CC
	metadata.CC = extractEmailList(msg.Header.Get("Cc"))

	// Extract BCC
	metadata.BCC = extractEmailList(msg.Header.Get("Bcc"))

	// Extract Subject
	metadata.Subject = msg.Header.Get("Subject")

	// Extract Body
	bodyBytes, err := io.ReadAll(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read email body: %w", err)
	}
	metadata.Body = string(bodyBytes)

	return metadata, nil
}

// normalizeEncoding attempts to convert text to UTF-8
func normalizeEncoding(text string) string {
	// Try to detect if it's Latin-1/ISO-8859-1 and convert to UTF-8
	// This is a simple heuristic - check for high bytes that aren't valid UTF-8
	if !strings.Contains(text, "\ufffd") && strings.ContainsAny(text, "\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f") {
		// Likely Latin-1, try to convert
		decoder := charmap.ISO8859_1.NewDecoder()
		result, _, err := transform.String(decoder, text)
		if err == nil {
			return result
		}
	}
	return text
}

// extractEmail extracts a single email address from a header field
func extractEmail(field string) string {
	if field == "" {
		return ""
	}

	// Parse address
	addr, err := mail.ParseAddress(field)
	if err != nil {
		// If parsing fails, try to extract email using simple regex-like approach
		parts := strings.Split(field, "<")
		if len(parts) > 1 {
			email := strings.Split(parts[1], ">")[0]
			return strings.TrimSpace(email)
		}
		// Return as-is if no brackets
		return strings.TrimSpace(field)
	}
	return addr.Address
}

// extractEmailList extracts multiple email addresses from a header field
func extractEmailList(field string) []string {
	if field == "" {
		return nil
	}

	// Parse address list
	addresses, err := mail.ParseAddressList(field)
	if err != nil {
		// Fall back to simple splitting by comma
		parts := strings.Split(field, ",")
		var emails []string
		for _, part := range parts {
			email := extractEmail(strings.TrimSpace(part))
			if email != "" {
				emails = append(emails, email)
			}
		}
		return emails
	}

	var emails []string
	for _, addr := range addresses {
		if addr.Address != "" {
			emails = append(emails, addr.Address)
		}
	}
	return emails
}

// parseAlternativeDate tries alternative date formats
func parseAlternativeDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
