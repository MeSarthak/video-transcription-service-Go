export interface User {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface Video {
  id: string;
  user_id: string;
  title: string;
  original_filename: string;
  s3_key: string;
  s3_bucket: string;
  duration_seconds: number;
  file_size_bytes: number;
  mime_type: string;
  status: 'pending' | 'uploaded' | 'processing' | 'ready' | 'failed';
  playback_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Job {
  id: string;
  video_id: string;
  user_id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  retry_count: number;
  max_retries: number;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface TranscriptSegment {
  id: string;
  transcript_id: string;
  segment_index: number;
  start_time: number;
  end_time: number;
  text: string;
  confidence: number;
  created_at: string;
}

export interface Transcript {
  id: string;
  video_id: string;
  job_id: string;
  language_code: string;
  full_text: string;
  duration_seconds: number;
  provider: string;
  segments?: TranscriptSegment[];
  created_at: string;
  updated_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface UploadUrlResponse {
  upload_url: string;
  s3_key: string;
  expires_in: number;
  video_id: string;
}

export interface VideoDetailResponse {
  video: Video;
  job?: Job;
  transcript?: Transcript;
  segments?: TranscriptSegment[];
}
