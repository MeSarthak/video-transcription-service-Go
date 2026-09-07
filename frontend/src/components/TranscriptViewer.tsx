import React, { useState, useEffect, useRef } from 'react';
import {
  Input,
  Button,
  Card,
  CardBody,
  Chip,
  Tooltip,
} from '@heroui/react';
import { Search, Copy, Check, Play, Clock, Sparkles } from 'lucide-react';
import { TranscriptSegment } from '../types';

interface TranscriptViewerProps {
  segments: TranscriptSegment[];
  currentTime: number;
  onSeek: (time: number) => void;
  fullText?: string;
}

export const TranscriptViewer: React.FC<TranscriptViewerProps> = ({
  segments,
  currentTime,
  onSeek,
  fullText,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [copiedFull, setCopiedFull] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const activeSegmentRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const filteredSegments = segments.filter((seg) =>
    seg.text.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCopyFull = () => {
    const textToCopy = fullText || segments.map((s) => s.text).join(' ');
    navigator.clipboard.writeText(textToCopy);
    setCopiedFull(true);
    setTimeout(() => setCopiedFull(false), 2000);
  };

  useEffect(() => {
    if (autoScroll && activeSegmentRef.current && containerRef.current) {
      activeSegmentRef.current.scrollIntoView({
        behavior: 'smooth',
        block: 'nearest',
      });
    }
  }, [currentTime, autoScroll]);

  return (
    <Card className="h-full flex flex-col border border-default-200 dark:border-yt-border bg-background dark:bg-yt-dark shadow-lg rounded-2xl overflow-hidden">
      {/* YouTube-style Transcript Drawer Header */}
      <div className="p-4 border-b border-default-200 dark:border-yt-border flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <h3 className="text-base font-bold text-foreground">Transcript</h3>
          <span className="text-[11px] text-default-400 font-medium">
            ({segments.length} cues)
          </span>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="flat"
            radius="full"
            onClick={() => setAutoScroll(!autoScroll)}
            className={`text-xs font-semibold ${
              autoScroll ? 'bg-yt-red/10 text-yt-red' : 'text-default-400'
            }`}
          >
            {autoScroll ? 'Auto-scroll On' : 'Auto-scroll Off'}
          </Button>

          <Button
            size="sm"
            variant="flat"
            radius="full"
            startContent={copiedFull ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
            onClick={handleCopyFull}
            className="text-xs font-medium"
          >
            {copiedFull ? 'Copied' : 'Copy'}
          </Button>
        </div>
      </div>

      {/* Search Input */}
      <div className="p-3 border-b border-default-100 dark:border-yt-border/80 bg-default-50/50 dark:bg-yt-surface">
        <Input
          size="sm"
          radius="full"
          placeholder="Search in transcript..."
          value={searchQuery}
          onValueChange={setSearchQuery}
          startContent={<Search className="w-3.5 h-3.5 text-default-400 ml-1" />}
          isClearable
          classNames={{
            inputWrapper:
              'bg-background dark:bg-yt-card border border-default-200 dark:border-yt-border focus-within:!border-yt-red',
            input: 'text-xs text-foreground',
          }}
        />
      </div>

      {/* Segment List (YouTube Transcript Layout) */}
      <CardBody ref={containerRef} className="p-2 space-y-1 overflow-y-auto flex-1 max-h-[580px]">
        {filteredSegments.length === 0 ? (
          <div className="text-center py-16 text-default-400 text-xs">
            {searchQuery ? 'No matching subtitle cues found.' : 'No transcript segments available.'}
          </div>
        ) : (
          filteredSegments.map((segment) => {
            const isActive =
              currentTime >= segment.start_time && currentTime <= segment.end_time;

            return (
              <div
                key={segment.sequence_number}
                ref={isActive ? activeSegmentRef : null}
                onClick={() => onSeek(segment.start_time)}
                className={`p-2.5 rounded-xl cursor-pointer transition-all duration-150 flex items-start gap-3 border ${
                  isActive
                    ? 'bg-default-100 dark:bg-yt-surface border-l-4 border-l-yt-red border-t-transparent border-r-transparent border-b-transparent shadow-xs'
                    : 'border-transparent hover:bg-default-100/60 dark:hover:bg-yt-surface/60'
                }`}
              >
                {/* Clickable Blue/Red Timestamp Pill */}
                <span
                  className={`text-xs font-mono font-bold shrink-0 mt-0.5 px-2 py-0.5 rounded-md ${
                    isActive
                      ? 'bg-yt-red text-white'
                      : 'text-blue-500 hover:text-blue-400 dark:text-blue-400 bg-blue-500/10'
                  }`}
                >
                  {formatTime(segment.start_time)}
                </span>

                {/* Cue Text */}
                <div className="flex-1 min-w-0">
                  <p
                    className={`text-xs sm:text-sm leading-relaxed ${
                      isActive ? 'text-foreground font-semibold' : 'text-default-700 dark:text-default-300'
                    }`}
                  >
                    {segment.text}
                  </p>
                </div>

                {/* Confidence Badge */}
                {segment.confidence > 0 && (
                  <span className="text-[10px] text-default-400 font-mono shrink-0">
                    {(segment.confidence * 100).toFixed(0)}%
                  </span>
                )}
              </div>
            );
          })
        )}
      </CardBody>
    </Card>
  );
};
