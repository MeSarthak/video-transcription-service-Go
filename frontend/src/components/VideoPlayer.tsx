import React, { forwardRef, useImperativeHandle, useRef, useState, useEffect } from 'react';
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
  play: () => void;
  pause: () => void;
}

export const VideoPlayer = forwardRef<VideoPlayerRef, VideoPlayerProps>(
  ({ src, poster, segments = [], onTimeUpdate }, ref) => {
    const videoRef = useRef<HTMLVideoElement>(null);
    const [currentSegment, setCurrentSegment] = useState<TranscriptSegment | null>(null);

    useImperativeHandle(ref, () => ({
      seekTo: (timeSeconds: number) => {
        if (videoRef.current) {
          videoRef.current.currentTime = timeSeconds;
          videoRef.current.play().catch(() => {});
        }
      },
      play: () => {
        videoRef.current?.play().catch(() => {});
      },
      pause: () => {
        videoRef.current?.pause();
      },
    }));

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

    useEffect(() => {
      setCurrentSegment(null);
    }, [src]);

    return (
      <div className="relative rounded-2xl overflow-hidden bg-black/95 shadow-xl border border-default-200 dark:border-default-800 aspect-video flex items-center justify-center group">
        {src ? (
          <>
            <video
              ref={videoRef}
              src={src}
              poster={poster}
              controls
              onTimeUpdate={handleTimeUpdate}
              className="w-full h-full object-contain"
            />
            {/* Live Synchronized Subtitle Overlay */}
            {currentSegment && (
              <div className="absolute bottom-16 left-1/2 -translate-x-1/2 max-w-[85%] text-center pointer-events-none transition-all duration-150">
                <span className="inline-block px-4 py-1.5 rounded-lg bg-black/80 text-white font-medium text-sm sm:text-base backdrop-blur-md border border-white/10 shadow-lg tracking-wide">
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
