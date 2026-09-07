package transcribe

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type awsTranscribeOutput struct {
	JobName string `json:"jobName"`
	Status  string `json:"status"`
	Results struct {
		Transcripts []struct {
			Transcript string `json:"transcript"`
		} `json:"transcripts"`
		Items []awsItem `json:"items"`
	} `json:"results"`
}

type awsItem struct {
	StartTime    string `json:"start_time,omitempty"`
	EndTime      string `json:"end_time,omitempty"`
	Type         string `json:"type"` // "pronunciation" or "punctuation"
	Alternatives []struct {
		Confidence string `json:"confidence"`
		Content    string `json:"content"`
	} `json:"alternatives"`
}

// ParseAWSOutput parses the raw JSON output from AWS Transcribe into full text and grouped subtitle segments.
func ParseAWSOutput(rawJSON []byte) (string, []SegmentResult, error) {
	var output awsTranscribeOutput
	if err := json.Unmarshal(rawJSON, &output); err != nil {
		return "", nil, fmt.Errorf("failed to parse AWS Transcribe JSON: %w", err)
	}

	var fullText string
	if len(output.Results.Transcripts) > 0 {
		fullText = output.Results.Transcripts[0].Transcript
	}

	// Group individual word items into coherent timestamped subtitle segments
	segments := buildSegments(output.Results.Items)

	// Fallback if no items but fullText exists
	if len(segments) == 0 && fullText != "" {
		segments = append(segments, SegmentResult{
			SequenceNumber: 1,
			StartTime:      0.0,
			EndTime:        5.0,
			Text:           fullText,
			Confidence:     1.0,
		})
	}

	return fullText, segments, nil
}

func buildSegments(items []awsItem) []SegmentResult {
	var segments []SegmentResult
	var currentWords []string
	var segStart float64 = -1
	var segEnd float64 = 0
	var totalConfidence float64 = 0
	var wordCount int = 0
	seqNum := 1

	flushSegment := func() {
		if len(currentWords) > 0 {
			segText := strings.TrimSpace(strings.Join(currentWords, " "))
			// Clean up spacing before punctuation
			segText = strings.ReplaceAll(segText, " ,", ",")
			segText = strings.ReplaceAll(segText, " .", ".")
			segText = strings.ReplaceAll(segText, " ?", "?")
			segText = strings.ReplaceAll(segText, " !", "!")

			avgConf := 1.0
			if wordCount > 0 {
				avgConf = totalConfidence / float64(wordCount)
			}

			segments = append(segments, SegmentResult{
				SequenceNumber: seqNum,
				StartTime:      segStart,
				EndTime:        segEnd,
				Text:           segText,
				Confidence:     avgConf,
			})

			seqNum++
			currentWords = nil
			segStart = -1
			segEnd = 0
			totalConfidence = 0
			wordCount = 0
		}
	}

	for _, item := range items {
		if len(item.Alternatives) == 0 {
			continue
		}

		content := item.Alternatives[0].Content
		confStr := item.Alternatives[0].Confidence
		conf, _ := strconv.ParseFloat(confStr, 64)

		if item.Type == "punctuation" {
			if len(currentWords) > 0 {
				currentWords[len(currentWords)-1] += content
			}
			// If sentence ending punctuation (. ? !), flush segment
			if content == "." || content == "?" || content == "!" {
				flushSegment()
			}
			continue
		}

		// Pronunciation word item
		start, _ := strconv.ParseFloat(item.StartTime, 64)
		end, _ := strconv.ParseFloat(item.EndTime, 64)

		if segStart < 0 {
			segStart = start
		}
		segEnd = end
		totalConfidence += conf
		wordCount++
		currentWords = append(currentWords, content)

		// Also break segment if it reaches ~8 words for readable subtitle length
		if wordCount >= 8 {
			flushSegment()
		}
	}

	flushSegment()
	return segments
}
