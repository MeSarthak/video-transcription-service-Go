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
    const ms = Math.floor((seconds % 1) * 10);
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}.${ms}`;
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
    <Card className="h-full flex flex-col border border-default-200 dark:border-default-800 bg-background/60 backdrop-blur-md shadow-sm">
      {/* Header controls */}
      <div className="p-4 border-b border-default-200 dark:border-default-800 flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="p-1.5 rounded-lg bg-primary/10 text-primary">
            <Sparkles className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-foreground">Interactive Transcript</h3>
            <p className="text-[11px] text-default-400">
              {segments.length} timestamped subtitle segments
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="flat"
            onClick={() => setAutoScroll(!autoScroll)}
            className={`text-xs ${autoScroll ? 'text-primary' : 'text-default-400'}`}
          >
            {autoScroll ? 'Auto-scroll: ON' : 'Auto-scroll: OFF'}
          </Button>
          <Button
            size="sm"
            variant="flat"
            startContent={copiedFull ? <Check className="w-3.5 h-3.5 text-success" /> : <Copy className="w-3.5 h-3.5" />}
            onClick={handleCopyFull}
            className="text-xs"
          >
            {copiedFull ? 'Copied' : 'Copy All'}
          </Button>
        </div>
      </div>

      {/* Search Bar */}
      <div className="p-3 border-b border-default-100 dark:border-default-800/60 bg-default-50/40 dark:bg-default-100/10">
        <Input
          size="sm"
          placeholder="Search keywords in transcript..."
          value={searchQuery}
          onValueChange={setSearchQuery}
          startContent={<Search className="w-3.5 h-3.5 text-default-400" />}
          isClearable
          variant="bordered"
        />
      </div>

      {/* Segment List */}
      <CardBody ref={containerRef} className="p-3 space-y-2 overflow-y-auto max-h-[520px]">
        {filteredSegments.length === 0 ? (
          <div className="text-center py-12 text-default-400 text-sm">
            {searchQuery ? 'No matching cues found.' : 'No transcript segments available.'}
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
                className={`p-3 rounded-xl cursor-pointer transition-all duration-150 border ${
                  isActive
                    ? 'bg-primary-500/10 border-primary shadow-sm scale-[1.01]'
                    : 'bg-default-50/50 dark:bg-default-100/20 border-transparent hover:border-default-300 dark:hover:border-default-700'
                }`}
              >
                <div className="flex items-center justify-between mb-1.5">
                  <span
                    className={`inline-flex items-center gap-1 text-xs font-mono font-semibold px-2 py-0.5 rounded-md ${
                      isActive
                        ? 'bg-primary text-white shadow-xs'
                        : 'bg-default-200 dark:bg-default-800 text-default-600'
                    }`}
                  >
                    <Clock className="w-3 h-3" />
                    {formatTime(segment.start_time)} - {formatTime(segment.end_time)}
                  </span>

                  <div className="flex items-center gap-1.5">
                    {segment.confidence > 0 && (
                      <Tooltip content={`Transcription confidence: ${(segment.confidence * 100).toFixed(0)}%`}>
                        <Chip
                          size="sm"
                          variant="dot"
                          color={segment.confidence > 0.85 ? 'success' : 'warning'}
                          className="h-5 text-[10px]"
                        >
                          {(segment.confidence * 100).toFixed(0)}%
                        </Chip>
                      </Tooltip>
                    )}
                    <Button
                      isIconOnly
                      size="sm"
                      variant="light"
                      className="w-6 h-6 min-w-6 text-default-400 hover:text-primary"
                      onClick={(e) => {
                        e.stopPropagation();
                        onSeek(segment.start_time);
                      }}
                      aria-label="Play segment"
                    >
                      <Play className="w-3 h-3" />
                    </Button>
                  </div>
                </div>

                <p
                  className={`text-sm leading-relaxed ${
                    isActive ? 'text-foreground font-medium' : 'text-default-700 dark:text-default-300'
                  }`}
                >
                  {segment.text}
                </p>
              </div>
            );
          })
        )}
      </CardBody>
    </Card>
  );
};
