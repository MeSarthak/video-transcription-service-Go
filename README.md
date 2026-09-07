# 🎥 TranscribeX — Distributed Video Transcription Engine

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://postgresql.org)
[![AWS S3](https://img.shields.io/badge/AWS-S3-569A31?style=flat&logo=amazons3)](https://aws.amazon.com/s3/)
[![AWS SQS](https://img.shields.io/badge/AWS-SQS-FF4F8B?style=flat&logo=amazonsqs)](https://aws.amazon.com/sqs/)
[![FFmpeg](https://img.shields.io/badge/FFmpeg-6.0-007808?style=flat&logo=ffmpeg)](https://ffmpeg.org)
[![HeroUI](https://img.shields.io/badge/HeroUI-React-000000?style=flat)](https://heroui.com)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)](https://docker.com)

A high-performance, asynchronous distributed video transcription service engineered in **Go**, featuring **direct-to-S3 client streaming uploads**, **SQS message queueing**, **resilient background workers with visibility heartbeat renewal**, **FFmpeg audio extraction/normalization**, **AWS Transcribe integration**, and a modern **HeroUI web dashboard** with a dedicated **Interactive Architecture & System Design Showcase (`/architecture`)**.

---

## 🌟 Architecture & Key System Design Decisions

```
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

### 💎 Distributed Systems Highlights

1. **Zero-RAM API Gateway Bypass (Direct S3 Streaming):**
   - The Go API issues HMAC-SHA256 presigned PUT URLs with 15-minute validity. Clients stream multi-gigabyte video files directly to S3. The API server memory usage remains flat regardless of concurrent video uploads.
2. **Decoupled Asynchronous Worker Engine:**
   - Media transcoding is decoupled from HTTP requests using Amazon SQS. Workers utilize 20-second long polling to minimize empty API calls.
3. **SQS Visibility Heartbeat Goroutine:**
   - Large video transcriptions can take minutes. A dedicated goroutine periodically extends SQS message visibility every 20 seconds, preventing duplicate message dispatch while preventing stale lockups if a worker crashes.
4. **Idempotency & Retry Limits:**
   - Workers verify job states atomically before processing. Jobs exceeding 3 retries are safely failed, preventing poison pills from halting queue consumers.
5. **Audio Normalization Pipeline:**
   - High-bitrate audio tracks are normalized via FFmpeg to **16,000 Hz, 16-bit single-channel (mono) PCM WAV**, dramatically reducing bandwidth and speech engine processing overhead.
6. **Scratch Disk Ephemeral Isolation:**
   - Isolated scratch directories (`/tmp/transcription-jobs/{job_id}/`) are automatically wiped with `defer os.RemoveAll()`, ensuring 100% disk reclamation.
7. **Zero-Cloud Local Test Doubles:**
   - Built-in thread-safe mock adapters (`MockStorage`, `MockQueue`, `MockTranscribeProvider`) allow complete end-to-end local development and testing without needing AWS credentials.
8. **Interactive Architecture Showcase (`/architecture`):**
   - Built directly into the frontend with interactive step-by-step pipeline visualizers, schema explorers, and tradeoff matrices to present system design principles.

---

## 🚀 Quick Start (Zero Cloud Dependencies)

### Option A: Run with Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-username/video-transcription-service.git
cd video-transcription-service

# Start PostgreSQL, API, Worker, and Frontend in Docker
docker-compose up -d --build
```
- **Web App & Architecture Showcase:** [http://localhost:3000](http://localhost:3000)
- **API Server:** [http://localhost:8080](http://localhost:8080)
- **Health Check:** [http://localhost:8080/health](http://localhost:8080/health)

---

### Option B: Local Native Development

#### 1. Setup Environment
```bash
cp .env.example .env
```

#### 2. Start PostgreSQL (e.g. via Docker)
```bash
docker run -d --name local-pg -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=transcription_db postgres:16-alpine
```

#### 3. Run Backend API & Worker
```bash
# Terminal 1: Start API Server
go run ./cmd/api

# Terminal 2: Start Background Worker
go run ./cmd/worker
```

#### 4. Run Frontend
```bash
cd frontend
npm install --legacy-peer-deps
npm run dev
```
Open [http://localhost:5173](http://localhost:5173) in your browser.

---

## 📡 REST API Reference

### Authentication
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Register new user account (`email`, `password`) |
| `POST` | `/api/v1/auth/login` | Authenticate and obtain JWT Access + Refresh token pair |
| `POST` | `/api/v1/auth/refresh` | Refresh expired access token using refresh token |
| `GET` | `/api/v1/auth/me` | Fetch authenticated user profile (Bearer token) |

### Video Management & Upload Pipeline
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/v1/videos/upload-url` | Generate S3 presigned PUT URL and register video |
| `POST` | `/api/v1/videos/:id/complete-upload` | Validate S3 upload completion via `HeadObject` |
| `GET` | `/api/v1/videos` | List all videos owned by user |
| `GET` | `/api/v1/videos/:id` | Get video metadata & presigned playback URL |
| `DELETE` | `/api/v1/videos/:id` | Delete video, S3 object, jobs, and all transcript segments |

### Transcription & Export
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/v1/videos/:id/transcribe` | Enqueue asynchronous transcription job into SQS |
| `GET` | `/api/v1/videos/:id/transcription/status` | Check job execution status and retry counters |
| `GET` | `/api/v1/videos/:id/transcription` | Fetch complete transcript and timestamped subtitle segments |
| `GET` | `/api/v1/videos/:id/transcript/export?format=srt` | Export subtitle file (`srt`, `vtt`, `txt`) |

---

## 🧪 Testing

Run the full backend test suite:
```bash
go test -v ./...
```

Run test suite with coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

---

## 📂 Project Structure

```
video-transcription-service/
├── cmd/
│   ├── api/                  # Go REST API main entrypoint
│   └── worker/               # Go Background Worker daemon main entrypoint
├── internal/
│   ├── auth/                 # Authentication handlers & services
│   ├── config/               # Typed environment configuration
│   ├── database/             # PostgreSQL pgxpool & schema migrator
│   ├── health/               # /health and /ready probes
│   ├── jobs/                 # SQS job dispatching & status tracking
│   ├── middleware/           # CORS, structured slog logging, panic recovery
│   ├── queue/                # SQS client & thread-safe MockQueue
│   ├── storage/              # S3 client & thread-safe MockStorage
│   ├── transcription/        # Subtitle formatters (.srt, .vtt, .txt) & export
│   ├── users/                # User repository & authentication logic
│   ├── videos/               # Video metadata CRUD & S3 validation
│   └── worker/               # Worker polling loop, heartbeat & pipeline processor
├── pkg/
│   ├── crypto/               # Bcrypt password hashing
│   ├── ffmpeg/               # FFmpeg normalization & probe wrappers
│   ├── jwt/                  # HS256 JWT tokens & claims
│   ├── logger/               # Standard library structured slog wrapper
│   └── transcribe/           # AWS Transcribe provider, word parser & mock
├── migrations/               # Embedded SQL migrations (000001 - 000005)
├── frontend/                 # Vite + React 19 + TypeScript + HeroUI + Tailwind
│   ├── src/
│   │   ├── components/       # Navbar, VideoPlayer, TranscriptViewer, UploadModal
│   │   ├── context/          # AuthContext, ThemeContext
│   │   ├── pages/            # DashboardPage, VideoDetailPage, ArchitecturePage
│   │   ├── services/         # Axios client with JWT interceptors
│   │   └── types/            # TypeScript domain interfaces
├── Dockerfile.api            # Multi-stage Docker build for API
├── Dockerfile.worker         # Multi-stage Docker build for Worker with FFmpeg
├── Dockerfile.frontend       # Multi-stage Docker build for Frontend (Nginx SPA)
├── docker-compose.yml        # Full-stack container orchestration
├── Makefile                  # Build, test, and dev automation
└── README.md
```

---

## 📜 License
Distributed under the MIT License.
