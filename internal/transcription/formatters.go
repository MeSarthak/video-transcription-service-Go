package transcription

import (
	"fmt"
	"math"
	"strings"
)

// FormatSRT generates SubRip Subtitle (.srt) format from segments.
func FormatSRT(segments []*TranscriptSegment) string {
	var builder strings.Builder

	for i, seg := range segments {
		start := formatSRTTime(seg.StartTime)
		end := formatSRTTime(seg.EndTime)

		builder.WriteString(fmt.Sprintf("%d\n", i+1))
		builder.WriteString(fmt.Sprintf("%s --> %s\n", start, end))
		builder.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(seg.Text)))
	}

	return builder.String()
}

// FormatVTT generates WebVTT (.vtt) format from segments.
func FormatVTT(segments []*TranscriptSegment) string {
	var builder strings.Builder
	builder.WriteString("WEBVTT\n\n")

	for i, seg := range segments {
		start := formatVTTTime(seg.StartTime)
		end := formatVTTTime(seg.EndTime)

		builder.WriteString(fmt.Sprintf("%d\n", i+1))
		builder.WriteString(fmt.Sprintf("%s --> %s\n", start, end))
		builder.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(seg.Text)))
	}

	return builder.String()
}

// FormatTXT returns plain text.
func FormatTXT(fullText string) string {
	return strings.TrimSpace(fullText) + "\n"
}

func formatSRTTime(seconds float64) string {
	return formatTimestamp(seconds, ",")
}

func formatVTTTime(seconds float64) string {
	return formatTimestamp(seconds, ".")
}

func formatTimestamp(seconds float64, sep string) string {
	hours := int(seconds / 3600)
	minutes := int(math.Mod(seconds, 3600) / 60)
	secs := int(math.Mod(seconds, 60))
	millis := int(math.Round(math.Mod(seconds, 1) * 1000))

	if millis >= 1000 {
		millis = 999
	}

	return fmt.Sprintf("%02d:%02d:%02d%s%03d", hours, minutes, secs, sep, millis)
}
