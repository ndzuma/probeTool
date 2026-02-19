package findings

import (
	"strings"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	content := `# Security Audit Report

## Critical
- SQL injection in login form allows unauthorized access
- Hardcoded API key exposed in source code

## High
- Missing CSRF protection on payment endpoint
- Weak password policy allows 4-character passwords

## Medium
- Verbose error messages leak stack traces
- Session tokens don't expire after logout

## Low
- Missing security headers on static assets
- Outdated dependencies with known CVEs
`

	findings := ParseMarkdown(content)

	if len(findings) == 0 {
		t.Error("ParseMarkdown should return findings")
	}

	// Check that we have findings from different severities
	var hasCritical, hasHigh, hasMedium, hasLow bool
	for _, f := range findings {
		switch f.Severity {
		case "critical":
			hasCritical = true
		case "high":
			hasHigh = true
		case "medium":
			hasMedium = true
		case "low":
			hasLow = true
		}
	}

	if !hasCritical {
		t.Error("Should have critical findings")
	}
	if !hasHigh {
		t.Error("Should have high findings")
	}
	if !hasMedium {
		t.Error("Should have medium findings")
	}
	if !hasLow {
		t.Error("Should have low findings")
	}
}

func TestParseMarkdownEmpty(t *testing.T) {
	findings := ParseMarkdown("")
	if len(findings) != 0 {
		t.Errorf("Empty content should return 0 findings, got %d", len(findings))
	}
}

func TestParseMarkdownNoFindings(t *testing.T) {
	content := `# Regular Document

This is just a regular document with no security findings.
It has some bullet points:
- Point one
- Point two
But these are not security issues.
`

	findings := ParseMarkdown(content)
	// May still find bullet points but without severity
	// This is okay behavior
	t.Logf("Found %d items in non-security document", len(findings))
}

func TestParseSection(t *testing.T) {
	seenTexts := make(map[string]bool)
	// Note: The pattern matches severity at end of line, so test with single-line heading first
	// In real usage, the section would be split and processed line by line
	section := "## Critical"

	findings := parseSection(section, seenTexts)

	// parseSection returns empty for headings without bullet points
	// This is expected behavior
	if len(findings) != 0 {
		t.Errorf("parseSection should return 0 findings for heading-only section, got %d", len(findings))
	}

	// Test with bullets - need to use the actual parser flow
	content := "# Security Report\n\n## Critical\n- SQL injection vulnerability in user input\n- Remote code execution via file upload"
	allFindings := ParseMarkdown(content)

	if len(allFindings) == 0 {
		t.Error("ParseMarkdown should find findings")
	}

	for _, f := range allFindings {
		if f.ID == "" {
			t.Error("Finding should have an ID")
		}
		if f.Text == "" {
			t.Error("Finding should have text")
		}
		if f.Severity == "" {
			t.Error("Finding should have severity")
		}
	}
}

func TestParseBulletPoints(t *testing.T) {
	seenTexts := make(map[string]bool)
	content := `## Findings
- High: Missing rate limiting on API endpoints
- Medium: Verbose error messages in production
- Low: Missing Content-Security-Policy header`

	findings := parseBulletPoints(content, seenTexts)

	if len(findings) == 0 {
		t.Error("parseBulletPoints should return findings")
	}

	// Check for inline severity markers
	var hasHigh bool
	for _, f := range findings {
		if f.Severity == "high" {
			hasHigh = true
		}
	}
	if !hasHigh {
		t.Error("Should detect inline high severity markers")
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"CRITICAL", "critical"},
		{"Critical", "critical"},
		{"severe", "critical"},
		{"high", "high"},
		{"HIGH", "high"},
		{"medium", "medium"},
		{"moderate", "medium"},
		{"low", "low"},
		{"minor", "low"},
		{"info", "info"},
		{"note", "info"},
		{"suggestion", "info"},
		{"unknown", "info"},
		{"", "info"},
	}

	for _, tt := range tests {
		result := normalizeSeverity(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeSeverity(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFindingStruct(t *testing.T) {
	finding := Finding{
		ID:       "abc123",
		Text:     "Test finding",
		Severity: "high",
	}

	if finding.ID != "abc123" {
		t.Errorf("Finding.ID = %v, want abc123", finding.ID)
	}
	if finding.Text != "Test finding" {
		t.Errorf("Finding.Text = %v, want Test finding", finding.Text)
	}
	if finding.Severity != "high" {
		t.Errorf("Finding.Severity = %v, want high", finding.Severity)
	}
}

func TestParseMarkdownDuplicatePrevention(t *testing.T) {
	content := `# Security Report

## Critical
- SQL injection vulnerability
- SQL injection vulnerability

## High
- SQL injection vulnerability`

	findings := ParseMarkdown(content)

	// Should not have duplicates
	seenTexts := make(map[string]bool)
	for _, f := range findings {
		text := strings.ToLower(f.Text)
		if seenTexts[text] {
			t.Errorf("Duplicate finding found: %s", f.Text)
		}
		seenTexts[text] = true
	}
}

func TestParseMarkdownWithSeverityMarkers(t *testing.T) {
	content := `# Security Report

## Findings
- [CRITICAL] Database credentials exposed in environment variables
- [HIGH] Missing input validation on user registration
- [MEDIUM] Weak session timeout configuration
- [LOW] Missing security headers`

	findings := ParseMarkdown(content)

	if len(findings) == 0 {
		t.Error("Should parse findings with inline severity markers")
	}

	// Check that severities were extracted
	severityCount := make(map[string]int)
	for _, f := range findings {
		severityCount[f.Severity]++
	}

	if severityCount["critical"] == 0 {
		t.Error("Should extract critical severity")
	}
	if severityCount["high"] == 0 {
		t.Error("Should extract high severity")
	}
}

func TestBulletPattern(t *testing.T) {
	// Test various bullet formats
	lines := []string{
		"- This is a bullet",
		"* This is also a bullet",
		"  - Indented bullet",
		"  * Indented asterisk",
	}

	for _, line := range lines {
		matches := bulletPattern.FindStringSubmatch(line)
		if matches == nil {
			t.Errorf("Should match bullet: %s", line)
		}
	}
}

func TestHeadingSeverityPattern(t *testing.T) {
	// Test various heading formats - pattern expects severity at END of heading
	headings := []string{
		"## Critical",
		"## HIGH",
		"## Criticals",
		"### Low",
		"## Info",
	}

	for _, heading := range headings {
		matches := headingSeverityPattern.FindStringSubmatch(heading)
		if matches == nil {
			t.Errorf("Should match severity heading: %s", heading)
		}
	}
}
