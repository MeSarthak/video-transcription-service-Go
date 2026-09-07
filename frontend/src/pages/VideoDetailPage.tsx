import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  Button,
  Card,
  CardBody,
  Spinner,
  Accordion,
  AccordionItem,
  Chip,
} from '@heroui/react';
import {
  ArrowLeft,
  RefreshCw,
  Clock,
  HardDrive,
  FileCode,
  Sparkles,
  AlertCircle,
  PlayCircle,
} from 'lucide-react';
import { api } from '../services/api';
import { Video, Job, Transcript, TranscriptSegment } from '../types';
import { VideoPlayer, VideoPlayerRef } from '../components/VideoPlayer';
import { TranscriptViewer } from '../components/TranscriptViewer';
import { ExportDropdown } from '../components/ExportDropdown';
import { JobStatusBadge } from '../components/JobStatusBadge';

export const VideoDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [video, setVideo] = useState<Video | null>(null);
  const [job, setJob] = useState<Job | null>(null);
  const [transcript, setTranscript] = useState<Transcript | null>(null);
  const [segments, setSegments] = useState<TranscriptSegment[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const playerRef = useRef<VideoPlayerRef>(null);

  const loadData = useCallback(async () => {
    if (!id) return;
    try {
      setError(null);
      const videoData = await api.getVideo(id);
      setVideo(videoData);

      // Check job status
      try {
        const jobData = await api.getJobStatus(id);
        setJob(jobData);
      } catch {
        // Job might not exist yet
      }

      // Check transcript if ready
      if (videoData.status === 'completed') {
        try {
          const transData = await api.getTranscription(id);
          setTranscript(transData);
          setSegments(transData.segments || []);
        } catch {
          // Transcript not yet ready
        }
      }
    } catch (err: unknown) {
      setError((err as Error).message || 'Failed to load video details');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadData();

    // Auto-poll if video or job is processing
    const interval = setInterval(() => {
      if (
        video?.status === 'processing' ||
        video?.status === 'uploading' ||
        job?.status === 'processing' ||
        job?.status === 'queued'
      ) {
        loadData();
      }
    }, 2500);

    return () => clearInterval(interval);
  }, [loadData, video?.status, job?.status]);

  const handleStartTranscription = async () => {
    if (!id) return;
    setActionLoading(true);
    try {
      const jobData = await api.startTranscription(id);
      setJob(jobData);
      if (video) {
        setVideo({ ...video, status: 'processing' });
      }
    } catch (err: unknown) {
      setError((err as Error).message || 'Failed to trigger transcription');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSeek = (time: number) => {
    setCurrentTime(time);
    playerRef.current?.seekTo(time);
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[70vh] gap-3">
        <Spinner size="lg" color="primary" />
        <p className="text-sm text-default-400">Loading video and transcription data...</p>
      </div>
    );
  }

  if (error || !video) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center space-y-4">
        <div className="p-4 rounded-full bg-danger-50 dark:bg-danger-900/20 text-danger w-16 h-16 mx-auto flex items-center justify-center">
          <AlertCircle className="w-8 h-8" />
        </div>
        <h2 className="text-xl font-bold">Video Not Found</h2>
        <p className="text-sm text-default-400">{error || 'Could not retrieve video information.'}</p>
        <Button as={Link} to="/" color="primary" variant="flat" startContent={<ArrowLeft className="w-4 h-4" />}>
          Back to Dashboard
        </Button>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-6 space-y-6">
      {/* Top Navigation & Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-default-200 dark:border-default-800">
        <div className="flex items-center gap-3">
          <Button
            as={Link}
            to="/"
            size="sm"
            variant="light"
            isIconOnly
            className="text-default-500"
            aria-label="Back"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground truncate max-w-lg">
                {video.filename}
              </h1>
              <JobStatusBadge status={video.status} />
            </div>
            <p className="text-xs text-default-400 mt-0.5 font-mono">{video.storage_key}</p>
          </div>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <Button
            size="sm"
            variant="flat"
            onClick={loadData}
            startContent={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Refresh
          </Button>

          {(video.status === 'uploaded' || video.status === 'failed') && (
            <Button
              size="sm"
              color="primary"
              onClick={handleStartTranscription}
              isLoading={actionLoading}
              startContent={<PlayCircle className="w-4 h-4" />}
            >
              Start Transcription
            </Button>
          )}

          <ExportDropdown
            videoId={video.id}
            videoTitle={video.filename}
            disabled={segments.length === 0}
          />
        </div>
      </div>

      {/* Main Content: Video + Transcript Split */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Player & Metadata */}
        <div className="lg:col-span-7 space-y-4">
          <VideoPlayer
            ref={playerRef}
            src={video.playback_url}
            segments={segments}
            onTimeUpdate={setCurrentTime}
            currentTime={currentTime}
          />

          {/* Video Metadata Card */}
          <Card className="border border-default-200 dark:border-default-800 bg-background/50">
            <CardBody className="p-4 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold uppercase tracking-wider text-default-400">
                  Media Properties
                </span>
                {job?.provider && (
                  <Chip size="sm" variant="flat" color="secondary" className="text-xs font-medium">
                    Engine: {job.provider}
                  </Chip>
                )}
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2 text-xs">
                <div className="p-2.5 rounded-xl bg-default-50 dark:bg-default-100/30">
                  <span className="text-default-400 flex items-center gap-1 mb-1">
                    <Clock className="w-3 h-3" /> Duration
                  </span>
                  <span className="font-semibold text-foreground">
                    {video.duration_seconds && video.duration_seconds > 0
                      ? `${video.duration_seconds.toFixed(1)}s`
                      : 'Calculated in worker'}
                  </span>
                </div>

                <div className="p-2.5 rounded-xl bg-default-50 dark:bg-default-100/30">
                  <span className="text-default-400 flex items-center gap-1 mb-1">
                    <HardDrive className="w-3 h-3" /> File Size
                  </span>
                  <span className="font-semibold text-foreground">
                    {(video.size_bytes / (1024 * 1024)).toFixed(2)} MB
                  </span>
                </div>

                <div className="p-2.5 rounded-xl bg-default-50 dark:bg-default-100/30">
                  <span className="text-default-400 flex items-center gap-1 mb-1">
                    <Sparkles className="w-3 h-3" /> Language
                  </span>
                  <span className="font-semibold text-foreground">
                    {transcript?.language || job?.language || 'en-US'}
                  </span>
                </div>

                <div className="p-2.5 rounded-xl bg-default-50 dark:bg-default-100/30">
                  <span className="text-default-400 flex items-center gap-1 mb-1">
                    <FileCode className="w-3 h-3" /> Content-Type
                  </span>
                  <span className="font-semibold text-foreground truncate block">
                    {video.content_type || 'video/mp4'}
                  </span>
                </div>
              </div>

              {job && job.error_message && (
                <div className="p-3 bg-danger-50 dark:bg-danger-900/20 text-danger border border-danger-200 dark:border-danger-800 rounded-xl text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  <span>Job Error: {job.error_message}</span>
                </div>
              )}
            </CardBody>
          </Card>
        </div>

        {/* Right Column: Interactive Subtitles */}
        <div className="lg:col-span-5 h-[580px]">
          <TranscriptViewer
            segments={segments}
            currentTime={currentTime}
            onSeek={handleSeek}
            fullText={transcript?.full_text}
          />
        </div>
      </div>

      {/* Raw JSON Debug Inspector */}
      {segments.length > 0 && (
        <div className="pt-4">
          <Accordion variant="bordered">
            <AccordionItem
              key="json-inspector"
              aria-label="Raw Transcription Model Inspector"
              title={
                <span className="text-xs font-semibold text-default-500 uppercase tracking-wider flex items-center gap-2">
                  <FileCode className="w-4 h-4" /> Raw Output & JSON Segments Model
                </span>
              }
            >
              <pre className="p-4 rounded-xl bg-default-900 text-default-100 text-xs overflow-x-auto max-h-72 font-mono">
                {JSON.stringify({ video, job, transcript }, null, 2)}
              </pre>
            </AccordionItem>
          </Accordion>
        </div>
      )}
    </div>
  );
};
