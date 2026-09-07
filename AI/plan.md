# Video Transcription Service — AWS + Go + HeroUI Frontend

## 1. Project Overview

Build a production-oriented video transcription platform with a **Go backend**, **AWS cloud infrastructure**, and a modern **HeroUI (React) frontend** with a built-in **Interactive Architecture & System Design Showcase**.

The service allows users to:

- Create accounts and authenticate securely (JWT access & refresh tokens).
- Upload video files directly from the browser to Amazon S3 via presigned URLs with real-time byte progress tracking.
- Store videos securely in private S3 buckets (never public).
- Asynchronously extract audio with FFmpeg in a dedicated background worker.
- Transcribe audio with AWS Transcribe (or pluggable speech-to-text providers).
- Store transcripts and timestamped segments in PostgreSQL.
- Monitor job progress with live status polling, heartbeat extensions, and resilient error/retry handling.
- View an interactive video player synchronized with clickable transcript timestamps and auto-scrolling highlights.
- Download/export transcripts as TXT, SRT, and VTT.
- Explore the **Interactive Architecture & System Design Showcase** directly inside the web UI, explaining the distributed systems decisions, database schema, SQS message flow, and cloud infrastructure.
- Scale transcription workers independently from the API.

---

## 2. Main Learning Goals

### Go / Backend
- Gin web framework & Clean Architecture
- REST API design & middleware (Auth, CORS, Structured Logging, Panic Recovery)
- Authentication (bcrypt password hashing + JWT access & refresh tokens)
- PostgreSQL & `pgxpool` connection pooling
- Database migrations with `golang-migrate` and embedded SQL
- Database transactions (`WithTx`) for atomic transcript persistence
- Direct-to-S3 presigned URLs & server-side object verification (`HeadObject`)
- SQS message consumption, long polling, visibility timeout management, & Dead-Letter Queues (DLQ)
- Worker crash resilience, heartbeats (`last_heartbeat_at`), and idempotency
- FFmpeg subprocess orchestration, audio normalization (16kHz mono WAV), and scratch disk cleanup
- Pluggable transcription provider interface (`AWS Transcribe`, `MockTranscribe`)
- Graceful shutdown & context cancellation
- Structured JSON logging (`slog`)

