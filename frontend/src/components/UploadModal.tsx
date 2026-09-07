import React, { useState, useRef } from 'react';
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  Progress,
} from '@heroui/react';
import { UploadCloud, FileVideo, CheckCircle2, AlertCircle } from 'lucide-react';
import { api } from '../services/api';
import { Video } from '../types';

interface UploadModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (video: Video) => void;
}

export const UploadModal: React.FC<UploadModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [statusText, setStatusText] = useState('');
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const selected = e.target.files[0];
      setFile(selected);
      setError(null);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const selected = e.dataTransfer.files[0];
      setFile(selected);
      setError(null);
    }
  };

  const resetState = () => {
    setFile(null);
    setUploading(false);
    setProgress(0);
    setStatusText('');
    setError(null);
  };

  const handleUpload = async () => {
    if (!file) {
      setError('Please select a video or audio file.');
      return;
    }

    setUploading(true);
    setError(null);
    setProgress(5);
    setStatusText('Requesting upload URL...');

    try {
      // 1. Request presigned upload URL
      const { upload_url, video_id } = await api.requestUploadUrl({
        filename: file.name,
        content_type: file.type || 'video/mp4',
      });

      // 2. Upload directly (S3 or local proxy)
      setStatusText('Streaming media bytes...');
      await api.uploadToS3(upload_url, file, (percent) => {
        // Map 0-100 progress into 10% - 85% range
        setProgress(10 + Math.round(percent * 0.75));
      });

      // 3. Confirm upload
      setProgress(90);
      setStatusText('Verifying storage object & metadata...');
      const video = await api.completeUpload(video_id);

      // 4. Dispatch transcription job
      setProgress(96);
      setStatusText('Dispatching transcription job...');
      try {
        await api.startTranscription(video_id);
      } catch (jobErr) {
        console.warn('Auto transcription trigger notice:', jobErr);
      }

      setProgress(100);
      setStatusText('Success! Ready.');

      setTimeout(() => {
        onSuccess(video);
        resetState();
        onClose();
      }, 700);
    } catch (err: unknown) {
      const errorMsg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        (err as Error).message ||
        'Upload failed. Please check network connection.';
      setError(errorMsg);
      setUploading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {
        if (!uploading) {
          resetState();
          onClose();
        }
      }}
      size="lg"
      backdrop="blur"
      placement="center"
      classNames={{
        base: 'bg-background dark:bg-yt-dark border border-default-200 dark:border-yt-border rounded-3xl shadow-2xl',
        header: 'border-b border-default-200 dark:border-yt-border pb-3',
        footer: 'border-t border-default-200 dark:border-yt-border pt-3',
      }}
    >
      <ModalContent>
        <ModalHeader className="flex flex-col gap-0.5">
          <h2 className="text-lg font-bold tracking-tight text-foreground flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-yt-red" />
            Upload Video to TranscribeX Studio
          </h2>
          <p className="text-xs text-default-400 font-normal">
            Direct high-speed streaming upload &bull; Asynchronous 16kHz FFmpeg worker engine
          </p>
        </ModalHeader>

        <ModalBody className="py-5">
          {error && (
            <div className="p-3 bg-danger-500/10 text-danger border border-danger-500/30 rounded-2xl text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div className="space-y-4">
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              accept="video/*,audio/*,.mp4,.webm,.mov,.mkv,.mp3,.wav,.m4a"
              className="hidden"
            />

            {!file ? (
              <div
                onDragOver={(e) => e.preventDefault()}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                className="border-2 border-dashed border-default-300 dark:border-yt-border hover:border-yt-red rounded-3xl p-10 flex flex-col items-center justify-center gap-4 cursor-pointer transition-all bg-default-50/50 dark:bg-yt-surface/50 hover:bg-yt-red/5 group"
              >
                <div className="w-16 h-16 rounded-full bg-yt-red/10 text-yt-red flex items-center justify-center group-hover:scale-110 transition-transform shadow-md shadow-red-600/10">
                  <UploadCloud className="w-8 h-8" />
                </div>
                <div className="text-center space-y-1">
                  <p className="text-sm font-bold text-foreground">
                    Select video files to transcribe
                  </p>
                  <p className="text-xs text-default-400">
                    MP4, WebM, MOV, MKV (zero-memory presigned streaming up to 500MB)
                  </p>
                </div>
                <Button
                  size="sm"
                  radius="full"
                  className="bg-default-200 dark:bg-yt-card text-foreground font-semibold text-xs mt-1 pointer-events-none"
                >
                  Select Files
                </Button>
              </div>
            ) : (
              <div className="p-4 border border-default-200 dark:border-yt-border rounded-2xl bg-default-50 dark:bg-yt-surface flex items-center justify-between">
                <div className="flex items-center gap-3 truncate">
                  <div className="w-10 h-10 rounded-xl bg-yt-red/10 text-yt-red flex items-center justify-center shrink-0">
                    <FileVideo className="w-5 h-5" />
                  </div>
                  <div className="truncate">
                    <p className="text-sm font-semibold text-foreground truncate">{file.name}</p>
                    <p className="text-xs text-default-400">
                      {(file.size / (1024 * 1024)).toFixed(2)} MB &bull; {file.type || 'video/mp4'}
                    </p>
                  </div>
                </div>
                {!uploading && (
                  <Button
                    size="sm"
                    variant="light"
                    radius="full"
                    color="danger"
                    onClick={() => setFile(null)}
                    className="text-xs font-semibold"
                  >
                    Change
                  </Button>
                )}
              </div>
            )}

            {uploading && (
              <div className="space-y-2 pt-2">
                <div className="flex justify-between text-xs font-semibold">
                  <span className="text-default-400 flex items-center gap-1.5">
                    {progress === 100 ? (
                      <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                    ) : (
                      <span className="w-2 h-2 rounded-full bg-yt-red animate-ping" />
                    )}
                    {statusText}
                  </span>
                  <span className="text-yt-red font-mono">{progress}%</span>
                </div>
                <Progress
                  value={progress}
                  color={progress === 100 ? 'success' : 'danger'}
                  size="sm"
                  radius="full"
                  aria-label="Upload progress"
                  classNames={{
                    indicator: progress === 100 ? 'bg-emerald-500' : 'bg-yt-red',
                  }}
                />
              </div>
            )}
          </div>
        </ModalBody>

        <ModalFooter className="gap-2">
          <Button
            variant="flat"
            radius="full"
            size="sm"
            onPress={() => {
              resetState();
              onClose();
            }}
            isDisabled={uploading}
            className="text-xs font-semibold"
          >
            Cancel
          </Button>
          <Button
            size="sm"
            radius="full"
            onPress={handleUpload}
            isLoading={uploading}
            isDisabled={!file || uploading}
            className="bg-yt-red hover:bg-red-700 text-white font-semibold text-xs shadow-md shadow-red-600/30 px-5"
          >
            Upload & Transcribe
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
