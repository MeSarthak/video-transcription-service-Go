import React, { useState } from 'react';
import {
  Dropdown,
  DropdownTrigger,
  DropdownMenu,
  DropdownItem,
  Button,
} from '@heroui/react';
import { Download, FileText, Subtitles, FileCode2 } from 'lucide-react';
import { api } from '../services/api';

interface ExportDropdownProps {
  videoId: string;
  videoTitle: string;
  disabled?: boolean;
}

export const ExportDropdown: React.FC<ExportDropdownProps> = ({
  videoId,
  videoTitle,
  disabled = false,
}) => {
  const [downloading, setDownloading] = useState(false);

  const handleExport = async (format: 'srt' | 'vtt' | 'txt') => {
    try {
      setDownloading(true);
      const blob = await api.exportTranscript(videoId, format);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      const sanitized = videoTitle.replace(/[^a-zA-Z0-9_-]/g, '_').toLowerCase();
      a.download = `${sanitized}.${format}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      console.error('Failed to export transcript:', err);
    } finally {
      setDownloading(false);
    }
  };

  return (
    <Dropdown placement="bottom-end">
      <DropdownTrigger>
        <Button
          color="primary"
          variant="flat"
          size="sm"
          startContent={<Download className="w-4 h-4" />}
          isLoading={downloading}
          isDisabled={disabled}
        >
          Export Captions
        </Button>
      </DropdownTrigger>
      <DropdownMenu
        aria-label="Export Formats"
        onAction={(key) => handleExport(key as 'srt' | 'vtt' | 'txt')}
      >
        <DropdownItem
          key="srt"
          description="SubRip Subtitle format with millisecond timestamps"
          startContent={<Subtitles className="w-4 h-4 text-primary" />}
        >
          SRT Subtitles (.srt)
        </DropdownItem>
        <DropdownItem
          key="vtt"
          description="HTML5 Web Video Text Tracks standard"
          startContent={<FileCode2 className="w-4 h-4 text-secondary" />}
        >
          WebVTT (.vtt)
        </DropdownItem>
        <DropdownItem
          key="txt"
          description="Clean plaintext transcript without timestamps"
          startContent={<FileText className="w-4 h-4 text-warning" />}
        >
          Plain Text (.txt)
        </DropdownItem>
      </DropdownMenu>
    </Dropdown>
  );
};