### Frontend (HeroUI + React)
- Modern UI development with [HeroUI](https://heroui.com/) (`@heroui/react` & `@heroui/theme`) and Tailwind CSS
- Client-side direct-to-S3 streaming uploads with accurate progress bars
- Interactive HTML5 video player synchronized with timestamped transcript cues
- Click-to-seek timestamp navigation
- Live job status polling and UI state transitions
- **Interactive Architecture & System Design Showcase** (`/architecture`)
- Responsive dark/light theme integration

### AWS & Cloud Architecture
- **IAM**: Least-privilege IAM roles for ECS Tasks (API vs. Worker)
- **S3**: Private buckets, presigned URLs, CORS configuration, lifecycle rules
- **SQS**: Standard queue, visibility timeout extension loop, redrive policy to DLQ
- **RDS**: Managed PostgreSQL in isolated private subnets
- **ECS Fargate**: Containerized API and Worker services with independent autoscaling
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

## 4. Key Architectural Principles & Distributed Systems Decisions

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
* Worker runs a background goroutine during processing that periodically calls `sqs.ChangeMessageVisibility` every 20s (extending visibility by 60s).
* When Transcribe finishes (or fails), the goroutine stops, worker writes results to DB, and deletes the SQS message (`sqs.DeleteMessage`).
* If the worker crashes, the heartbeat stops $\rightarrow$ visibility timeout expires $\rightarrow$ SQS redelivers $\rightarrow$ after `maxReceiveCount` (3 attempts), message routes to the DLQ.

---

### 4.3 Worker Crash & Zombie Job Handling
**The Challenge:** If a worker container dies midway, the database job record might stay stuck in `processing`.

**The Solution:**
* Add `last_heartbeat_at TIMESTAMP` and `attempts INT` to `transcription_jobs`.
* Worker updates `last_heartbeat_at` in the database every 20s during active execution.
* **Idempotency & Reclaim Logic:** When a worker receives a job message from SQS:
  * If `completed`: Acknowledge and delete message immediately (duplicate avoidance).
  * If `processing`: Check `last_heartbeat_at`. If older than 2 minutes (or SQS `ApproximateReceiveCount > 1`), treat previous run as crashed, increment `attempts`, log warning, and take over processing.
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Transcripts
CREATE TABLE transcripts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL UNIQUE REFERENCES videos(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES transcription_jobs(id) ON DELETE CASCADE,
    language VARCHAR(20) NOT NULL,
    full_text TEXT NOT NULL,
    raw_s3_key TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
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
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_videos_user_id ON videos(user_id);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_videos_created_at ON videos(created_at DESC);
CREATE INDEX idx_jobs_video_id ON transcription_jobs(video_id);
CREATE INDEX idx_jobs_user_id ON transcription_jobs(user_id);
CREATE INDEX idx_jobs_status ON transcription_jobs(status);
CREATE INDEX idx_jobs_heartbeat ON transcription_jobs(status, last_heartbeat_at);
CREATE INDEX idx_transcripts_video_id ON transcripts(video_id);
CREATE INDEX idx_transcripts_job_id ON transcripts(job_id);
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

## 8. Frontend Architecture & Architecture Showcase ([HeroUI](https://heroui.com/))

### 8.1 Tech Stack
* **Framework:** React with Vite (TypeScript)
* **UI Component Library:** [HeroUI](https://heroui.com/) (`@heroui/react` & `@heroui/theme`)
* **Styling:** Tailwind CSS + Lucide React Icons + Framer Motion
* **State Management / Data Fetching:** TanStack Query (React Query) + Axios
* **Routing:** React Router v6

### 8.2 Frontend Directory Structure

```text
frontend/
├── src/
│   ├── components/
│   │   ├── Navbar.tsx                // HeroUI Navbar with User avatar, theme switcher & Architecture link
│   │   ├── UploadModal.tsx           // HeroUI Modal with drag-drop & Progress bar for S3 PUT
│   │   ├── VideoCard.tsx             // HeroUI Card showing thumbnail, duration, status Chip
│   │   ├── VideoPlayer.tsx           // HTML5 video player with subtitle sync
│   │   ├── TranscriptViewer.tsx      // Clickable timestamp segment list with search filter
│   │   ├── ExportDropdown.tsx        // HeroUI Dropdown for TXT / SRT / VTT download
│   │   ├── StatusBadge.tsx           // HeroUI Chip with color states (warning, primary, success, danger)
│   │   └── ProtectedRoute.tsx        // Auth guard wrapper
│   │
│   ├── pages/
│   │   ├── LoginPage.tsx             // HeroUI Card, Input, Button, Link
│   │   ├── RegisterPage.tsx          // Form with validation
│   │   ├── DashboardPage.tsx         // Video list with HeroUI Table/Grid, Search, Upload Button
│   │   ├── VideoDetailPage.tsx       // Split view: Left = VideoPlayer; Right = TranscriptViewer
│   │   └── ArchitecturePage.tsx      // 🏛️ Dedicated Interactive Architecture & System Design Showcase
│   │
│   ├── services/
│   │   ├── api.ts                    // Axios instance with JWT interceptor & auto-refresh
│   │   ├── authService.ts            // Auth API calls
│   │   ├── videoService.ts           // Video & S3 presigned upload handler
│   │   └── transcriptService.ts      // Status polling & transcript queries
│   │
│   └── App.tsx
```

### 8.3 🏛️ Interactive Architecture Showcase (`/architecture`)
A dedicated interactive showcase directly accessible in the web app:
1. **Interactive System Architecture Flow:** Visual step-by-step request tracer (Browser $\rightarrow$ CloudFront $\rightarrow$ ALB $\rightarrow$ Go API $\rightarrow$ S3/SQS/RDS $\rightarrow$ Go Worker $\rightarrow$ FFmpeg/Transcribe).
2. **Key Engineering Challenges & Trade-offs:** Detailed interactive cards explaining:
   * *Presigned Uploads vs. API Bandwidth Saturation*
   * *SQS Visibility Timeout vs. Transcribe Latency (20s Heartbeat)*
   * *Zero Zombie Jobs (Heartbeats & SQS Redrive)*
   * *Ephemeral Disk Hygiene (`defer` scratch cleanup)*
   * *Server-side Upload Verification (`s3.HeadObject`)*
3. **Database Schema Explorer:** Interactive table inspector for `users`, `videos`, `jobs`, `transcripts`, `segments` with foreign keys and index strategies.
4. **Message & API Contract Inspector:** Interactive JSON payload explorer for SQS and REST contracts.
5. **Tech Stack & Cloud Topology:** Badges, architecture tiers, and cloud security boundaries.

> [!IMPORTANT]
> **Continuous Architecture Updates:** Whenever any new phase (Docker, VPC, ECS, RDS, CloudFront, Autoscaling) is implemented, the `/architecture` showcase in the frontend will be updated with live architecture specs and deployment diagrams!

---

## 9. Phased Implementation Roadmap

- [x] **Phase 0: Requirements, System Contracts & Design** (DDL, API contracts, SQS schema)
- [x] **Phase 1: Go API Foundation** (Gin, `slog`, config loader, CORS, recovery, `/health`, `/ready`)
- [x] **Phase 2: PostgreSQL Database & Migrations** (`pgxpool`, `golang-migrate` embedded SQL, dynamic readiness check)
- [x] **Phase 3: Authentication & Authorization** (bcrypt, signed JWT access/refresh tokens, user repo, Gin auth middleware)
- [x] **Phase 4: AWS S3 & Presigned Uploads** (AWS SDK v2 S3 client, presigned PUT/GET, `HeadObject` verifier, mock storage)
- [x] **Phase 5: Video Management Module** (Video repository, presigned upload generation, `HeadObject` upload check, playback links)
- [x] **Phase 6: SQS Job Queue Integration** (AWS SDK v2 SQS, long polling, `TranscriptionMessage` schema, `POST /videos/:id/transcribe`)
- [x] **Phase 7: Resilient Worker Engine** (Long polling worker, visibility timeout heartbeat loop every 20s, idempotency, `cmd/worker`)
- [x] **Phase 8: FFmpeg & Scratch Disk Management** (16kHz mono WAV extraction, `ffprobe` duration check, `/tmp` scratch cleanup)
- [x] **Phase 9: AWS Transcribe Provider & Pipeline** (Transcribe SDK, word-level JSON output parser, DB transaction persistence)
- [x] **Phase 10: Transcript Features & Export Formats** (SRT, WebVTT, TXT subtitle formatters, download endpoints)
- [ ] **Phase 11: Web Frontend with HeroUI ([heroui.com](https://heroui.com/)) & Architecture Showcase**
  - [ ] Initialize React + TypeScript + Vite project in `frontend/`.
  - [ ] Configure Tailwind CSS and HeroUI (`@heroui/react`, `@heroui/theme`, `framer-motion`).
  - [ ] Build Authentication screens (Login, Register with validation & token management).
  - [ ] Build Dashboard & Video Library (HeroUI Table/Grid, Search, Status Badges).
  - [ ] Build Direct-to-S3 Upload Modal with real-time byte progress bar.
  - [ ] Build Video Player & Synchronized Interactive Transcript Viewer (auto-scroll highlighting, click-to-seek).
  - [ ] Build Subtitle Export Dropdown (1-click `.srt`, `.vtt`, `.txt` download).
  - [ ] Build **Interactive Architecture & System Design Showcase** (`/architecture` page with 5 interactive tabs).
- [ ] **Phase 12: Docker & Containerization** (`Dockerfile.api`, `Dockerfile.worker`, `Dockerfile.frontend`, `docker-compose.yml`)
- [ ] **Phase 13: Amazon ECR** (Image versioning and registry)
- [ ] **Phase 14: AWS Networking & VPC Architecture** (Public/private subnets, NAT Gateway, Security Groups)
- [ ] **Phase 15: Amazon RDS PostgreSQL** (Managed private DB)
- [ ] **Phase 16: ECS Fargate Deployment** (API & Worker services)
- [ ] **Phase 17: Application Load Balancer (ALB)** (SSL termination, health checks)
- [ ] **Phase 18: AWS Secrets Manager & IAM Least Privilege** (Task roles)
- [ ] **Phase 19: Static Frontend Hosting (S3 + CloudFront)** (Unified `/api/*` and `/*` routing)
- [ ] **Phase 20: CloudWatch Observability & Worker Autoscaling** (SQS queue depth scaling)
- [ ] **Phase 21: End-to-End Testing & Verification**
