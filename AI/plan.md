# Video Transcription Service — AWS + Go + HeroUI Frontend

## 1. Project Overview

Build a production-oriented video transcription platform with a **Go backend**, **AWS cloud infrastructure**, and a modern **HeroUI (React) frontend**.

The service allows users to:

- Create accounts and authenticate securely (JWT).
- Upload video files directly from the browser to Amazon S3 via presigned URLs with upload progress tracking.
- Store videos securely in private S3 buckets.
- Asynchronously extract audio with FFmpeg in a dedicated background worker.
- Transcribe audio with AWS Transcribe (or pluggable providers).
- Store transcripts and timestamped segments in PostgreSQL.
- Monitor job progress with live status polling and resilient error/retry handling.
- View an interactive video player synchronized with clickable transcript timestamps.
- Download/export transcripts as TXT, SRT, and VTT.
- Search within transcripts.
- Scale transcription workers independently from the API.

---

## 2. Main Learning Goals

### Go / Backend
- Gin web framework
- Clean architecture & repository pattern
- REST API design & middleware (Auth, CORS, Logging, Rate limiting)
- Authentication (bcrypt/Argon2 + JWT access & refresh tokens)
- PostgreSQL & `pgxpool` connection pooling
- Database migrations with `golang-migrate`
- Database transactions & state machine transitions
- Direct-to-S3 presigned URLs & object verification (`HeadObject`)
- SQS message consumption, long polling, visibility timeout management, & Dead-Letter Queues (DLQ)
- Worker crash resilience, heartbeats, and idempotency
- FFmpeg subprocess orchestration, audio normalization, and scratch disk cleanup
- Pluggable transcription provider interface (`AWS Transcribe`, `Whisper`, etc.)
- Graceful shutdown & context cancellation
- Structured JSON logging (`slog`)

