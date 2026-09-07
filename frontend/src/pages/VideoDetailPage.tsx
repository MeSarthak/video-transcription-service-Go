import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  Button,
  Spinner,
} from '@heroui/react';
import {
  ArrowLeft,
  RefreshCw,
  AlertCircle,
  Play,
  CheckCircle2,
  FileVideo,
  Copy,
  Check,
} from 'lucide-react';
import { api } from '../services/api';
import { Video, Job, Transcript, TranscriptSegment } from '../types';
import { VideoPlayer, VideoPlayerRef } from '../components/VideoPlayer';
import { TranscriptViewer } from '../components/TranscriptViewer';
import { ExportDropdown } from '../components/ExportDropdown';

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
  const [copiedKey, setCopiedKey] = useState(false);

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

  const videoRef = useRef(video);
  const jobRef = useRef(job);

  useEffect(() => {
    videoRef.current = video;
    jobRef.current = job;
  }, [video, job]);

  // Initial load
  useEffect(() => {
    loadData();
  }, [loadData]);

  // Auto-poll if video or job is processing
  useEffect(() => {
    const interval = setInterval(() => {
      const isProcessing =
        videoRef.current?.status === 'processing' ||
        videoRef.current?.status === 'uploading' ||
        jobRef.current?.status === 'processing' ||
        jobRef.current?.status === 'queued';

      if (isProcessing) {
        loadData();
      }
    }, 2500);

    return () => clearInterval(interval);
  }, [loadData]);

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

  const copyStorageKey = () => {
    if (video?.storage_key) {
      navigator.clipboard.writeText(video.storage_key);
      setCopiedKey(true);
      setTimeout(() => setCopiedKey(false), 2000);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[75vh] gap-3">
        <Spinner size="lg" color="danger" />
        <p className="text-sm text-default-400">Loading media player and transcript...</p>
      </div>
    );
  }

  if (error || !video) {
    return (
      <div className="max-w-xl mx-auto px-4 py-20 text-center space-y-4">
        <div className="w-16 h-16 rounded-full bg-danger-500/10 text-danger mx-auto flex items-center justify-center">
          <AlertCircle className="w-8 h-8" />
        </div>
        <h2 className="text-xl font-bold">Video Unavailable</h2>
        <p className="text-sm text-default-400">{error || 'Could not retrieve video details from storage.'}</p>
        <Button as={Link} to="/" radius="full" variant="flat" startContent={<ArrowLeft className="w-4 h-4" />}>
          Back to Videos
        </Button>
      </div>
    );
  }

  const isCompleted = video.status === 'completed';
  const isProcessing = video.status === 'processing' || video.status === 'uploading';

  return (
    <div className="max-w-[1650px] mx-auto px-3 sm:px-6 py-5 space-y-5">
      {/* YouTube Watch Layout: Left Player + Info, Right Transcript Drawer */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* LEFT COLUMN: Theater Player & Details (~65% width) */}
        <div className="lg:col-span-8 space-y-4">
          {/* Main Video Player */}
          <div className="rounded-2xl overflow-hidden bg-black shadow-2xl border border-default-200 dark:border-yt-border">
            <VideoPlayer
              ref={playerRef}
              src={video.playback_url}
              segments={segments}
              onTimeUpdate={setCurrentTime}
              currentTime={currentTime}
            />
          </div>

          {/* Video Title */}
          <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground leading-snug">
            {video.filename}
          </h1>

          {/* YouTube-style Channel & Action Row */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-3 border-b border-default-200 dark:border-yt-border">
            {/* Channel Info */}
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-yt-red/10 border border-yt-red/30 flex items-center justify-center">
                <FileVideo className="w-5 h-5 text-yt-red" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="font-bold text-sm text-foreground">TranscribeX Studio</h3>
                  <CheckCircle2 className="w-3.5 h-3.5 text-yt-red fill-yt-red text-white" />
                </div>
                <p className="text-xs text-default-400">Distributed Media Engine</p>
              </div>
            </div>

            {/* Pill Action Buttons */}
            <div className="flex items-center gap-2 flex-wrap">
              {/* Status Badge */}
              {isCompleted && (
                <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-bold bg-emerald-950/80 text-emerald-400 border border-emerald-500/30">
                  <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                  Transcribed
                </span>
              )}
              {isProcessing && (
                <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-bold bg-amber-950/80 text-amber-400 border border-amber-500/30">
                  <span className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
                  Transcribing...
                </span>
              )}

              {/* Start Transcription Button */}
              {(video.status === 'uploaded' || video.status === 'failed') && (
                <Button
                  size="sm"
                  radius="full"
                  startContent={<Play className="w-3.5 h-3.5 fill-white ml-0.5" />}
                  onClick={handleStartTranscription}
                  isLoading={actionLoading}
                  className="bg-yt-red hover:bg-red-700 text-white font-semibold text-xs shadow-md shadow-red-600/30 px-4"
                >
                  Transcribe
                </Button>
              )}

              {/* Export Subtitles Dropdown */}
              <ExportDropdown
                videoId={video.id}
                videoTitle={video.filename}
                disabled={segments.length === 0}
              />

              {/* Refresh */}
              <Button
                isIconOnly
                size="sm"
                radius="full"
                variant="flat"
                onClick={loadData}
                aria-label="Refresh video data"
                className="text-default-500 hover:text-foreground"
              >
                <RefreshCw className="w-4 h-4" />
              </Button>
            </div>
          </div>

          {/* YouTube Expandable Description Box */}
          <div className="p-4 rounded-2xl bg-default-100/70 dark:bg-yt-surface border border-default-200 dark:border-yt-border space-y-3">
            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs font-semibold text-foreground">
              {video.size_bytes > 0 && (
                <>
                  <span>{(video.size_bytes / (1024 * 1024)).toFixed(1)} MB</span>
                  <span>&bull;</span>
                </>
              )}
              <span>
                {video.duration_seconds && video.duration_seconds > 0
                  ? `${Math.floor(video.duration_seconds / 60)}m ${Math.floor(video.duration_seconds % 60)}s duration`
                  : 'Processing duration...'}
              </span>
              <span>&bull;</span>
              <span className="uppercase">{video.content_type || 'video/mp4'}</span>
              <span>&bull;</span>
              <span className="text-default-400">Language: {transcript?.language || job?.language || 'en-US'}</span>
            </div>

            {/* Storage Key Snippet with copy */}
            <div className="flex items-center justify-between p-2.5 rounded-xl bg-default-200/50 dark:bg-yt-card border border-default-200 dark:border-yt-border text-xs font-mono text-default-400">
              <span className="truncate pr-2">{video.storage_key}</span>
              <Button
                isIconOnly
                size="sm"
                variant="light"
                onClick={copyStorageKey}
                aria-label="Copy key"
                className="w-6 h-6 min-w-6 text-default-400 hover:text-foreground"
              >
                {copiedKey ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
              </Button>
            </div>

            {/* Pipeline Notice or Error Notice */}
            {job && job.error_message && (
              <div className="p-3 bg-danger-50 dark:bg-danger-900/20 text-danger border border-danger-200 dark:border-danger-800 rounded-xl text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                <span>Job notice: {job.error_message}</span>
              </div>
            )}
          </div>
        </div>

        {/* RIGHT COLUMN: Interactive YouTube Transcript Drawer (~35% width) */}
        <div className="lg:col-span-4 h-[680px] sticky top-20">
          <TranscriptViewer
            segments={segments}
            currentTime={currentTime}
            onSeek={handleSeek}
            fullText={transcript?.full_text}
          />
        </div>
      </div>
    </div>
  );
};
