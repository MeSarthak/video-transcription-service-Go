import React, { useState } from 'react';
import { Card, CardBody, Chip } from '@heroui/react';
import {
  Server,
  Database,
  HardDrive,
  MessageSquare,
  Zap,
  ArrowRight,
  CheckCircle2,
  AlertTriangle,
  FileVideo,
  Layers,
  Volume2,
} from 'lucide-react';

export const ArchitectureDiagrams: React.FC = () => {
  const [activeDiagram, setActiveDiagram] = useState<'pipeline' | 'upload' | 'worker' | 'schema'>('pipeline');
  const [selectedNode, setSelectedNode] = useState<string | null>('worker');

  return (
    <div className="space-y-6">
      {/* Diagram Mode Selector */}
      <div className="flex flex-wrap items-center gap-2 p-1.5 bg-default-100/80 dark:bg-yt-surface rounded-2xl border border-default-200 dark:border-yt-border w-fit">
        <button
          onClick={() => setActiveDiagram('pipeline')}
          className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all flex items-center gap-2 ${
            activeDiagram === 'pipeline'
              ? 'bg-yt-red text-white shadow-md shadow-red-600/30'
              : 'text-default-600 hover:text-foreground'
          }`}
        >
          <Layers className="w-3.5 h-3.5" />
          <span>End-to-End Pipeline</span>
        </button>

        <button
          onClick={() => setActiveDiagram('upload')}
          className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all flex items-center gap-2 ${
            activeDiagram === 'upload'
              ? 'bg-yt-red text-white shadow-md shadow-red-600/30'
              : 'text-default-600 hover:text-foreground'
          }`}
        >
          <HardDrive className="w-3.5 h-3.5" />
          <span>Zero-Memory Direct S3 Flow</span>
        </button>

        <button
          onClick={() => setActiveDiagram('worker')}
          className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all flex items-center gap-2 ${
            activeDiagram === 'worker'
              ? 'bg-yt-red text-white shadow-md shadow-red-600/30'
              : 'text-default-600 hover:text-foreground'
          }`}
        >
          <Zap className="w-3.5 h-3.5" />
          <span>Worker & FFmpeg State Machine</span>
        </button>

        <button
          onClick={() => setActiveDiagram('schema')}
          className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all flex items-center gap-2 ${
            activeDiagram === 'schema'
              ? 'bg-yt-red text-white shadow-md shadow-red-600/30'
              : 'text-default-600 hover:text-foreground'
          }`}
        >
          <Database className="w-3.5 h-3.5" />
          <span>PostgreSQL Relational Schema (ER)</span>
        </button>
      </div>

      {/* DIAGRAM 1: END-TO-END DISTRIBUTED PIPELINE */}
      {activeDiagram === 'pipeline' && (
        <Card className="border border-default-200 dark:border-yt-border bg-background dark:bg-yt-dark shadow-xl overflow-hidden">
          <CardBody className="p-6 space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-default-100 dark:border-yt-border pb-4">
              <div>
                <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                  <span className="w-2.5 h-2.5 rounded-full bg-yt-red animate-pulse" />
                  Distributed System Architecture Map
                </h3>
                <p className="text-xs text-default-400">
                  Asynchronous distributed media ingestion, normalization, and speech recognition pipeline.
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Chip size="sm" variant="flat" color="primary">
                  100% Non-Blocking
                </Chip>
                <Chip size="sm" variant="flat" color="success">
                  Active
                </Chip>
              </div>
            </div>

            {/* Visual SVG & Interactive Nodes Diagram */}
            <div className="relative overflow-x-auto py-8 px-2">
              <div className="min-w-[860px] grid grid-cols-5 gap-4 relative">
                {/* Connecting SVG Flow lines */}
                <svg className="absolute top-1/2 -translate-y-1/2 left-0 w-full h-12 pointer-events-none z-0">
                  <defs>
                    <linearGradient id="flowGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                      <stop offset="0%" stopColor="#FF0000" stopOpacity="0.8" />
                      <stop offset="50%" stopColor="#f59e0b" stopOpacity="0.8" />
                      <stop offset="100%" stopColor="#10b981" stopOpacity="0.8" />
                    </linearGradient>
                  </defs>
                  <line
                    x1="10%"
                    y1="50%"
                    x2="90%"
                    y2="50%"
                    stroke="url(#flowGrad)"
                    strokeWidth="3"
                    className="animate-flow-line"
                  />
                </svg>

                {/* Node 1: Client Web SPA */}
                <div
                  onClick={() => setSelectedNode('client')}
                  className={`relative z-10 p-4 rounded-2xl border transition-all cursor-pointer ${
                    selectedNode === 'client'
                      ? 'border-yt-red bg-yt-red/10 dark:bg-yt-red/15 ring-2 ring-yt-red/30 scale-105'
                      : 'border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface hover:border-default-400'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="p-2 rounded-xl bg-blue-500/10 text-blue-500">
                      <FileVideo className="w-5 h-5" />
                    </div>
                    <span className="text-[10px] font-mono text-default-400 font-bold">01</span>
                  </div>
                  <h4 className="text-sm font-bold text-foreground">React 19 Client</h4>
                  <p className="text-[11px] text-default-400 mt-1">
                    Direct S3 chunking, JWT auth & subtitle sync playback
                  </p>
                  <div className="mt-3 pt-2 border-t border-default-200 dark:border-yt-border flex items-center justify-between text-[10px] font-mono text-blue-400">
                    <span>SPA / HeroUI</span>
                    <span>HTTPS</span>
                  </div>
                </div>

                {/* Node 2: Go API Gateway */}
                <div
                  onClick={() => setSelectedNode('api')}
                  className={`relative z-10 p-4 rounded-2xl border transition-all cursor-pointer ${
                    selectedNode === 'api'
                      ? 'border-yt-red bg-yt-red/10 dark:bg-yt-red/15 ring-2 ring-yt-red/30 scale-105'
                      : 'border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface hover:border-default-400'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="p-2 rounded-xl bg-purple-500/10 text-purple-500">
                      <Server className="w-5 h-5" />
                    </div>
                    <span className="text-[10px] font-mono text-default-400 font-bold">02</span>
                  </div>
                  <h4 className="text-sm font-bold text-foreground">Go REST API</h4>
                  <p className="text-[11px] text-default-400 mt-1">
                    Issues presigned PUT URLs, handles CRUD, enqueues SQS
                  </p>
                  <div className="mt-3 pt-2 border-t border-default-200 dark:border-yt-border flex items-center justify-between text-[10px] font-mono text-purple-400">
                    <span>Gin Engine</span>
                    <span>&lt; 2ms latency</span>
                  </div>
                </div>

                {/* Node 3: AWS SQS & Storage */}
                <div
                  onClick={() => setSelectedNode('sqs')}
                  className={`relative z-10 p-4 rounded-2xl border transition-all cursor-pointer ${
                    selectedNode === 'sqs'
                      ? 'border-yt-red bg-yt-red/10 dark:bg-yt-red/15 ring-2 ring-yt-red/30 scale-105'
                      : 'border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface hover:border-default-400'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="p-2 rounded-xl bg-amber-500/10 text-amber-500">
                      <MessageSquare className="w-5 h-5" />
                    </div>
                    <span className="text-[10px] font-mono text-default-400 font-bold">03</span>
                  </div>
                  <h4 className="text-sm font-bold text-foreground">AWS S3 & SQS</h4>
                  <p className="text-[11px] text-default-400 mt-1">
                    Direct binary storage + decoupled asynchronous job queue
                  </p>
                  <div className="mt-3 pt-2 border-t border-default-200 dark:border-yt-border flex items-center justify-between text-[10px] font-mono text-amber-400">
                    <span>Zero Memory</span>
                    <span>20s Long Poll</span>
                  </div>
                </div>

                {/* Node 4: Go Worker Daemon */}
                <div
                  onClick={() => setSelectedNode('worker')}
                  className={`relative z-10 p-4 rounded-2xl border transition-all cursor-pointer ${
                    selectedNode === 'worker'
                      ? 'border-yt-red bg-yt-red/10 dark:bg-yt-red/15 ring-2 ring-yt-red/30 scale-105'
                      : 'border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface hover:border-default-400'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="p-2 rounded-xl bg-yt-red/10 text-yt-red">
                      <Zap className="w-5 h-5" />
                    </div>
                    <span className="text-[10px] font-mono text-default-400 font-bold">04</span>
                  </div>
                  <h4 className="text-sm font-bold text-foreground">Worker Daemon</h4>
                  <p className="text-[11px] text-default-400 mt-1">
                    FFmpeg 16kHz WAV extractor, heartbeat lease + fallback
                  </p>
                  <div className="mt-3 pt-2 border-t border-default-200 dark:border-yt-border flex items-center justify-between text-[10px] font-mono text-red-400">
                    <span>FFmpeg 6.1</span>
                    <span>Goroutines</span>
                  </div>
                </div>

                {/* Node 5: PostgreSQL 16 */}
                <div
                  onClick={() => setSelectedNode('postgres')}
                  className={`relative z-10 p-4 rounded-2xl border transition-all cursor-pointer ${
                    selectedNode === 'postgres'
                      ? 'border-yt-red bg-yt-red/10 dark:bg-yt-red/15 ring-2 ring-yt-red/30 scale-105'
                      : 'border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface hover:border-default-400'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-500">
                      <Database className="w-5 h-5" />
                    </div>
                    <span className="text-[10px] font-mono text-default-400 font-bold">05</span>
                  </div>
                  <h4 className="text-sm font-bold text-foreground">PostgreSQL 16</h4>
                  <p className="text-[11px] text-default-400 mt-1">
                    ACID batch segment inserts, foreign keys, cascade deletes
                  </p>
                  <div className="mt-3 pt-2 border-t border-default-200 dark:border-yt-border flex items-center justify-between text-[10px] font-mono text-emerald-400">
                    <span>pgxpool / SQL</span>
                    <span>SRT / VTT</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Deep Dive Inspector for Selected Node */}
            <div className="p-4 rounded-2xl bg-default-50/70 dark:bg-yt-surface border border-default-200 dark:border-yt-border">
              {selectedNode === 'client' && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <FileVideo className="w-4 h-4 text-blue-500" />
                    <h4 className="font-bold text-sm text-foreground">Client Tier: React 19 Single Page App</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Built with TypeScript, Vite, Tailwind CSS, and HeroUI. Implements direct binary upload via HTML5 File API.
                    Upon receiving a presigned URL, it streams the raw bytes via HTTP PUT directly to S3 storage.
                    Features synchronized audio/video cue highlighting and millisecond-accurate seek controls.
                  </p>
                </div>
              )}
              {selectedNode === 'api' && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Server className="w-4 h-4 text-purple-500" />
                    <h4 className="font-bold text-sm text-foreground">API Tier: Go REST Microservice</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Stateless Go HTTP server with Gin router. Issues AWS S3 presigned PUT URLs with 15-minute expiration,
                    verifies object presence via S3 HeadObject, and enqueues tasks into SQS. Zero video bytes enter
                    server memory, ensuring maximum concurrency and zero memory spikes.
                  </p>
                </div>
              )}
              {selectedNode === 'sqs' && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <MessageSquare className="w-4 h-4 text-amber-500" />
                    <h4 className="font-bold text-sm text-foreground">Queue Tier: AWS SQS Decoupled Buffer</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Decouples ingress web traffic from heavy media processing. Standard or FIFO queue with 20-second
                    long-polling. Features automatic fallback to local database polling when running offline or in zero-cloud development.
                  </p>
                </div>
              )}
              {selectedNode === 'worker' && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Zap className="w-4 h-4 text-yt-red" />
                    <h4 className="font-bold text-sm text-foreground">Worker Tier: Background Go Daemon & FFmpeg</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Spawns an isolated scratch directory per job. Probes duration with ffprobe, extracts 16kHz 16-bit mono
                    PCM WAV audio, and includes automatic fallback for silent videos using anullsrc filter. Maintains visibility
                    timeout extensions with concurrent heartbeat goroutines.
                  </p>
                </div>
              )}
              {selectedNode === 'postgres' && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Database className="w-4 h-4 text-emerald-500" />
                    <h4 className="font-bold text-sm text-foreground">Database Tier: PostgreSQL 16 Enterprise ACID</h4>
                  </div>
                  <p className="text-xs text-default-400 leading-relaxed">
                    Stores normalized metadata, users, video states, jobs, and subtitle segments. All subtitle segments are
                    committed inside an atomic WithTx transaction. Cascade constraints ensure clean data hygiene when videos are deleted.
                  </p>
                </div>
              )}
            </div>
          </CardBody>
        </Card>
      )}

      {/* DIAGRAM 2: DIRECT S3 VS TRADITIONAL API GATEWAY */}
      {activeDiagram === 'upload' && (
        <Card className="border border-default-200 dark:border-yt-border bg-background dark:bg-yt-dark shadow-xl">
          <CardBody className="p-6 space-y-6">
            <div className="border-b border-default-100 dark:border-yt-border pb-4">
              <h3 className="text-lg font-bold text-foreground">Direct-to-S3 vs Traditional API Gateway Flow</h3>
              <p className="text-xs text-default-400 mt-0.5">
                Why direct presigned uploads outperform conventional multipart API gateways in production.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Traditional Flawed Approach */}
              <div className="p-5 rounded-2xl border border-danger-200/50 dark:border-danger-900/30 bg-danger-500/5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold uppercase tracking-wider text-danger flex items-center gap-1.5">
                    <AlertTriangle className="w-4 h-4" /> Traditional Multipart Proxy (Bottleneck)
                  </span>
                  <Chip size="sm" color="danger" variant="flat">
                    High Risk
                  </Chip>
                </div>

                <div className="space-y-3 py-2">
                  <div className="p-3 rounded-xl bg-background dark:bg-yt-surface border border-default-200 dark:border-yt-border text-xs flex items-center justify-between">
                    <span>1. Client uploads 500MB video</span>
                    <ArrowRight className="w-3.5 h-3.5 text-default-400" />
                  </div>
                  <div className="p-3 rounded-xl bg-danger-500/10 border border-danger-500/30 text-xs text-danger font-semibold flex items-center justify-between">
                    <span>2. API buffers 500MB in RAM</span>
                    <span className="text-[10px] font-mono">Memory Spike!</span>
                  </div>
                  <div className="p-3 rounded-xl bg-background dark:bg-yt-surface border border-default-200 dark:border-yt-border text-xs flex items-center justify-between">
                    <span>3. API re-uploads bytes to S3</span>
                    <ArrowRight className="w-3.5 h-3.5 text-default-400" />
                  </div>
                </div>

                <div className="p-3 rounded-xl bg-danger-50 dark:bg-danger-900/20 text-danger text-[11px] space-y-1">
                  <p className="font-bold">Failure Modes:</p>
                  <p>&bull; 10 concurrent uploads consume 5GB+ server RAM.</p>
                  <p>&bull; API gateway OOM kills during video spikes.</p>
                  <p>&bull; Double bandwidth consumption and double latency.</p>
                </div>
              </div>

              {/* TranscribeX Direct Presigned Architecture */}
              <div className="p-5 rounded-2xl border border-success-200/50 dark:border-success-900/30 bg-success-500/5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold uppercase tracking-wider text-success flex items-center gap-1.5">
                    <CheckCircle2 className="w-4 h-4" /> TranscribeX Direct S3 Architecture
                  </span>
                  <Chip size="sm" color="success" variant="flat">
                    Recommended
                  </Chip>
                </div>

                <div className="space-y-3 py-2">
                  <div className="p-3 rounded-xl bg-background dark:bg-yt-surface border border-default-200 dark:border-yt-border text-xs flex items-center justify-between">
                    <span>1. Client requests URL (&lt;1KB JSON)</span>
                    <ArrowRight className="w-3.5 h-3.5 text-default-400" />
                  </div>
                  <div className="p-3 rounded-xl bg-success-500/10 border border-success-500/30 text-xs text-success font-semibold flex items-center justify-between">
                    <span>2. Client streams directly to S3</span>
                    <span className="text-[10px] font-mono">0% API RAM</span>
                  </div>
                  <div className="p-3 rounded-xl bg-background dark:bg-yt-surface border border-default-200 dark:border-yt-border text-xs flex items-center justify-between">
                    <span>3. Client confirms with S3 HeadObject</span>
                    <CheckCircle2 className="w-3.5 h-3.5 text-success" />
                  </div>
                </div>

                <div className="p-3 rounded-xl bg-success-50 dark:bg-success-900/20 text-success text-[11px] space-y-1">
                  <p className="font-bold">Production Advantages:</p>
                  <p>&bull; API memory footprint remains steady under 50MB.</p>
                  <p>&bull; S3 handles massive terabyte-scale uploads natively.</p>
                  <p>&bull; Presigned URLs expire automatically after 15 minutes.</p>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>
      )}

      {/* DIAGRAM 3: WORKER STATE MACHINE & FFmpeg EXTRACTION */}
      {activeDiagram === 'worker' && (
        <Card className="border border-default-200 dark:border-yt-border bg-background dark:bg-yt-dark shadow-xl">
          <CardBody className="p-6 space-y-6">
            <div className="border-b border-default-100 dark:border-yt-border pb-4">
              <h3 className="text-lg font-bold text-foreground">Worker Daemon & FFmpeg State Machine</h3>
              <p className="text-xs text-default-400 mt-0.5">
                Lifecycle of an asynchronous transcription job from SQS receipt to database transaction.
              </p>
            </div>

            <div className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                <div className="p-3.5 rounded-2xl bg-default-50 dark:bg-yt-surface border border-default-200 dark:border-yt-border">
                  <div className="flex items-center gap-2 mb-1.5 text-xs font-bold text-foreground">
                    <span className="w-5 h-5 rounded-full bg-primary/20 text-primary flex items-center justify-center text-[10px] font-mono">1</span>
                    ReceiveMessage
                  </div>
                  <p className="text-[11px] text-default-400">
                    20s long-polling from SQS. Falls back to DB queue if cloud credentials not present.
                  </p>
                </div>

                <div className="p-3.5 rounded-2xl bg-default-50 dark:bg-yt-surface border border-default-200 dark:border-yt-border">
                  <div className="flex items-center gap-2 mb-1.5 text-xs font-bold text-foreground">
                    <span className="w-5 h-5 rounded-full bg-warning/20 text-warning flex items-center justify-center text-[10px] font-mono">2</span>
                    Heartbeat Goroutine
                  </div>
                  <p className="text-[11px] text-default-400">
                    Ticks every 20s calling ChangeMessageVisibility(+60s) to keep worker lease alive.
                  </p>
                </div>

                <div className="p-3.5 rounded-2xl bg-default-50 dark:bg-yt-surface border border-default-200 dark:border-yt-border">
                  <div className="flex items-center gap-2 mb-1.5 text-xs font-bold text-foreground">
                    <span className="w-5 h-5 rounded-full bg-yt-red/20 text-yt-red flex items-center justify-center text-[10px] font-mono">3</span>
                    FFmpeg Normalization
                  </div>
                  <p className="text-[11px] text-default-400">
                    Extracts 16kHz 16-bit mono WAV. Detects silent videos and generates anullsrc fallback.
                  </p>
                </div>

                <div className="p-3.5 rounded-2xl bg-default-50 dark:bg-yt-surface border border-default-200 dark:border-yt-border">
                  <div className="flex items-center gap-2 mb-1.5 text-xs font-bold text-foreground">
                    <span className="w-5 h-5 rounded-full bg-success/20 text-success flex items-center justify-center text-[10px] font-mono">4</span>
                    ACID Commit (WithTx)
                  </div>
                  <p className="text-[11px] text-default-400">
                    Parses word-level millisecond timestamps and batch inserts subtitle segments atomically.
                  </p>
                </div>
              </div>

              {/* Audio Extraction Branch Highlight */}
              <div className="p-4 rounded-2xl bg-yt-red/5 border border-yt-red/20 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div className="space-y-1">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-yt-red flex items-center gap-1.5">
                    <Volume2 className="w-4 h-4" /> Resilient Audio Detection & Silent Media Handling
                  </h4>
                  <p className="text-xs text-default-400">
                    Videos without audio streams (such as stock footage or screen animations) are detected gracefully.
                    FFmpeg synthesizes a matching silent PCM track, preventing worker exit code 234 crashes.
                  </p>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <Chip size="sm" variant="flat" color="primary">
                    16,000 Hz Mono
                  </Chip>
                  <Chip size="sm" variant="flat" color="success">
                    Zero-Crash
                  </Chip>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>
      )}

      {/* DIAGRAM 4: DATABASE RELATIONAL SCHEMA (ER) */}
      {activeDiagram === 'schema' && (
        <Card className="border border-default-200 dark:border-yt-border bg-background dark:bg-yt-dark shadow-xl">
          <CardBody className="p-6 space-y-6">
            <div className="border-b border-default-100 dark:border-yt-border pb-4">
              <h3 className="text-lg font-bold text-foreground">PostgreSQL 16 Relational Schema (ER Diagram)</h3>
              <p className="text-xs text-default-400 mt-0.5">
                Normalized data model with cascading foreign key constraints and transactional consistency.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {/* Table: videos */}
              <div className="p-4 rounded-2xl border border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface space-y-3">
                <div className="flex items-center justify-between border-b border-default-200 dark:border-yt-border pb-2">
                  <span className="font-mono text-xs font-bold text-primary flex items-center gap-1.5">
                    <FileVideo className="w-3.5 h-3.5" /> videos
                  </span>
                  <span className="text-[10px] text-default-400 font-mono">TABLE</span>
                </div>
                <div className="space-y-1.5 text-xs font-mono">
                  <div className="flex justify-between text-foreground">
                    <span className="text-yt-red font-bold">id</span>
                    <span className="text-default-400">UUID (PK)</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>user_id</span>
                    <span className="text-default-400">UUID (FK)</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>filename</span>
                    <span className="text-default-400">VARCHAR</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>storage_key</span>
                    <span className="text-default-400">TEXT</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>status</span>
                    <span className="text-default-400">video_status</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>duration_secs</span>
                    <span className="text-default-400">NUMERIC</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>size_bytes</span>
                    <span className="text-default-400">BIGINT</span>
                  </div>
                </div>
              </div>

              {/* Table: transcription_jobs */}
              <div className="p-4 rounded-2xl border border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface space-y-3">
                <div className="flex items-center justify-between border-b border-default-200 dark:border-yt-border pb-2">
                  <span className="font-mono text-xs font-bold text-amber-500 flex items-center gap-1.5">
                    <Zap className="w-3.5 h-3.5" /> transcription_jobs
                  </span>
                  <span className="text-[10px] text-default-400 font-mono">TABLE</span>
                </div>
                <div className="space-y-1.5 text-xs font-mono">
                  <div className="flex justify-between text-foreground">
                    <span className="text-amber-500 font-bold">id</span>
                    <span className="text-default-400">UUID (PK)</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>video_id</span>
                    <span className="text-default-400">UUID (FK)</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>user_id</span>
                    <span className="text-default-400">UUID (FK)</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>status</span>
                    <span className="text-default-400">job_status</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>attempts</span>
                    <span className="text-default-400">INT</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>error_message</span>
                    <span className="text-default-400">TEXT</span>
                  </div>
                </div>
              </div>

              {/* Table: transcripts & segments */}
              <div className="p-4 rounded-2xl border border-default-200 dark:border-yt-border bg-default-50 dark:bg-yt-surface space-y-3">
                <div className="flex items-center justify-between border-b border-default-200 dark:border-yt-border pb-2">
                  <span className="font-mono text-xs font-bold text-emerald-500 flex items-center gap-1.5">
                    <Layers className="w-3.5 h-3.5" /> transcript_segments
                  </span>
                  <span className="text-[10px] text-default-400 font-mono">TABLE</span>
                </div>
                <div className="space-y-1.5 text-xs font-mono">
                  <div className="flex justify-between text-foreground">
                    <span className="text-emerald-500 font-bold">id</span>
                    <span className="text-default-400">UUID (PK)</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>transcript_id</span>
                    <span className="text-default-400">UUID (FK)</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>sequence_number</span>
                    <span className="text-default-400">INT</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>start_time</span>
                    <span className="text-default-400">FLOAT8</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>end_time</span>
                    <span className="text-default-400">FLOAT8</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>text</span>
                    <span className="text-default-400">TEXT</span>
                  </div>
                  <div className="flex justify-between text-foreground">
                    <span>confidence</span>
                    <span className="text-default-400">FLOAT4</span>
                  </div>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>
      )}
    </div>
  );
};
