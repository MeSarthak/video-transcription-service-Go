package transcription

import (
	"strings"
	"testing"
)

func TestFormatters(t *testing.T) {
	segments := []*TranscriptSegment{
		{
			SequenceNumber: 1,
			StartTime:      1.25,
			EndTime:        4.50,
			Text:           "Hello and welcome to the course.",
			Confidence:     0.99,
		},
		{
			SequenceNumber: 2,
			StartTime:      5.00,
			EndTime:        8.75,
			Text:           "Today we will build a transcription service.",
			Confidence:     0.98,
		},
	}

	t.Run("FormatSRT produces valid SRT format", func(t *testing.T) {
		srt := FormatSRT(segments)

		expectedFirstHeader := "1\n00:00:01,250 --> 00:00:04,500\nHello and welcome to the course.\n\n"
		expectedSecondHeader := "2\n00:00:05,000 --> 00:00:08,750\nToday we will build a transcription service.\n\n"

		if !strings.Contains(srt, expectedFirstHeader) {
			t.Errorf("expected first SRT chunk: %q, got: %q", expectedFirstHeader, srt)
		}
		if !strings.Contains(srt, expectedSecondHeader) {
			t.Errorf("expected second SRT chunk: %q, got: %q", expectedSecondHeader, srt)
		}
	})

	t.Run("FormatVTT produces valid WebVTT format", func(t *testing.T) {
		vtt := FormatVTT(segments)

		if !strings.HasPrefix(vtt, "WEBVTT\n\n") {
			t.Errorf("expected WEBVTT header, got: %s", vtt)
		}
		if !strings.Contains(vtt, "00:00:01.250 --> 00:00:04.500") {
			t.Errorf("expected period decimal in VTT timestamps, got: %s", vtt)
		}
	})

	t.Run("FormatTXT produces clean plain text", func(t *testing.T) {
		txt := FormatTXT("Hello and welcome to the course. Today we will build a transcription service.")
		if txt != "Hello and welcome to the course. Today we will build a transcription service.\n" {
			t.Errorf("unexpected TXT formatting: %q", txt)
		}
	})
}
