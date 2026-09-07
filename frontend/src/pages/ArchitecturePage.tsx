import React, { useState } from 'react';
import {
  Tabs,
  Tab,
  Card,
  CardBody,
  CardHeader,
  Button,
  Chip,
  Divider,
} from '@heroui/react';
import {
  Cpu,
  Layers,
  Database,
  ShieldCheck,
  Zap,
  ArrowRight,
  ArrowLeft,
  Server,
  HardDrive,
  MessageSquare,
  Activity,
  Code2,
  FileCheck2,
  RefreshCw,
  Sparkles,
  GitBranch,
  Network,
} from 'lucide-react';
import { ArchitectureDiagrams } from '../components/ArchitectureDiagrams';

export const ArchitecturePage: React.FC = () => {
  const [activeStep, setActiveStep] = useState(0);

  const pipelineSteps = [
    {
      id: 'upload',
      title: '1. S3 Presigned Direct Upload',
      tagline: 'Zero-Memory API Gateway Bypass',
      icon: HardDrive,
      color: 'primary',
      description:
        'Client requests a cryptographically signed HMAC-SHA256 upload URL from the Go API. The client streams raw video bytes directly into AWS S3, completely bypassing the backend server to prevent memory exhaustion and bandwidth saturation.',
      codeSnippet: `// 1. Client requests upload URL -> Go API generates S3 Presign PUT
req, err := s3PresignClient.PresignPutObject(ctx, &s3.PutObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(s3Key),
}, s3.WithPresignExpires(15 * time.Minute))`,
      metrics: [
        { label: 'API Memory Saved', value: '100%' },
        { label: 'Max File Size', value: '500 MB' },
        { label: 'Expiration', value: '15 Minutes' },
      ],
    },
    {
      id: 'enqueue',
      title: '2. SQS Message Enqueueing',
      tagline: 'Decoupled Asynchronous Job Distribution',
      icon: MessageSquare,
      color: 'secondary',
      description:
        'Upon upload completion, the API validates the S3 object via HeadObject and pushes a lightweight JSON job payload into AWS SQS. Jobs are decoupled from API request lifecycles for horizontal worker scalability.',
      codeSnippet: `// 2. Dispatch job payload to Amazon SQS FIFO/Standard queue
_, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
    QueueUrl:    aws.String(queueURL),
    MessageBody: aws.String(string(jobPayloadJSON)),
})`,
      metrics: [
        { label: 'Queue Type', value: 'AWS SQS' },
        { label: 'Payload Size', value: '< 2 KB' },
        { label: 'Decoupling', value: 'Full Async' },
      ],
    },
    {
      id: 'heartbeat',
      title: '3. Worker Polling & Heartbeat',
      tagline: 'Visibility Timeout Extension Goroutine',
      icon: Activity,
      color: 'warning',
      description:
        'Background worker processes poll SQS using 20s long-polling. When a job begins, a concurrent heartbeat goroutine ticks every 20s calling ChangeMessageVisibility to prevent SQS from re-delivering the job during lengthy video processing.',
      codeSnippet: `// 3. Heartbeat goroutine keeps SQS visibility alive
go func() {
    ticker := time.NewTicker(20 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done(): return
        case <-ticker.C:
            sqsClient.ChangeMessageVisibility(ctx, msgReceipt, 60)
        }
    }
}()`,
      metrics: [
        { label: 'Long Poll Wait', value: '20s' },
        { label: 'Heartbeat Interval', value: '20s' },
        { label: 'Visibility Lease', value: '+60s' },
      ],
    },
    {
      id: 'ffmpeg',
      title: '4. Ephemeral Scratch & FFmpeg Normalization',
      tagline: '16kHz 16-bit Mono WAV Extraction',
      icon: Zap,
      color: 'success',
      description:
        'Worker provisions an isolated scratch directory under /tmp/transcription-jobs/{job_id}/. FFmpeg extracts audio from high-bitrate video and normalizes it to single-channel 16,000Hz PCM WAV, minimizing payload transfer to the speech engine.',
      codeSnippet: `// 4. Normalize audio with FFmpeg to 16kHz mono WAV
cmd := exec.CommandContext(ctx, "ffmpeg", "-y",
    "-i", videoFilePath,
    "-ar", "16000",
    "-ac", "1",
    "-c:a", "pcm_s16le",
    audioWavPath,
)`,
      metrics: [
        { label: 'Sample Rate', value: '16,000 Hz' },
        { label: 'Channels', value: '1 (Mono)' },
        { label: 'Scratch Disk', value: 'Auto-Cleaned' },
      ],
    },
    {
      id: 'transcribe',
      title: '5. AWS Transcribe / Mock Engine',
      tagline: 'Automated Speech-to-Text & Subtitle Splitting',
      icon: Sparkles,
      color: 'primary',
      description:
        'Normalized audio is uploaded to S3 and submitted to AWS Transcribe. The parser processes word-level JSON output, grouping words into timestamped subtitle segments with start/end millisecond precision and confidence scores.',
      codeSnippet: `// 5. Submit transcription job and parse word-level timestamps
output, err := transcribeProvider.TranscribeAudio(ctx, s3AudioURI)
segments := parser.ParseAWSOutput(output.RawJSON)`,
      metrics: [
        { label: 'Word Precision', value: 'Milliseconds' },
        { label: 'Confidence Score', value: '0.0 - 1.0' },
        { label: 'Zero-Cloud Mode', value: 'Mock Included' },
      ],
    },
    {
      id: 'commit',
      title: '6. PostgreSQL Transaction & Export',
      tagline: 'ACID Batch Segment Persistence & Subtitle Formats',
      icon: Database,
      color: 'secondary',
      description:
        'All subtitle segments and full transcripts are inserted within an atomic database transaction. Users can immediately view synchronized cues or export in standard SRT, WebVTT, and Plain Text formats.',
      codeSnippet: `// 6. Persist transcript and batch segments in single DB transaction
err := db.WithTx(ctx, func(tx pgx.Tx) error {
    transcriptRepo.CreateWithTx(tx, transcript)
    return segmentRepo.BatchInsertWithTx(tx, segments)
})`,
      metrics: [
        { label: 'Consistency', value: 'ACID WithTx' },
        { label: 'Export Formats', value: '.srt, .vtt, .txt' },
        { label: 'Cascade', value: 'ON DELETE CASCADE' },
      ],
    },
  ];

  const currentStep = pipelineSteps[activeStep];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-8">
      {/* Header Banner */}
      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-r from-primary-900/40 via-indigo-900/30 to-background border border-primary/20 p-8">
        <div className="max-w-3xl space-y-3">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-semibold">
            <Cpu className="w-3.5 h-3.5" />
            <span>Production System Design & Architecture</span>
          </div>
          <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
            Distributed Video Transcription Engine
          </h1>
          <p className="text-default-400 text-sm sm:text-base leading-relaxed">
            An enterprise-grade, asynchronous video processing system built with Go, AWS S3, SQS,
            FFmpeg, PostgreSQL, and HeroUI. Engineered for high concurrency, fault tolerance, and zero API memory bottlenecks.
          </p>
        </div>
      </div>

      {/* Navigation Tabs */}
      <Tabs
        aria-label="Architecture Sections"
        color="danger"
        variant="underlined"
        defaultSelectedKey="diagrams"
        classNames={{
          tabList: 'gap-6 w-full border-b border-default-200 dark:border-yt-border',
          cursor: 'w-full bg-yt-red',
          tab: 'max-w-fit px-0 h-12 text-sm font-medium',
        }}
      >
        {/* TAB 1: VISUAL ARCHITECTURE DIAGRAMS */}
        <Tab
          key="diagrams"
          title={
            <div className="flex items-center gap-2 text-yt-red font-semibold">
              <Network className="w-4 h-4 text-yt-red" />
              <span>Interactive Architecture Diagrams</span>
            </div>
          }
        >
          <div className="pt-6">
            <ArchitectureDiagrams />
          </div>
        </Tab>

        {/* TAB 2: INTERACTIVE SYSTEM FLOW */}
        <Tab
          key="flow"
          title={
            <div className="flex items-center gap-2">
              <Layers className="w-4 h-4" />
              <span>Pipeline Deep Dive</span>
            </div>
          }
        >
          <div className="pt-6 space-y-6">
            {/* Step Selector Pills */}
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-2">
              {pipelineSteps.map((step, idx) => {
                const Icon = step.icon;
                const isSelected = activeStep === idx;
                return (
                  <button
                    key={step.id}
                    onClick={() => setActiveStep(idx)}
                    className={`p-3 rounded-2xl border text-left transition-all flex flex-col gap-2 ${
                      isSelected
                        ? 'border-primary bg-primary/10 shadow-sm ring-2 ring-primary/20'
                        : 'border-default-200 dark:border-default-800 bg-default-50/40 hover:bg-default-100/50'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-default-400">0{idx + 1}</span>
                      <Icon className={`w-4 h-4 ${isSelected ? 'text-primary' : 'text-default-500'}`} />
                    </div>
                    <span className="text-xs font-semibold truncate text-foreground">
                      {step.title.split('. ')[1]}
                    </span>
                  </button>
                );
              })}
            </div>

            {/* Active Step Deep Dive Card */}
            <Card className="border border-default-200 dark:border-default-800 bg-background/80 shadow-lg">
              <CardHeader className="p-6 pb-2 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                  <div className="flex items-center gap-2.5 mb-1">
                    <span className="px-2.5 py-0.5 rounded-full bg-primary/20 text-primary text-xs font-bold font-mono">
                      Phase {activeStep + 1} of {pipelineSteps.length}
                    </span>
                    <h2 className="text-xl font-bold text-foreground">{currentStep.title}</h2>
                  </div>
                  <p className="text-xs text-primary font-medium">{currentStep.tagline}</p>
                </div>

                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant="flat"
                    isIconOnly
                    isDisabled={activeStep === 0}
                    onClick={() => setActiveStep((prev) => prev - 1)}
                    aria-label="Previous step"
                  >
                    <ArrowLeft className="w-4 h-4" />
                  </Button>
                  <Button
                    size="sm"
                    color="primary"
                    isDisabled={activeStep === pipelineSteps.length - 1}
                    onClick={() => setActiveStep((prev) => prev + 1)}
                    endContent={<ArrowRight className="w-4 h-4" />}
                  >
                    Next Phase
                  </Button>
                </div>
              </CardHeader>

              <CardBody className="p-6 space-y-6">
                <p className="text-sm text-default-600 leading-relaxed">{currentStep.description}</p>

                {/* Metrics Badges */}
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  {currentStep.metrics.map((metric, i) => (
                    <div
                      key={i}
                      className="p-3 rounded-xl bg-default-100/60 dark:bg-default-800/40 border border-default-200/60 dark:border-default-700/40"
                    >
                      <p className="text-[11px] text-default-400 font-medium uppercase tracking-wider">
                        {metric.label}
                      </p>
                      <p className="text-base font-bold text-foreground mt-0.5">{metric.value}</p>
                    </div>
                  ))}
                </div>

                {/* Code Snippet Box */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs font-semibold text-default-400 uppercase tracking-wider flex items-center gap-1.5">
                      <Code2 className="w-3.5 h-3.5 text-primary" /> Implementation Snapshot (Go)
                    </span>
                  </div>
                  <pre className="p-4 rounded-xl bg-default-900 text-default-100 text-xs font-mono overflow-x-auto border border-default-800 shadow-inner">
                    {currentStep.codeSnippet}
                  </pre>
                </div>
              </CardBody>
            </Card>
          </div>
        </Tab>

        {/* TAB 2: DISTRIBUTED COMPONENTS */}
        <Tab
          key="components"
          title={
            <div className="flex items-center gap-2">
              <Server className="w-4 h-4" />
              <span>Distributed Components</span>
            </div>
          }
        >
          <div className="pt-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <Card className="border border-default-200 dark:border-default-800 bg-background/50">
              <CardBody className="p-5 space-y-3">
                <div className="p-2.5 rounded-xl bg-primary-500/10 text-primary w-fit">
                  <Server className="w-6 h-6" />
                </div>
                <h3 className="font-bold text-base text-foreground">Go REST API Server</h3>
                <p className="text-xs text-default-400 leading-relaxed">
                  High-throughput HTTP server utilizing standard Go net/http + chi. Handles JWT authentication, S3 presigned URL issuance, and CRUD operations with sub-millisecond response latency.
                </p>
                <div className="pt-2 border-t border-default-100 dark:border-default-800 flex flex-wrap gap-1">
                  <Chip size="sm" variant="flat">net/http</Chip>
                  <Chip size="sm" variant="flat">JWT HS256</Chip>
                  <Chip size="sm" variant="flat">slog JSON</Chip>
                </div>
              </CardBody>
            </Card>

            <Card className="border border-default-200 dark:border-default-800 bg-background/50">
              <CardBody className="p-5 space-y-3">
                <div className="p-2.5 rounded-xl bg-secondary-500/10 text-secondary w-fit">
                  <MessageSquare className="w-6 h-6" />
                </div>
                <h3 className="font-bold text-base text-foreground">AWS SQS Message Broker</h3>
                <p className="text-xs text-default-400 leading-relaxed">
                  Decouples video upload traffic spikes from media transcoding. Offers 20-second long polling, message deduplication, and isolated queueing for resilient worker load leveling.
                </p>
                <div className="pt-2 border-t border-default-100 dark:border-default-800 flex flex-wrap gap-1">
                  <Chip size="sm" variant="flat">AWS SDK v2</Chip>
                  <Chip size="sm" variant="flat">Long Polling</Chip>
                  <Chip size="sm" variant="flat">Visibility Timeout</Chip>
                </div>
              </CardBody>
            </Card>

            <Card className="border border-default-200 dark:border-default-800 bg-background/50">
              <CardBody className="p-5 space-y-3">
                <div className="p-2.5 rounded-xl bg-warning-500/10 text-warning w-fit">
                  <Zap className="w-6 h-6" />
                </div>
                <h3 className="font-bold text-base text-foreground">Asynchronous Worker Daemon</h3>
                <p className="text-xs text-default-400 leading-relaxed">
                  Autonomous Go worker consuming SQS messages. Features visibility heartbeat renewal, ephemeral scratch directory isolation, and FFmpeg audio extraction.
                </p>
                <div className="pt-2 border-t border-default-100 dark:border-default-800 flex flex-wrap gap-1">
                  <Chip size="sm" variant="flat">FFmpeg 16kHz</Chip>
                  <Chip size="sm" variant="flat">Heartbeat Goroutine</Chip>
                  <Chip size="sm" variant="flat">Idempotent</Chip>
                </div>
              </CardBody>
            </Card>

            <Card className="border border-default-200 dark:border-default-800 bg-background/50">
              <CardBody className="p-5 space-y-3">
                <div className="p-2.5 rounded-xl bg-success-500/10 text-success w-fit">
                  <HardDrive className="w-6 h-6" />
                </div>
                <h3 className="font-bold text-base text-foreground">AWS S3 Object Storage</h3>
                <p className="text-xs text-default-400 leading-relaxed">
                  Stores raw user videos, extracted audio WAVs, and subtitle artifacts. Uploads occur client-direct via presigned URLs with 15-minute validity.
                </p>
                <div className="pt-2 border-t border-default-100 dark:border-default-800 flex flex-wrap gap-1">
                  <Chip size="sm" variant="flat">Direct Streaming</Chip>
                  <Chip size="sm" variant="flat">Presigned URLs</Chip>
                  <Chip size="sm" variant="flat">S3 Mock Fallback</Chip>
                </div>
              </CardBody>
            </Card>

            <Card className="border border-default-200 dark:border-default-800 bg-background/50">
              <CardBody className="p-5 space-y-3">
                <div className="p-2.5 rounded-xl bg-primary-500/10 text-primary w-fit">
                  <Database className="w-6 h-6" />
                </div>
                <h3 className="font-bold text-base text-foreground">PostgreSQL 16 Engine</h3>
                <p className="text-xs text-default-400 leading-relaxed">
                  Relational data persistence with foreign keys, indexes on video_id, and ON DELETE CASCADE. Managed via golang-migrate and thread-safe pgxpool.
                </p>
                <div className="pt-2 border-t border-default-100 dark:border-default-800 flex flex-wrap gap-1">
                  <Chip size="sm" variant="flat">pgxpool</Chip>
                  <Chip size="sm" variant="flat">WithTx Helpers</Chip>
                  <Chip size="sm" variant="flat">Embedded Migrations</Chip>
                </div>
              </CardBody>
            </Card>

            <Card className="border border-default-200 dark:border-default-800 bg-background/50">
              <CardBody className="p-5 space-y-3">
                <div className="p-2.5 rounded-xl bg-indigo-500/10 text-indigo-400 w-fit">
                  <Sparkles className="w-6 h-6" />
                </div>
                <h3 className="font-bold text-base text-foreground">Speech-to-Text Transcribe</h3>
                <p className="text-xs text-default-400 leading-relaxed">
                  Pluggable speech recognition engine supporting AWS Transcribe and zero-dependency local mock doubles for offline testing and continuous integration.
                </p>
                <div className="pt-2 border-t border-default-100 dark:border-default-800 flex flex-wrap gap-1">
                  <Chip size="sm" variant="flat">AWS Transcribe</Chip>
                  <Chip size="sm" variant="flat">Mock Provider</Chip>
                  <Chip size="sm" variant="flat">Confidence Scores</Chip>
                </div>
              </CardBody>
            </Card>
          </div>
        </Tab>

        {/* TAB 3: DATABASE SCHEMA */}
        <Tab
          key="schema"
          title={
            <div className="flex items-center gap-2">
              <Database className="w-4 h-4" />
              <span>Database Schema & Relations</span>
            </div>
          }
        >
          <div className="pt-6 space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {/* Users Table */}
              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardHeader className="bg-default-100/50 dark:bg-default-800/30 p-3 border-b border-default-200 dark:border-default-800 flex items-center justify-between">
                  <span className="font-mono text-xs font-bold text-foreground">users</span>
                  <Chip size="sm" variant="flat" color="primary">Primary Entity</Chip>
                </CardHeader>
                <CardBody className="p-3 font-mono text-xs space-y-1.5">
                  <div className="flex justify-between text-primary font-bold"><span>id</span><span>UUID PK</span></div>
                  <div className="flex justify-between text-default-500"><span>email</span><span>VARCHAR UNIQUE</span></div>
                  <div className="flex justify-between text-default-500"><span>password_hash</span><span>VARCHAR</span></div>
                  <div className="flex justify-between text-default-400"><span>created_at</span><span>TIMESTAMPTZ</span></div>
                </CardBody>
              </Card>

              {/* Videos Table */}
              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardHeader className="bg-default-100/50 dark:bg-default-800/30 p-3 border-b border-default-200 dark:border-default-800 flex items-center justify-between">
                  <span className="font-mono text-xs font-bold text-foreground">videos</span>
                  <Chip size="sm" variant="flat" color="secondary">Media Entity</Chip>
                </CardHeader>
                <CardBody className="p-3 font-mono text-xs space-y-1.5">
                  <div className="flex justify-between text-primary font-bold"><span>id</span><span>UUID PK</span></div>
                  <div className="flex justify-between text-secondary"><span>user_id</span><span>UUID FK -&gt; users</span></div>
                  <div className="flex justify-between text-default-500"><span>title</span><span>VARCHAR</span></div>
                  <div className="flex justify-between text-default-500"><span>s3_key</span><span>VARCHAR</span></div>
                  <div className="flex justify-between text-default-500"><span>status</span><span>VARCHAR</span></div>
                  <div className="flex justify-between text-default-400"><span>duration_seconds</span><span>FLOAT</span></div>
                </CardBody>
              </Card>

              {/* Jobs Table */}
              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardHeader className="bg-default-100/50 dark:bg-default-800/30 p-3 border-b border-default-200 dark:border-default-800 flex items-center justify-between">
                  <span className="font-mono text-xs font-bold text-foreground">jobs</span>
                  <Chip size="sm" variant="flat" color="warning">Queue Tracking</Chip>
                </CardHeader>
                <CardBody className="p-3 font-mono text-xs space-y-1.5">
                  <div className="flex justify-between text-primary font-bold"><span>id</span><span>UUID PK</span></div>
                  <div className="flex justify-between text-secondary"><span>video_id</span><span>UUID FK -&gt; videos</span></div>
                  <div className="flex justify-between text-default-500"><span>status</span><span>VARCHAR</span></div>
                  <div className="flex justify-between text-default-500"><span>retry_count</span><span>INT (max 3)</span></div>
                  <div className="flex justify-between text-default-400"><span>error_message</span><span>TEXT</span></div>
                </CardBody>
              </Card>

              {/* Transcripts Table */}
              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardHeader className="bg-default-100/50 dark:bg-default-800/30 p-3 border-b border-default-200 dark:border-default-800 flex items-center justify-between">
                  <span className="font-mono text-xs font-bold text-foreground">transcripts</span>
                  <Chip size="sm" variant="flat" color="success">Parent Document</Chip>
                </CardHeader>
                <CardBody className="p-3 font-mono text-xs space-y-1.5">
                  <div className="flex justify-between text-primary font-bold"><span>id</span><span>UUID PK</span></div>
                  <div className="flex justify-between text-secondary"><span>video_id</span><span>UUID FK -&gt; videos</span></div>
                  <div className="flex justify-between text-default-500"><span>language_code</span><span>VARCHAR</span></div>
                  <div className="flex justify-between text-default-500"><span>full_text</span><span>TEXT</span></div>
                  <div className="flex justify-between text-default-400"><span>provider</span><span>VARCHAR</span></div>
                </CardBody>
              </Card>

              {/* Transcript Segments Table */}
              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardHeader className="bg-default-100/50 dark:bg-default-800/30 p-3 border-b border-default-200 dark:border-default-800 flex items-center justify-between">
                  <span className="font-mono text-xs font-bold text-foreground">transcript_segments</span>
                  <Chip size="sm" variant="flat" color="danger">Timestamped Cues</Chip>
                </CardHeader>
                <CardBody className="p-3 font-mono text-xs space-y-1.5">
                  <div className="flex justify-between text-primary font-bold"><span>id</span><span>UUID PK</span></div>
                  <div className="flex justify-between text-secondary"><span>transcript_id</span><span>UUID FK -&gt; transcripts</span></div>
                  <div className="flex justify-between text-default-500"><span>start_time</span><span>FLOAT</span></div>
                  <div className="flex justify-between text-default-500"><span>end_time</span><span>FLOAT</span></div>
                  <div className="flex justify-between text-default-500"><span>text</span><span>TEXT</span></div>
                  <div className="flex justify-between text-default-400"><span>confidence</span><span>FLOAT</span></div>
                </CardBody>
              </Card>
            </div>
          </div>
        </Tab>

        {/* TAB 4: RESILIENCE PATTERNS */}
        <Tab
          key="patterns"
          title={
            <div className="flex items-center gap-2">
              <ShieldCheck className="w-4 h-4" />
              <span>Resilience Patterns & Deep Dives</span>
            </div>
          }
        >
          <div className="pt-6 space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardBody className="p-6 space-y-3">
                  <div className="flex items-center gap-2 text-primary font-bold text-base">
                    <Activity className="w-5 h-5" />
                    <h4>Visibility Heartbeat Goroutine</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Long video transcriptions can exceed standard SQS visibility timeouts (30s). A concurrent goroutine refreshes message visibility every 20 seconds, ensuring worker monopolization without duplicate work assignment.
                  </p>
                </CardBody>
              </Card>

              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardBody className="p-6 space-y-3">
                  <div className="flex items-center gap-2 text-secondary font-bold text-base">
                    <RefreshCw className="w-5 h-5" />
                    <h4>Idempotent Processing & Retry Limits</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Before processing, workers check if a job is already 'completed' or 'processing'. If retries exceed 3, the job is marked 'failed' to prevent poison pill loops from stalling queue consumers.
                  </p>
                </CardBody>
              </Card>

              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardBody className="p-6 space-y-3">
                  <div className="flex items-center gap-2 text-warning font-bold text-base">
                    <Zap className="w-5 h-5" />
                    <h4>Scratch Disk Ephemeral Cleanup</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Every media transcoding job receives a sandboxed /tmp/transcription-jobs/{'{job_id}'}/ directory. Go defer guarantees 100% disk reclamation regardless of processing success or panic recovery.
                  </p>
                </CardBody>
              </Card>

              <Card className="border border-default-200 dark:border-default-800 bg-background/50">
                <CardBody className="p-6 space-y-3">
                  <div className="flex items-center gap-2 text-success font-bold text-base">
                    <ShieldCheck className="w-5 h-5" />
                    <h4>Zero-Cloud Local Test Doubles</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Thread-safe in-memory implementations (MockStorage, MockQueue, MockTranscribeProvider) allow instantaneous local development and CI testing without requiring active AWS accounts or incurring cloud bills.
                  </p>
                </CardBody>
              </Card>
            </div>
          </div>
        </Tab>

        {/* TAB 5: ARCHITECTURAL TRADEOFFS */}
        <Tab
          key="tradeoffs"
          title={
            <div className="flex items-center gap-2">
              <GitBranch className="w-4 h-4" />
              <span>Architectural Tradeoffs</span>
            </div>
          }
        >
          <div className="pt-6 space-y-4">
            <Card className="border border-default-200 dark:border-default-800 bg-background/50 overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-default-100/50 dark:bg-default-800/30 text-default-500 uppercase tracking-wider font-semibold border-b border-default-200 dark:border-default-800">
                    <tr>
                      <th className="p-4">Design Choice</th>
                      <th className="p-4 text-primary">Selected Solution</th>
                      <th className="p-4">Alternative Considered</th>
                      <th className="p-4">Key Rationale</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-default-100 dark:divide-default-800">
                    <tr>
                      <td className="p-4 font-bold text-foreground">Media Uploads</td>
                      <td className="p-4 font-semibold text-primary">S3 Presigned Direct Upload</td>
                      <td className="p-4 text-default-400">API Gateway Multipart Proxy</td>
                      <td className="p-4 text-default-400">Eliminates Go server RAM saturation and socket timeouts on large video files.</td>
                    </tr>
                    <tr>
                      <td className="p-4 font-bold text-foreground">Queue Technology</td>
                      <td className="p-4 font-semibold text-primary">AWS SQS with Long Polling</td>
                      <td className="p-4 text-default-400">Apache Kafka / RabbitMQ</td>
                      <td className="p-4 text-default-400">Zero broker cluster management, automatic visibility extension, and serverless scaling.</td>
                    </tr>
                    <tr>
                      <td className="p-4 font-bold text-foreground">Worker Language</td>
                      <td className="p-4 font-semibold text-primary">Go (Goroutines + slog)</td>
                      <td className="p-4 text-default-400">Python / Celery</td>
                      <td className="p-4 text-default-400">Low-latency subprocess execution, single binary deployment, minimal CPU overhead.</td>
                    </tr>
                    <tr>
                      <td className="p-4 font-bold text-foreground">Database Transactions</td>
                      <td className="p-4 font-semibold text-primary">PostgreSQL WithTx Helper</td>
                      <td className="p-4 text-default-400">Non-relational Document DB</td>
                      <td className="p-4 text-default-400">ACID atomicity across transcript and hundreds of millisecond subtitle segments.</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </Card>
          </div>
        </Tab>
      </Tabs>
    </div>
  );
};
