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
- Chi/net/http web routing & Clean Architecture
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
               +-----------------------------------------------------------+
               |                   React + HeroUI Client                   |
               +-----------------------------------------------------------+
                     /                                  \
    1. Request Presigned PUT                       2. Direct Binary Stream
                   /                                      \
                  v                                        v
     +-------------------------+                 +--------------------+
     |    Go HTTP REST API     |                 |   AWS S3 Storage   |
     | (net/http + chi + auth) |                 | (Raw Media & Audio)|
     +-------------------------+                 +--------------------+
                 |                                         ^
                 | 3. Enqueue Job Payload                  |
                 v                                         |
     +-------------------------+                           |
     |      AWS SQS Queue      |                           |
     +-------------------------+                           |
                 |                                         | 5. Download Video /
                 | 4. 20s Long Poll & Heartbeat            |    Upload 16kHz WAV
                 v                                         v
     +-----------------------------------------------------------------+
     |                       Go Worker Daemon                          |
     |  - Concurrent Visibility Heartbeat Goroutine (ticks every 20s)  |
     |  - Ephemeral Scratch Directory: /tmp/transcription-jobs/{id}/   |
     |  - FFmpeg Audio Extraction -> 16,000Hz 16-bit Mono WAV          |
     |  - AWS Transcribe Provider / Mock Provider Fallback             |
     |  - Word-level Timestamp Parser -> Subtitle Segments             |
     +-----------------------------------------------------------------+
                 |
                 | 6. Atomic Batch Commit (WithTx)
                 v
     +-------------------------+
     |   PostgreSQL 16 DB      |
     | (Users, Videos, Jobs,   |
     |  Transcripts, Segments) |
     +-------------------------+
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

---

## 5. Phased Implementation Status

- [x] **Phase 0: Requirements, System Contracts & Design** (DDL, API contracts, SQS schema)
- [x] **Phase 1: Go API Foundation** (`slog`, config loader, CORS, recovery, `/health`, `/ready`)
- [x] **Phase 2: PostgreSQL Database & Migrations** (`pgxpool`, `golang-migrate` embedded SQL, dynamic readiness check)
- [x] **Phase 3: Authentication & Authorization** (bcrypt, signed JWT access/refresh tokens, user repo, auth middleware)
- [x] **Phase 4: AWS S3 & Presigned Uploads** (AWS SDK v2 S3 client, presigned PUT/GET, `HeadObject` verifier, mock storage)
- [x] **Phase 5: Video Management Module** (Video repository, presigned upload generation, `HeadObject` upload check, playback links)
- [x] **Phase 6: SQS Job Queue Integration** (AWS SDK v2 SQS, long polling, `TranscriptionMessage` schema, `POST /videos/:id/transcribe`)
- [x] **Phase 7: Resilient Worker Engine** (Long polling worker, visibility timeout heartbeat loop every 20s, idempotency, `cmd/worker`)
- [x] **Phase 8: FFmpeg & Scratch Disk Management** (16kHz mono WAV extraction, `ffprobe` duration check, `/tmp` scratch cleanup)
- [x] **Phase 9: AWS Transcribe Provider & Pipeline** (Transcribe SDK, word-level JSON output parser, DB transaction persistence)
- [x] **Phase 10: Transcript Features & Export Formats** (SRT, WebVTT, TXT subtitle formatters, download endpoints)
- [x] **Phase 11: Web Frontend with HeroUI & Architecture Showcase**
  - [x] React + TypeScript + Vite initialized in `frontend/`.
  - [x] Tailwind CSS & HeroUI (`@heroui/react`, `@heroui/theme`, `framer-motion`) configured.
  - [x] Authentication screens (Login/Register with JWT storage and Axios refresh interceptor).
  - [x] Dashboard & Video Library (HeroUI Grid, Search, Status Badges, Delete modal).
  - [x] Direct-to-S3 Upload Modal with byte progress bar and instant job dispatch.
  - [x] Video Player & Synchronized Interactive Transcript Viewer (auto-scroll highlighting, click-to-seek, confidence chips).
  - [x] Subtitle Export Dropdown (1-click `.srt`, `.vtt`, `.txt` download).
  - [x] **Interactive Architecture & System Design Showcase** (`/architecture` page with 5 interactive tabs).
- [x] **Phase 12: Docker & Containerization** (`Dockerfile.api`, `Dockerfile.worker`, `Dockerfile.frontend`, `docker-compose.yml`, `Makefile`, `scripts/dev.bat`)
- [x] **Phase 13: Project Documentation & Architecture Walkthrough** (Comprehensive `README.md` & `walkthrough.md`)
