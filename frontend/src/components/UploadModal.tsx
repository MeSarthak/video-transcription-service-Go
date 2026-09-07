import React, { useState, useRef } from 'react';
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  Input,
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
  const [title, setTitle] = useState('');
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [statusText, setStatusText] = useState('');
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const selected = e.target.files[0];
      setFile(selected);
      if (!title) {
        setTitle(selected.name.replace(/\.[^/.]+$/, ''));
      }
      setError(null);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const selected = e.dataTransfer.files[0];
      setFile(selected);
      if (!title) {
        setTitle(selected.name.replace(/\.[^/.]+$/, ''));
      }
      setError(null);
    }
  };

  const resetState = () => {
    setFile(null);
    setTitle('');
    setUploading(false);
    setProgress(0);
    setStatusText('');
    setError(null);
  };

  const handleUpload = async () => {
    if (!file || !title.trim()) {
      setError('Please provide a video file and title');
      return;
    }

    setUploading(true);
    setError(null);
    setProgress(5);
    setStatusText('Requesting S3 Presigned URL...');

    try {
      // 1. Request presigned upload URL
      const { upload_url, video_id } = await api.requestUploadUrl({
        title: title.trim(),
        filename: file.name,
        file_size: file.size,
        mime_type: file.type || 'video/mp4',
      });

      // 2. Upload directly to S3
      setStatusText('Uploading binary directly to S3...');
      await api.uploadToS3(upload_url, file, (percent) => {
        // Map 0-100 progress into 10% - 85% range
        setProgress(10 + Math.round(percent * 0.75));
      });

      // 3. Confirm upload
      setProgress(90);
      setStatusText('Verifying S3 Object & registering video...');
      const completeRes = await api.completeUpload(video_id);

      // 4. Dispatch transcription job
      setProgress(98);
      setStatusText('Dispatching SQS transcription job...');
      await api.startTranscription(video_id);

      setProgress(100);
      setStatusText('Success! Job queued.');

      setTimeout(() => {
        onSuccess(completeRes.video);
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
    >
      <ModalContent>
        <ModalHeader className="flex flex-col gap-1">
          <h2 className="text-xl font-bold tracking-tight">Upload Video for Transcription</h2>
          <p className="text-xs text-default-400 font-normal">
            Direct streaming to AWS S3 &bull; Distributed async processing via SQS worker
          </p>
        </ModalHeader>

        <ModalBody className="py-4">
          {error && (
            <div className="p-3 bg-danger-50 dark:bg-danger-900/20 text-danger border border-danger-200 dark:border-danger-800 rounded-xl text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div className="space-y-4">
            <Input
              label="Video Title"
              placeholder="e.g. Q3 Architecture Overview Keynote"
              value={title}
              onValueChange={setTitle}
              isDisabled={uploading}
              variant="bordered"
              isRequired
            />

            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              accept="video/*,audio/*"
              className="hidden"
            />

            {!file ? (
              <div
                onDragOver={(e) => e.preventDefault()}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                className="border-2 border-dashed border-default-300 dark:border-default-700 hover:border-primary rounded-2xl p-8 flex flex-col items-center justify-center gap-3 cursor-pointer transition-colors bg-default-50/50 hover:bg-default-100/50"
              >
                <div className="p-3 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary">
                  <UploadCloud className="w-8 h-8" />
                </div>
                <div className="text-center">
                  <p className="text-sm font-semibold text-foreground">
                    Click to browse or drag & drop video
                  </p>
                  <p className="text-xs text-default-400 mt-1">
                    MP4, WebM, MOV, MKV, MP3, WAV (up to 500MB)
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-4 border border-default-200 dark:border-default-800 rounded-2xl bg-default-50 dark:bg-default-100/30 flex items-center justify-between">
                <div className="flex items-center gap-3 truncate">
                  <div className="p-2.5 rounded-xl bg-primary-500/10 text-primary">
                    <FileVideo className="w-6 h-6" />
                  </div>
                  <div className="truncate">
                    <p className="text-sm font-medium text-foreground truncate">{file.name}</p>
                    <p className="text-xs text-default-400">
                      {(file.size / (1024 * 1024)).toFixed(2)} MB &bull; {file.type || 'video'}
                    </p>
                  </div>
                </div>
                {!uploading && (
                  <Button
                    size="sm"
                    variant="light"
                    color="danger"
                    onClick={() => setFile(null)}
                  >
                    Change
                  </Button>
                )}
              </div>
            )}

            {uploading && (
              <div className="space-y-2 pt-2">
                <div className="flex justify-between text-xs font-medium">
                  <span className="text-default-500 flex items-center gap-1.5">
                    {progress === 100 ? (
                      <CheckCircle2 className="w-3.5 h-3.5 text-success" />
                    ) : (
                      <span className="w-2 h-2 rounded-full bg-primary animate-ping" />
                    )}
                    {statusText}
                  </span>
                  <span className="text-primary font-semibold">{progress}%</span>
                </div>
                <Progress
                  value={progress}
                  color={progress === 100 ? 'success' : 'primary'}
                  size="sm"
                  aria-label="Upload progress"
                />
              </div>
            )}
          </div>
        </ModalBody>

        <ModalFooter>
          <Button
            variant="flat"
            onPress={() => {
              resetState();
              onClose();
            }}
            isDisabled={uploading}
          >
            Cancel
          </Button>
          <Button
            color="primary"
            onPress={handleUpload}
            isLoading={uploading}
            isDisabled={!file || !title.trim()}
            className="shadow-md shadow-primary/20"
          >
            {uploading ? 'Processing...' : 'Upload & Transcribe'}
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
