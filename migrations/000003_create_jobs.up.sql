CREATE TABLE IF NOT EXISTS transcription_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'queued', -- queued, processing, completed, failed
    provider VARCHAR(50) NOT NULL DEFAULT 'aws_transcribe',
    language VARCHAR(20) DEFAULT 'en-US',
    error_message TEXT,
    attempts INT NOT NULL DEFAULT 0,
    last_heartbeat_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_video_id ON transcription_jobs(video_id);
CREATE INDEX IF NOT EXISTS idx_jobs_user_id ON transcription_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON transcription_jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_heartbeat ON transcription_jobs(status, last_heartbeat_at);
