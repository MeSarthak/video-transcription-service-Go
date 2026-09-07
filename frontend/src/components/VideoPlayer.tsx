import React, { forwardRef, useImperativeHandle, useRef, useState, useEffect, useCallback } from 'react';
import { RotateCcw, RotateCw } from 'lucide-react';
import { TranscriptSegment } from '../types';

interface VideoPlayerProps {
  src?: string;
  poster?: string;
  segments?: TranscriptSegment[];
  onTimeUpdate?: (currentTime: number) => void;
  currentTime?: number;
}

export interface VideoPlayerRef {
  seekTo: (timeSeconds: number) => void;
  skipBy: (deltaSeconds: number) => void;
  play: () => void;
  pause: () => void;
}

export const VideoPlayer = forwardRef<VideoPlayerRef, VideoPlayerProps>(
  ({ src, poster, segments = [], onTimeUpdate }, ref) => {
    const videoRef = useRef<HTMLVideoElement>(null);
    const [currentSegment, setCurrentSegment] = useState<TranscriptSegment | null>(null);
    const [seekFeedback, setSeekFeedback] = useState<string | null>(null);
    const feedbackTimeoutRef = useRef<number | null>(null);

    const showFeedback = useCallback((text: string) => {
      if (feedbackTimeoutRef.current !== null) {
        window.clearTimeout(feedbackTimeoutRef.current);
      }
      setSeekFeedback(text);
      feedbackTimeoutRef.current = window.setTimeout(() => {
        setSeekFeedback(null);
      }, 750);
    }, []);

    const seekToTime = useCallback((timeSeconds: number) => {
      if (!videoRef.current) return;
      const target = Math.max(0, Math.min(timeSeconds, videoRef.current.duration || Infinity));
      videoRef.current.currentTime = target;
      videoRef.current.play().catch(() => {});
    }, []);

    const skipByDelta = useCallback((deltaSeconds: number) => {
      if (!videoRef.current) return;
      const newTime = Math.max(0, (videoRef.current.currentTime || 0) + deltaSeconds);
      videoRef.current.currentTime = newTime;
      showFeedback(deltaSeconds > 0 ? `+${deltaSeconds}s` : `${deltaSeconds}s`);
      videoRef.current.play().catch(() => {});
    }, [showFeedback]);

    useImperativeHandle(ref, () => ({
      seekTo: seekToTime,
      skipBy: skipByDelta,
      play: () => {
        videoRef.current?.play().catch(() => {});
      },
      pause: () => {
        videoRef.current?.pause();
      },
    }), [seekToTime, skipByDelta]);

    const handleTimeUpdate = () => {
      if (!videoRef.current) return;
      const time = videoRef.current.currentTime;
      if (onTimeUpdate) {
        onTimeUpdate(time);
      }

      // Find active segment
      const active = segments.find(
        (s) => time >= s.start_time && time <= s.end_time
      );
      setCurrentSegment(active || null);
    };

    // Keyboard shortcuts for YouTube-style playback
    useEffect(() => {
      const handleKeyDown = (e: KeyboardEvent) => {
        // Ignore if user is typing in an input or textarea
        if (['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement)?.tagName)) {
          return;
        }

        if (e.key === 'ArrowLeft') {
          e.preventDefault();
          skipByDelta(-5);
        } else if (e.key === 'ArrowRight') {
          e.preventDefault();
          skipByDelta(5);
        } else if (e.key === 'j' || e.key === 'J') {
          e.preventDefault();
          skipByDelta(-10);
        } else if (e.key === 'l' || e.key === 'L') {
          e.preventDefault();
          skipByDelta(10);
        } else if (e.key === 'k' || e.key === 'K') {
          e.preventDefault();
          if (videoRef.current) {
            if (videoRef.current.paused) {
              videoRef.current.play().catch(() => {});
            } else {
              videoRef.current.pause();
            }
          }
        }
      };

      window.addEventListener('keydown', handleKeyDown);
      return () => window.removeEventListener('keydown', handleKeyDown);
    }, [skipByDelta]);

    useEffect(() => {
      setCurrentSegment(null);
    }, [src]);

    return (
      <div className="relative rounded-2xl overflow-hidden bg-black shadow-xl border border-default-200 dark:border-default-800 aspect-video flex items-center justify-center group select-none">
        {src ? (
          <>
            <video
              ref={videoRef}
              src={src}
              poster={poster}
              controls
              preload="metadata"
              playsInline
              onTimeUpdate={handleTimeUpdate}
              className="w-full h-full object-contain"
            />

            {/* Quick Skip Buttons Overlay on Hover */}
            <div className="absolute top-4 right-4 flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-auto z-10">
              <button
                type="button"
                onClick={() => skipByDelta(-10)}
                title="Rewind 10s (J / ←)"
                className="flex items-center gap-1 px-2.5 py-1.5 rounded-full bg-black/70 hover:bg-black/90 text-white text-xs font-semibold backdrop-blur-md border border-white/20 shadow-md transition-all active:scale-95"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                <span>10s</span>
              </button>
              <button
                type="button"
                onClick={() => skipByDelta(10)}
                title="Forward 10s (L / →)"
                className="flex items-center gap-1 px-2.5 py-1.5 rounded-full bg-black/70 hover:bg-black/90 text-white text-xs font-semibold backdrop-blur-md border border-white/20 shadow-md transition-all active:scale-95"
              >
                <span>10s</span>
                <RotateCw className="w-3.5 h-3.5" />
              </button>
            </div>

            {/* Visual Skip Indicator Pill */}
            {seekFeedback && (
              <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 pointer-events-none transition-all animate-in fade-in zoom-in-95 duration-150 z-20">
                <div className="px-5 py-2.5 rounded-2xl bg-black/80 text-white font-bold text-lg backdrop-blur-lg border border-white/20 shadow-2xl tracking-wider">
                  {seekFeedback}
                </div>
              </div>
            )}

            {/* Live Synchronized Subtitle Overlay */}
            {currentSegment && (
              <div className="absolute bottom-16 left-1/2 -translate-x-1/2 max-w-[85%] text-center pointer-events-none transition-all duration-150 z-10">
                <span className="inline-block px-4 py-1.5 rounded-lg bg-black/85 text-white font-medium text-sm sm:text-base backdrop-blur-md border border-white/10 shadow-lg tracking-wide">
                  {currentSegment.text}
                </span>
              </div>
            )}
          </>
        ) : (
          <div className="text-center p-8 text-default-400">
            <p className="text-sm">No video source available or media still processing</p>
          </div>
        )}
      </div>
    );
  }
);

VideoPlayer.displayName = 'VideoPlayer';
