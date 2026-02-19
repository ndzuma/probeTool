package findings

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Finding struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Severity string `json:"severity"`
}

var severityPatterns = map[string]*regexp.Regexp{
	"critical": regexp.MustCompile(`(?i)(?:critical|severe)\s*[:\-]?\s*(.+?)(?:\n|$)`),
	"high":     regexp.MustCompile(`(?i)(?:high)\s*[:\-]?\s*(.+?)(?:\n|$)`),
	"medium":   regexp.MustCompile(`(?i)(?:medium|moderate)\s*[:\-]?\s*(.+?)(?:\n|$)`),
	"low":      regexp.MustCompile(`(?i)(?:low|minor)\s*[:\-]?\s*(.+?)(?:\n|$)`),
	"info":     regexp.MustCompile(`(?i)(?:info|note|suggestion)\s*[:\-]?\s*(.+?)(?:\n|$)`),
}

var bulletPattern = regexp.MustCompile(`(?m)^\s*[-*]\s*(.+)$`)
var headingSeverityPattern = regexp.MustCompile(`(?i)#+\s*(critical|high|medium|low|info|note|suggestion)s?\s*$`)

func ParseMarkdown(content string) []Finding {
	var findings []Finding
	seenTexts := make(map[string]bool)

	sections := strings.Split(content, "\n#")
	for _, section := range sections {
		if !strings.HasPrefix(section, "#") {
			section = "#" + section
		}
		findings = append(findings, parseSection(section, seenTexts)...)
	}

	bulletFindings := parseBulletPoints(content, seenTexts)
	findings = append(findings, bulletFindings...)

	return findings
}

func parseSection(section string, seenTexts map[string]bool) []Finding {
	var findings []Finding

	headingMatch := headingSeverityPattern.FindStringSubmatch(section)
	if len(headingMatch) < 2 {
		return findings
	}

	severity := normalizeSeverity(headingMatch[1])

	bullets := bulletPattern.FindAllStringSubmatch(section, -1)
	for _, match := range bullets {
		if len(match) > 1 {
			text := strings.TrimSpace(match[1])
			if text == "" || len(text) < 10 {
				continue
			}
			if seenTexts[strings.ToLower(text)] {
				continue
			}
			seenTexts[strings.ToLower(text)] = true
			findings = append(findings, Finding{
				ID:       uuid.New().String()[:8],
				Text:     text,
				Severity: severity,
			})
		}
	}

	return findings
}

func parseBulletPoints(content string, seenTexts map[string]bool) []Finding {
	var findings []Finding

	inlineSeverity := regexp.MustCompile(`(?i)^\s*[-*]\s*\[?(critical|high|medium|low|info)\]?\s*[:\-]?\s*(.+)$`)

	lines := strings.Split(content, "\n")
	var currentSeverity string

	for _, line := range lines {
		headingMatch := headingSeverityPattern.FindStringSubmatch(line)
		if len(headingMatch) > 1 {
			currentSeverity = normalizeSeverity(headingMatch[1])
			continue
		}

		inlineMatch := inlineSeverity.FindStringSubmatch(line)
		if len(inlineMatch) >= 3 {
			severity := normalizeSeverity(inlineMatch[1])
			text := strings.TrimSpace(inlineMatch[2])
			if text != "" && len(text) >= 10 && !seenTexts[strings.ToLower(text)] {
				seenTexts[strings.ToLower(text)] = true
				findings = append(findings, Finding{
					ID:       uuid.New().String()[:8],
					Text:     text,
					Severity: severity,
				})
			}
			continue
		}

		bulletMatch := bulletPattern.FindStringSubmatch(line)
		if len(bulletMatch) > 1 && currentSeverity != "" {
			text := strings.TrimSpace(bulletMatch[1])
			if text != "" && len(text) >= 10 && !seenTexts[strings.ToLower(text)] {
				seenTexts[strings.ToLower(text)] = true
				findings = append(findings, Finding{
					ID:       uuid.New().String()[:8],
					Text:     text,
					Severity: currentSeverity,
				})
			}
		}
	}

	return findings
}

func normalizeSeverity(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "critical", "severe":
		return "critical"
	case "high":
		return "high"
	case "medium", "moderate":
		return "medium"
	case "low", "minor":
		return "low"
	default:
		return "info"
	}
}
