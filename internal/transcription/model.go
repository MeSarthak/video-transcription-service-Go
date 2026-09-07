package transcription

import (
	"time"

	"github.com/google/uuid"
)

type Transcript struct {
	ID        uuid.UUID `json:"id"`
	VideoID   uuid.UUID `json:"video_id"`
	JobID     uuid.UUID `json:"job_id"`
	Language  string    `json:"language"`
	FullText  string    `json:"full_text"`
	RawS3Key  string    `json:"raw_s3_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TranscriptSegment struct {
	ID             uuid.UUID `json:"id"`
	TranscriptID   uuid.UUID `json:"transcript_id"`
	SequenceNumber int       `json:"sequence_number"`
	StartTime      float64   `json:"start_time"`
	EndTime        float64   `json:"end_time"`
	Text           string    `json:"text"`
	Confidence     float64   `json:"confidence"`
}

type SegmentResponse struct {
	SequenceNumber int     `json:"sequence_number"`
	StartTime      float64 `json:"start_time"`
	EndTime        float64 `json:"end_time"`
	Text           string  `json:"text"`
	Confidence     float64 `json:"confidence"`
}

type TranscriptResponse struct {
	ID        uuid.UUID          `json:"id"`
	VideoID   uuid.UUID          `json:"video_id"`
	Language  string             `json:"language"`
	FullText  string             `json:"full_text"`
	Segments  []*SegmentResponse `json:"segments"`
	CreatedAt time.Time          `json:"created_at"`
}

func ToResponse(t *Transcript, segments []*TranscriptSegment) *TranscriptResponse {
	var segResponses []*SegmentResponse
	for _, s := range segments {
		segResponses = append(segResponses, &SegmentResponse{
			SequenceNumber: s.SequenceNumber,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			Text:           s.Text,
			Confidence:     s.Confidence,
		})
	}

	return &TranscriptResponse{
		ID:        t.ID,
		VideoID:   t.VideoID,
		Language:  t.Language,
		FullText:  t.FullText,
		Segments:  segResponses,
		CreatedAt: t.CreatedAt,
	}
}