### Frontend (HeroUI + React)
- Modern UI development with [HeroUI](https://heroui.com/) (`@heroui/react`) and Tailwind CSS
- Client-side direct-to-S3 streaming uploads with accurate progress bars
- Interactive HTML5 video player synchronized with timestamped transcript cues
- Click-to-seek timestamp navigation
- Live job status polling and UI state transitions
- Responsive dark/light theme integration

### AWS & Cloud Architecture
- **IAM**: Least-privilege IAM roles for ECS Tasks (API vs. Worker)
- **S3**: Private buckets, presigned URLs, CORS configuration, lifecycle rules
- **SQS**: Standard queue, visibility timeout extension, redrive policy to DLQ
- **RDS**: Managed PostgreSQL in isolated private subnets
- **ECS Fargate**: Containerized API and Worker services with independent scaling
- **ECR**: Container image repositories and versioning
- **VPC**: Public/private subnet topology, NAT Gateway, Security Groups
- **ALB**: Application Load Balancer with SSL termination and health checks
- **CloudFront**: CDN delivering static HeroUI frontend from S3 and routing `/api/*` to ALB
- **Secrets Manager**: Zero secrets stored in code or environment variables
- **CloudWatch**: Logs, metrics, alarms, and worker queue-depth autoscaling
- **AWS Transcribe**: Asynchronous speech-to-text processing and output normalization

---

## 3. High-Level Architecture

```text
                                  ┌─────────────────────────────┐
                                  │           Browser           │
                                  │   (HeroUI React Frontend)   │
                                  └──────────────┬──────────────┘
                                                 │
                                                 │ HTTPS
                                                 ▼
                                  ┌─────────────────────────────┐
                                  │      Amazon CloudFront      │
                                  │     (CDN & Route Router)    │
                                  └──────┬───────────────┬──────┘
                                         │               │
                           Static Assets │               │ /api/v1/*
                                         ▼               ▼
                        ┌──────────────────┐   ┌────────────────────────┐
                        │ S3 (Frontend App)│   │  App Load Balancer     │
                        └──────────────────┘   └───────────┬────────────┘
                                                           │
                                                           ▼
                                               ┌────────────────────────┐
                                               │   Go API (ECS Fargate) │
                                               └───────┬─────────┬──────┘
                                                       │         │
                                               ┌───────┘         └──────────────┐
                                               ▼                                ▼
                                        ┌──────────────┐                ┌──────────────┐
                                        │ RDS Postgres │                │ S3 Bucket    │
                                        │ (Private DB) │                │ Videos/Audio │
                                        └──────────────┘                │ Transcripts  │
                                                                        └──────┬───────┘
                                                                               │
                                                                               ▼
                                                                        ┌──────────────┐
                                                                        │ SQS Queue    │
                                                                        │ + SQS DLQ    │
                                                                        └──────┬───────┘
                                                                               │
                                                                               ▼
                                                                 ┌───────────────────────────┐
                                                                 │ Go Worker (ECS Fargate)   │
                                                                 │ Heartbeat + Auto-cleanup  │
                                                                 └─────────────┬─────────────┘
                                                                               │
                                                                 ┌─────────────┴─────────────┐
                                                                 ▼                           ▼
                                                           ┌───────────┐               ┌─────────────┐
                                                           │  FFmpeg   │               │ AWS         │
                                                           │ (Extract) │               │ Transcribe  │
                                                           └───────────┘               └─────────────┘
```

---

## 4. Key Architectural Principles & Fixes

### 4.1 Direct Browser-to-S3 Presigned Uploads
**Principle:** Do NOT stream video uploads through the Go API server.

```text
Browser (HeroUI)              Go API                       S3 Bucket
      │                         │                              │
      │ 1. POST /videos/upload  │                              │
      ├────────────────────────>│                              │
      │    (title, size, type)  │ 2. Create DB record ('uploading')
      │                         │    Generate S3 Presigned PUT URL
      │ 3. Return Presigned URL │                              │
      │<────────────────────────┤                              │
      │                                                        │
      │ 4. Direct PUT video stream with progress bar           │
      ├───────────────────────────────────────────────────────>│
      │                                                        │
      │ 5. POST /videos/:id/complete-upload                    │
      ├────────────────────────>│                              │
      │                         │ 6. s3.HeadObject verification│
      │                         │    (checks existence & size) │
      │                         ├─────────────────────────────>│
      │                         │                              │
      │                         │ 7. Update status: 'uploaded' │
      │ 8. Return Success       │                              │
      │<────────────────────────┤                              │
```

* **S3 CORS**: S3 bucket must have CORS configured allowing `PUT`, `GET`, `HEAD` from the frontend origin.
* **Upload Verification**: `complete-upload` calls `s3.HeadObject` to ensure the object actually exists, matches expected size/type, and prevents 0-byte or fake uploads.

---

### 4.2 SQS Visibility Timeout vs. AWS Transcribe Latency
**The Challenge:** AWS Transcribe is asynchronous (`StartTranscriptionJob`). Long audio files take minutes to transcribe. If the worker blocks polling Transcribe while SQS visibility timeout is short (e.g., 30s), SQS assumes the worker crashed and delivers the duplicate job to another worker.

**The Solution (Worker Heartbeat Extension):**
* Worker runs a background goroutine during processing that periodically calls `sqs.ChangeMessageVisibility` (e.g. extending visibility by 60s every 20s).
* When Transcribe finishes (or fails), the goroutine stops, worker writes results to DB, and deletes the SQS message (`sqs.DeleteMessage`).
* If the worker genuinely crashes (OOM/kill), the heartbeat stops -> visibility timeout expires -> SQS redelivers -> after `maxReceiveCount` (e.g., 3 attempts), message lands safely in the DLQ.

---

### 4.3 Worker Crash & Zombie Job Handling
**The Challenge:** If a worker container dies midway, the database job record might stay stuck in `processing`.

**The Solution:**
* Add `last_heartbeat_at TIMESTAMP` and `attempts INT` to `transcription_jobs`.
* Worker updates `last_heartbeat_at` in the database every 30s during active execution.
* **Idempotency & Reclaim Logic:** When a worker receives a job message from SQS:
  * Check current DB status.
  * If `completed`: Acknowledge and delete message immediately (duplicate avoidance).
  * If `processing`: Check `last_heartbeat_at`. If older than 2 minutes (or if SQS `ApproximateReceiveCount > 1`), treat previous run as crashed, increment `attempts`, log warning, and take over processing.
  * If `attempts > max_attempts`: Mark job as `failed` in DB with reason `"Max attempts exceeded"`, delete from SQS.

---

### 4.4 Ephemeral Disk & FFmpeg Cleanup
**The Challenge:** Downloading multi-GB videos and writing extracted audio can fill ECS Fargate ephemeral storage if not cleaned up aggressively.

**The Solution:**
* Isolate scratch directories per job: `/tmp/transcription-jobs/{job_id}/`.
* Enforce strict `defer os.RemoveAll(jobDir)` immediately upon directory creation.
* Wrap FFmpeg command execution in a Go `exec.CommandContext` with a strict execution timeout (e.g., 5 minutes) to prevent runaway hanging processes.

---

## 5. Database Schema (PostgreSQL)

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Videos
CREATE TABLE videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    content_type VARCHAR(100),
    size_bytes BIGINT NOT NULL DEFAULT 0,
    duration_seconds DOUBLE PRECISION,
    status VARCHAR(50) NOT NULL DEFAULT 'uploading', -- uploading, uploaded, processing, completed, failed, deleted
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transcription Jobs
CREATE TABLE transcription_jobs (
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transcripts
CREATE TABLE transcripts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL UNIQUE REFERENCES videos(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES transcription_jobs(id) ON DELETE CASCADE,
    language VARCHAR(20) NOT NULL,
    full_text TEXT NOT NULL,
    raw_s3_key TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transcript Segments (for timestamped subtitle sync)
CREATE TABLE transcript_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transcript_id UUID NOT NULL REFERENCES transcripts(id) ON DELETE CASCADE,
    sequence_number INT NOT NULL,
    start_time DOUBLE PRECISION NOT NULL,
    end_time DOUBLE PRECISION NOT NULL,
    text TEXT NOT NULL,
    confidence DOUBLE PRECISION
);

-- Indexes
CREATE INDEX idx_videos_user_id ON videos(user_id);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_jobs_video_id ON transcription_jobs(video_id);
CREATE INDEX idx_jobs_status ON transcription_jobs(status);
CREATE INDEX idx_segments_transcript_id ON transcript_segments(transcript_id);
CREATE INDEX idx_segments_sequence ON transcript_segments(transcript_id, sequence_number);
```

---

## 6. S3 Bucket Layout & CORS

### 6.1 Object Structure
```text
transcription-service-bucket/
├── videos/{user_id}/{video_id}/original.mp4
├── audio/{user_id}/{video_id}/audio.wav
└── transcripts/{user_id}/{video_id}/
    ├── raw_transcribe_output.json
    ├── transcript.txt
    ├── transcript.srt
    └── transcript.vtt
```

### 6.2 S3 Bucket CORS Configuration
Required for direct browser `PUT` uploads:
```json
[
  {
    "AllowedHeaders": ["*"],
    "AllowedMethods": ["PUT", "GET", "HEAD"],
    "AllowedOrigins": ["http://localhost:5173", "https://yourdomain.com"],
    "ExposeHeaders": ["ETag"]
  }
]
```

---

## 7. API Design & Endpoints

Base URL: `/api/v1`

### Authentication
* `POST /auth/register` — Create account (`email`, `password`)
* `POST /auth/login` — Authenticate and receive JWT access + refresh tokens
* `POST /auth/refresh` — Refresh expired access token
* `POST /auth/logout` — Invalidate session
* `GET  /auth/me` — Return authenticated user profile

### Videos
* `POST   /videos/upload-url` — Request presigned upload URL for a new video
* `POST   /videos/:id/complete-upload` — Verify uploaded S3 object via `HeadObject` and mark `uploaded`
* `GET    /videos` — List user's videos (paginated)
* `GET    /videos/:id` — Get video metadata and playback presigned URL
* `DELETE /videos/:id` — Delete video and associated transcripts from S3 & DB

### Transcription
* `POST /videos/:id/transcribe` — Trigger asynchronous transcription job (pushes message to SQS)
* `GET  /videos/:id/transcription` — Get full transcript with timestamped segments
* `GET  /videos/:id/transcription/status` — Lightweight status check (`queued`, `processing`, `completed`, `failed`)

### Export
* `GET /videos/:id/transcript/export?format=txt` — Download raw text file
* `GET /videos/:id/transcript/export?format=srt` — Download SubRip subtitle file
* `GET /videos/:id/transcript/export?format=vtt` — Download WebVTT subtitle file

---

## 8. Frontend Architecture (HeroUI)

### 8.1 Tech Stack
* **Framework:** React with Vite (TypeScript)
* **UI Component Library:** [HeroUI](https://heroui.com/) (`@heroui/react` & `@heroui/theme`)
* **Styling:** Tailwind CSS + Lucide React Icons
* **State Management / Data Fetching:** TanStack Query (React Query) + Axios
* **Routing:** React Router v6

### 8.2 HeroUI Pages & Components Breakdown

```text
frontend/
├── src/
│   ├── components/
│   │   ├── Navbar.tsx                // HeroUI Navbar with User avatar, theme switcher & logout
│   │   ├── UploadModal.tsx           // HeroUI Modal with drag-drop & Progress bar for S3 PUT
│   │   ├── VideoCard.tsx             // HeroUI Card showing thumbnail, duration, status Chip
│   │   ├── VideoPlayer.tsx           // HTML5 video player with subtitle sync
│   │   ├── TranscriptViewer.tsx      // Clickable timestamp segment list with search filter
│   │   ├── ExportDropdown.tsx        // HeroUI Dropdown for TXT / SRT / VTT download
│   │   ├── StatusBadge.tsx           // HeroUI Chip with color states (warning, primary, success, danger)
│   │   └── ProtectedRoute.tsx        // Auth guard wrapper
│   ├── pages/
│   │   ├── LoginPage.tsx             // HeroUI Card, Input, Button, Link
│   │   ├── RegisterPage.tsx          // Form with validation
│   │   ├── DashboardPage.tsx         // Video list with HeroUI Table/Grid, Search, Upload Button
│   │   └── VideoDetailPage.tsx       // Split view: Left = VideoPlayer; Right = TranscriptViewer
│   ├── services/
│   │   ├── api.ts                    // Axios instance with JWT interceptor & auto-refresh
│   │   ├── authService.ts            // Auth API calls
│   │   ├── videoService.ts           // Video & S3 presigned upload handler
│   │   └── transcriptService.ts      // Status polling & transcript queries
│   └── App.tsx
```

### 8.3 Key Interactive UI Features
1. **Direct Upload with Real Progress:**
   * Browser requests presigned URL -> uses `axios.put(presignedUrl, file, { onUploadProgress })` -> binds directly to HeroUI `<Progress value={uploadPercent} color="primary" />`.
2. **Video & Transcript Synchronization:**
   * Playing the video automatically highlights and scrolls to the active transcript segment based on `video.currentTime`.
   * Clicking any transcript segment sets `video.currentTime = segment.startTime`.
3. **Live Job Polling:**
   * React Query auto-refreshes `GET /videos/:id/transcription/status` every 3 seconds while `status === 'processing'`, updating HeroUI `<Progress isIndeterminate />` to `<Chip color="success">Completed</Chip>`.

---

## 9. Project Directory Structure

```text
video-transcription-service/
├── cmd/
│   ├── api/
│   │   └── main.go                  // API server entrypoint
│   └── worker/
│       └── main.go                  // SQS worker entrypoint
│
├── internal/
│   ├── auth/                        // JWT, password hashing, auth service & handler
│   ├── users/                       // User repository & domain models
│   ├── videos/                      // Video metadata, presigned URL generator & verification
│   ├── transcription/               // Transcripts & segments repository & export formatters
│   ├── jobs/                        // Job state machine, heartbeat management, & reaper
│   ├── queue/                       // SQS publisher, consumer, visibility heartbeat controller
│   ├── storage/                     // AWS S3 client wrapper & presigned URL helper
│   ├── database/                    // PostgreSQL connection pool (pgxpool) & migrations
│   └── middleware/                  // Auth JWT, CORS, structured logging, recovery
│
├── pkg/
│   ├── ffmpeg/                      // Audio extraction, metadata probing, scratch cleanup
│   ├── transcribe/                  // Pluggable TranscriptionProvider (AWS Transcribe SDK)
│   ├── logger/                      // Structured JSON slog wrapper
│   └── validator/                   // Input validation helpers
│
├── migrations/                      // SQL migrations for golang-migrate
│   ├── 000001_create_users.up.sql
│   ├── 000002_create_videos.up.sql
│   ├── 000003_create_jobs.up.sql
│   └── 000004_create_transcripts.up.sql
│
├── frontend/                        // React + HeroUI Single Page Application
│   ├── src/
│   ├── package.json
│   ├── tailwind.config.js
│   ├── vite.config.ts
│   └── Dockerfile.frontend
│
├── Dockerfile.api                   // Multi-stage Go API container
├── Dockerfile.worker                // Multi-stage Go Worker container (includes FFmpeg)
├── docker-compose.yml               // Local development (API, Worker, Postgres, LocalStack/Dev S3)
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

---

## 10. Step-by-Step Phased Implementation Plan

### Phase 0: Requirements, System Contracts & Design
- [ ] Define API contract endpoints and JSON request/response formats.
- [ ] Finalize PostgreSQL DDL migrations with indexes and foreign keys.
- [ ] Define SQS message schema:
  ```json
  {
    "job_id": "uuid",
    "video_id": "uuid",
    "user_id": "uuid",
    "s3_video_key": "videos/user/video/original.mp4",
    "language": "en-US",
    "attempt": 1
  }
  ```

---

### Phase 1: Go API Foundation
- [ ] Set up Go project module and directory structure.
- [ ] Configure environment variable loading (`godotenv` + typed config struct).
- [ ] Set up structured JSON logger with Go standard `log/slog`.
- [ ] Initialize Gin engine with global recovery, structured request logging, and CORS middleware.
- [ ] Implement graceful shutdown handling `SIGINT`/`SIGTERM` with context timeout.
- [ ] Create `/health` and `/ready` endpoints.

---

### Phase 2: PostgreSQL Database & Connection Pooling
- [ ] Initialize `pgxpool` with max connection limits and health checks.
- [ ] Integrate `golang-migrate` for automatic schema migrations on startup.
- [ ] Create initial migration files (`users`, `videos`, `transcription_jobs`, `transcripts`, `transcript_segments`).
- [ ] Implement database health check in `/ready`.

---

### Phase 3: Authentication & Authorization (JWT)
- [ ] Implement password hashing using `bcrypt`.
- [ ] Implement user repository with `CreateUser` and `GetByEmail`.
- [ ] Implement JWT token generation (access token + refresh token pair).
- [ ] Build Gin JWT Auth middleware extracting and verifying bearer tokens.
- [ ] Implement endpoints:
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `GET  /api/v1/auth/me`

---

### Phase 4: AWS S3 & Presigned Uploads
- [ ] Initialize AWS SDK for Go v2 S3 Client.
- [ ] Implement S3 presigned URL generator for `PUT` operations with expiration.
- [ ] Implement `s3.HeadObject` verifier to validate uploaded object existence, content type, and byte size.
- [ ] Configure S3 bucket CORS for direct browser PUT access.

---

### Phase 5: Video Management Module
- [ ] Implement video repository (CRUD operations scoped by `user_id`).
- [ ] Build `POST /api/v1/videos/upload-url` (generates unique UUID, DB record, and S3 Presigned URL).
- [ ] Build `POST /api/v1/videos/:id/complete-upload` (runs `HeadObject` check, updates status to `uploaded`).
- [ ] Build `GET /api/v1/videos` (list user videos) and `GET /api/v1/videos/:id` (with playback URL).
- [ ] Build `DELETE /api/v1/videos/:id` (cleans up DB & S3 objects).

---

### Phase 6: SQS Job Queue Integration
- [ ] Initialize AWS SDK for Go v2 SQS Client.
- [ ] Implement SQS publisher in API service to push transcription messages.
- [ ] Implement SQS consumer in worker with long-polling (`WaitTimeSeconds: 20`).
- [ ] Implement `POST /api/v1/videos/:id/transcribe` (creates job record in DB -> publishes to SQS).

---

### Phase 7: Resilient Worker Engine & Crash Recovery
- [ ] Create `cmd/worker/main.go` background loop.
- [ ] Implement **Visibility Timeout Extension Goroutine**: worker calls `sqs.ChangeMessageVisibility` every 20s while job is active.
- [ ] Implement **Job Heartbeat & Reclaim Logic**: worker updates `last_heartbeat_at` in DB; recovers crashed/zombie jobs if heartbeat is stale.
- [ ] Enforce idempotency: skip or safely acknowledge messages for already-completed jobs.
- [ ] Implement DLQ redrive handling when job fails max retry attempts.

---

### Phase 8: FFmpeg Subprocess & Scratch Disk Management
- [ ] Implement Go FFmpeg wrapper using `os/exec.CommandContext` with strict execution timeouts.
- [ ] Create isolated scratch workspace per job (`/tmp/transcription-jobs/{job_id}`).
- [ ] Enforce strict cleanup with `defer os.RemoveAll(scratchDir)`.
- [ ] Extract normalized 16kHz mono audio WAV/MP3 from video.
- [ ] Upload extracted audio to S3 (`audio/{user_id}/{video_id}/audio.wav`).

---

### Phase 9: AWS Transcribe Provider
- [ ] Define `TranscriptionProvider` interface in Go.
- [ ] Implement `AWSTranscribeProvider` using AWS SDK v2 (`StartTranscriptionJob`, `GetTranscriptionJob`).
- [ ] Implement polling loop with backoff while keeping SQS visibility heartbeat active.
- [ ] Fetch completed transcription JSON output from S3.
- [ ] Parse and normalize raw AWS Transcribe JSON into `Transcript` and `TranscriptSegment` structs.
- [ ] Save transcript and segments to PostgreSQL within a database transaction.
- [ ] Acknowledge and delete message from SQS upon successful completion.

---

### Phase 10: Transcript Export Formats (TXT, SRT, VTT)
- [ ] Implement SRT subtitle formatter (`00:01:20,000 --> 00:01:24,500\nText\n`).
- [ ] Implement WebVTT subtitle formatter (`00:01:20.000 --> 00:01:24.500\nText\n`).
- [ ] Implement Plain Text formatter.
- [ ] Build endpoints:
  - `GET /api/v1/videos/:id/transcription`
  - `GET /api/v1/videos/:id/transcription/status`
  - `GET /api/v1/videos/:id/transcript/export?format=txt|srt|vtt`

---

### Phase 11: Web Frontend with HeroUI (heroui.com)
- [ ] Initialize React + TypeScript + Vite project in `frontend/`.
- [ ] Install and configure Tailwind CSS and HeroUI (`@heroui/react`, `@heroui/theme`, `framer-motion`).
- [ ] Set up HeroUI dark/light theme provider and responsive layout.
- [ ] Build **Authentication Pages**:
  - HeroUI `Card`, `Input`, and `Button` for Login and Register with form validation.
  - Store JWT tokens and configure Axios request/response interceptors with automatic token refresh.
- [ ] Build **Dashboard & Video Library**:
  - HeroUI `Navbar` with user profile avatar and logout.
  - Video library using HeroUI `Table` or `Card` grid showing filename, duration, creation date, and status `Chip`.
  - Search and filter controls.
- [ ] Build **Direct S3 Upload Modal**:
  - HeroUI `Modal` with drag-and-drop file picker.
  - Stream file directly to S3 presigned URL with real-time byte progress bound to HeroUI `<Progress />`.
  - Call `complete-upload` on finish and automatically refresh dashboard.
- [ ] Build **Video Player & Synchronized Transcript Viewer**:
  - Left pane: HTML5 video player.
  - Right pane: Clickable transcript segment list with active highlight auto-scrolling to match current video playback time.
  - Clicking any segment seeks video to exact timestamp.
  - HeroUI `Dropdown` to export TXT, SRT, or VTT files.
- [ ] Build **Live Job Status Indicator**:
  - Poll `/transcription/status` while state is `queued` or `processing`, showing animated HeroUI `Progress` or `Skeleton` until completion.

---

### Phase 12: Docker & Containerization
- [ ] Write `Dockerfile.api`: Multi-stage Go build resulting in minimal scratch/alpine runtime image.
- [ ] Write `Dockerfile.worker`: Multi-stage Go build with FFmpeg installed and non-root execution.
- [ ] Write `Dockerfile.frontend`: Multi-stage Node build with Nginx for local static serving.
- [ ] Write `docker-compose.yml` orchestrating API, Worker, Frontend, and PostgreSQL for 1-click local development.

---

### Phase 13: Amazon ECR (Elastic Container Registry)
- [ ] Create ECR repositories for `transcription-api` and `transcription-worker`.
- [ ] Set up GitHub Actions or build scripts to tag images with Git SHA and push to ECR.

---

### Phase 14: AWS Networking & VPC Architecture
- [ ] Provision VPC with public and private subnets across 2 Availability Zones.
- [ ] Configure Internet Gateway for public subnets and NAT Gateway for private subnets.
- [ ] Configure Security Groups:
  - ALB Security Group: Inbound 80/443 from Internet.
  - ECS API Security Group: Inbound from ALB only.
  - ECS Worker Security Group: Outbound to S3/SQS/Transcribe via NAT Gateway.
  - RDS Security Group: Inbound port 5432 from ECS API and Worker security groups only (no public access).

---

### Phase 15: Amazon RDS PostgreSQL
- [ ] Provision RDS PostgreSQL instance inside private DB subnet group.
- [ ] Configure automated daily backups, encryption at rest (KMS), and SSL connections.
- [ ] Run database migrations against RDS.

---

### Phase 16: ECS Fargate Deployment (API & Worker Services)
- [ ] Create ECS Cluster.
- [ ] Define API Task Definition & ECS Service (placed behind ALB).
- [ ] Define Worker Task Definition & ECS Service (standalone background consumer).
- [ ] Allocate CPU/Memory and configure logging to CloudWatch.

---

### Phase 17: Application Load Balancer (ALB)
- [ ] Create ALB in public subnets.
- [ ] Configure Target Group pointing to ECS API tasks with `/health` health checks.
- [ ] Configure HTTPS listener with AWS Certificate Manager (ACM) SSL certificate.

---

### Phase 18: AWS Secrets Manager & IAM Least Privilege
- [ ] Store `DATABASE_URL`, `JWT_SECRET`, etc. in AWS Secrets Manager.
- [ ] Configure ECS Task Execution Role to fetch secrets securely at startup.
- [ ] Grant ECS API Task Role permissions only for S3 Presigned URLs, SQS SendMessage, and Secrets Manager.
- [ ] Grant ECS Worker Task Role permissions only for S3 Get/Put, SQS Receive/Delete/ChangeVisibility, and Transcribe.

---

### Phase 19: Static Frontend Hosting (S3 + CloudFront)
- [ ] Create private S3 bucket for static frontend assets.
- [ ] Configure Amazon CloudFront Distribution with Origin Access Control (OAC):
  - Default route (`/*`): Serves HeroUI React static build from S3.
  - API route (`/api/*`): Forwards requests directly to the ALB.
- [ ] Configure custom domain in Route 53 with ACM SSL certificate.

---

### Phase 20: CloudWatch Observability & Worker Autoscaling
- [ ] Configure CloudWatch JSON log streaming with log retention policies.
- [ ] Create CloudWatch Metric Alarm monitoring SQS `ApproximateNumberOfMessagesVisible`.
- [ ] Configure ECS Target Tracking / Step Scaling policy to automatically scale worker containers based on SQS queue depth.
- [ ] Set up CloudWatch alarms for API 5xx errors and Dead-Letter Queue message counts.

---

### Phase 21: End-to-End Testing & Verification
- [ ] Unit tests for services, auth JWT, SRT/VTT formatters, and FFmpeg command parser.
- [ ] Integration tests verifying API -> S3 -> SQS -> Worker -> RDS flow.
- [ ] End-to-end browser test: Register -> Login -> Direct S3 Video Upload -> Transcribe -> Video Player sync -> SRT export.

---

## 11. Definition of Done & Success Checklist

- [x] **Direct S3 Uploads**: Videos upload directly to S3 without saturating API bandwidth; verified with `HeadObject`.
- [x] **Async Queue Decoupling**: API never blocks on FFmpeg or transcription; SQS buffers all load.
- [x] **Zero Zombie Jobs**: SQS Visibility Timeout heartbeat and DB `last_heartbeat_at` prevent stuck or orphaned jobs.
- [x] **Scratch Disk Safety**: FFmpeg temporary files are strictly cleaned up via `defer` handlers.
- [x] **Modern UI**: Full-featured HeroUI frontend with video-transcript synchronization and direct upload progress.
- [x] **Cloud Native**: Isolated VPC, private RDS, IAM least-privilege, ECS autoscaling, and CloudFront unified routing.
