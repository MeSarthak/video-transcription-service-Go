export interface User {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface Video {
  id: string;
  user_id: string;
  filename: string;
  storage_key: string;
  content_type: string;
  size_bytes: number;
  duration_seconds?: number;
  status: 'uploading' | 'uploaded' | 'processing' | 'completed' | 'failed' | 'deleted';
  playback_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Job {
  id: string;
  video_id: string;
  user_id: string;
  status: 'queued' | 'processing' | 'completed' | 'failed';
  provider: string;
  language: string;
  error_message?: string;
  attempts: number;
  last_heartbeat_at?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface TranscriptSegment {
  sequence_number: number;
  start_time: number;
  end_time: number;
  text: string;
  confidence: number;
}

export interface Transcript {
  id: string;
  video_id: string;
  language: string;
  full_text: string;
  segments: TranscriptSegment[];
  created_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface UploadUrlResponse {
  video_id: string;
  storage_key: string;
  upload_url: string;
  expires_in: number;
}
