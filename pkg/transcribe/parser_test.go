package transcribe

import (
	"testing"
)

func TestParseAWSOutput(t *testing.T) {
	rawSampleJSON := []byte(`{
		"jobName": "test-job",
		"status": "COMPLETED",
		"results": {
			"transcripts": [
				{"transcript": "Hello world. Welcome to the Go transcription service."}
			],
			"items": [
				{
					"start_time": "0.10",
					"end_time": "0.40",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.99", "content": "Hello"}]
				},
				{
					"start_time": "0.41",
					"end_time": "0.80",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.98", "content": "world"}]
				},
				{
					"type": "punctuation",
					"alternatives": [{"content": "."}]
				},
				{
					"start_time": "1.00",
					"end_time": "1.40",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.95", "content": "Welcome"}]
				},
				{
					"start_time": "1.45",
					"end_time": "1.60",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.99", "content": "to"}]
				},
				{
					"start_time": "1.65",
					"end_time": "1.80",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.99", "content": "the"}]
				},
				{
					"start_time": "1.85",
					"end_time": "2.10",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.97", "content": "Go"}]
				},
				{
					"start_time": "2.15",
					"end_time": "2.70",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.96", "content": "transcription"}]
				},
				{
					"start_time": "2.75",
					"end_time": "3.20",
					"type": "pronunciation",
					"alternatives": [{"confidence": "0.98", "content": "service"}]
				},
				{
					"type": "punctuation",
					"alternatives": [{"content": "."}]
				}
			]
		}
	}`)

	fullText, segments, err := ParseAWSOutput(rawSampleJSON)
	if err != nil {
		t.Fatalf("ParseAWSOutput failed: %v", err)
	}

	if fullText != "Hello world. Welcome to the Go transcription service." {
		t.Errorf("unexpected full text: %s", fullText)
	}

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments separated by period, got %d", len(segments))
	}

	// First segment
	if segments[0].Text != "Hello world." {
		t.Errorf("expected 'Hello world.', got '%s'", segments[0].Text)
	}
	if segments[0].StartTime != 0.10 || segments[0].EndTime != 0.80 {
		t.Errorf("expected 0.10 -> 0.80, got %f -> %f", segments[0].StartTime, segments[0].EndTime)
	}

	// Second segment
	if segments[1].Text != "Welcome to the Go transcription service." {
		t.Errorf("expected 'Welcome to the Go transcription service.', got '%s'", segments[1].Text)
	}
	if segments[1].StartTime != 1.00 || segments[1].EndTime != 3.20 {
		t.Errorf("expected 1.00 -> 3.20, got %f -> %f", segments[1].StartTime, segments[1].EndTime)
	}
}
