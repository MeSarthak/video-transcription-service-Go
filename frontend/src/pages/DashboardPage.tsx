import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Card,
  CardBody,
  CardFooter,
  Button,
  Input,
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Spinner,
} from '@heroui/react';
import {
  Video as VideoIcon,
  Search,
  Upload,
  Trash2,
  Play,
  Clock,
  HardDrive,
  FileVideo,
  Sparkles,
  RefreshCw,
} from 'lucide-react';
import { api } from '../services/api';
import { Video } from '../types';
import { JobStatusBadge } from '../components/JobStatusBadge';
import { UploadModal } from '../components/UploadModal';

export const DashboardPage: React.FC = () => {
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [deleteVideoId, setDeleteVideoId] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchVideos = useCallback(async () => {
    try {
      const res = await api.getVideos();
      setVideos(res.videos || []);
    } catch (err) {
      console.error('Failed to fetch videos:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchVideos();
    // Poll every 3s if any video is processing or queued
    const interval = setInterval(() => {
      const hasActive = videos.some(
        (v) =>
          v.status === 'processing' ||
          v.status === 'uploading' ||
          v.status === 'uploaded'
      );
      if (hasActive) {
        fetchVideos();
      }
    }, 3000);

    return () => clearInterval(interval);
  }, [fetchVideos, videos]);

  const handleDelete = async () => {
    if (!deleteVideoId) return;
    setDeleting(true);
    try {
      await api.deleteVideo(deleteVideoId);
      setVideos((prev) => prev.filter((v) => v.id !== deleteVideoId));
      setDeleteVideoId(null);
    } catch (err) {
      console.error('Failed to delete video:', err);
    } finally {
      setDeleting(false);
    }
  };

  const filteredVideos = videos.filter((v) =>
    v.filename.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const formatFileSize = (bytes: number) => {
    if (!bytes) return '0 MB';
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDuration = (secs?: number) => {
    if (!secs) return '--:--';
    const m = Math.floor(secs / 60);
    const s = Math.floor(secs % 60);
    return `${m}:${s.toString().padStart(2, '0')}`;
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-8">
      {/* Hero Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-6 border-b border-default-200 dark:border-default-800">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight text-foreground flex items-center gap-3">
            <span>Video Transcription Dashboard</span>
            <span className="p-1 rounded-full bg-primary-100 dark:bg-primary-900/40 text-primary text-xs font-semibold px-2.5">
              Live
            </span>
          </h1>
          <p className="text-sm text-default-500 mt-1">
            Asynchronous media pipeline with FFmpeg normalization, Speech-to-Text & Subtitle sync.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="flat"
            size="sm"
            onClick={fetchVideos}
            startContent={<RefreshCw className="w-4 h-4" />}
          >
            Refresh
          </Button>
          <Button
            color="primary"
            size="md"
            startContent={<Upload className="w-4 h-4" />}
            onClick={() => setIsUploadOpen(true)}
            className="shadow-md shadow-primary/20 font-medium"
          >
            Upload Video
          </Button>
        </div>
      </div>

      {/* Stats Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="border border-default-200 dark:border-default-800 bg-background/50 backdrop-blur-sm">
          <CardBody className="flex flex-row items-center gap-4 p-4">
            <div className="p-3 rounded-xl bg-primary-500/10 text-primary">
              <FileVideo className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs text-default-400 font-medium">Total Videos</p>
              <h3 className="text-2xl font-bold">{videos.length}</h3>
            </div>
          </CardBody>
        </Card>

        <Card className="border border-default-200 dark:border-default-800 bg-background/50 backdrop-blur-sm">
          <CardBody className="flex flex-row items-center gap-4 p-4">
            <div className="p-3 rounded-xl bg-warning-500/10 text-warning">
              <Clock className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs text-default-400 font-medium">Active Processing</p>
              <h3 className="text-2xl font-bold">
                {videos.filter((v) => v.status === 'processing' || v.status === 'uploading' || v.status === 'uploaded').length}
              </h3>
            </div>
          </CardBody>
        </Card>

        <Card className="border border-default-200 dark:border-default-800 bg-background/50 backdrop-blur-sm">
          <CardBody className="flex flex-row items-center gap-4 p-4">
            <div className="p-3 rounded-xl bg-success-500/10 text-success">
              <Sparkles className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs text-default-400 font-medium">Completed</p>
              <h3 className="text-2xl font-bold">
                {videos.filter((v) => v.status === 'completed').length}
              </h3>
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Search & Filter Toolbar */}
      <div className="flex items-center justify-between gap-4">
        <Input
          placeholder="Search by video filename..."
          value={searchQuery}
          onValueChange={setSearchQuery}
          startContent={<Search className="w-4 h-4 text-default-400" />}
          className="max-w-md"
          variant="bordered"
          size="sm"
          isClearable
        />
      </div>

      {/* Video Grid */}
      {loading ? (
        <div className="flex flex-col items-center justify-center py-20 gap-3">
          <Spinner size="lg" color="primary" />
          <p className="text-sm text-default-400">Loading your video library...</p>
        </div>
      ) : filteredVideos.length === 0 ? (
        <div className="text-center py-20 border-2 border-dashed border-default-200 dark:border-default-800 rounded-2xl p-8 space-y-4">
          <div className="p-4 rounded-full bg-default-100 dark:bg-default-800 w-16 h-16 mx-auto flex items-center justify-center text-default-400">
            <VideoIcon className="w-8 h-8" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-foreground">No videos yet</h3>
            <p className="text-xs text-default-400 max-w-sm mx-auto mt-1">
              {searchQuery
                ? 'No videos match your search query.'
                : 'Upload your first video to extract audio and generate synchronized transcripts.'}
            </p>
          </div>
          <Button
            color="primary"
            size="sm"
            startContent={<Upload className="w-4 h-4" />}
            onClick={() => setIsUploadOpen(true)}
          >
            Upload Now
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredVideos.map((video) => (
            <Card
              key={video.id}
              className="border border-default-200 dark:border-default-800 hover:border-primary/50 transition-all shadow-sm hover:shadow-md group flex flex-col justify-between"
            >
              <CardBody className="p-5 space-y-3">
                <div className="flex items-start justify-between gap-2">
                  <h4 className="font-bold text-base text-foreground line-clamp-1 group-hover:text-primary transition-colors">
                    {video.filename}
                  </h4>
                  <JobStatusBadge status={video.status} />
                </div>

                <div className="flex items-center gap-4 text-xs text-default-500 pt-2 border-t border-default-100 dark:border-default-800/60">
                  <span className="flex items-center gap-1">
                    <Clock className="w-3.5 h-3.5 text-default-400" />
                    {formatDuration(video.duration_seconds)}
                  </span>
                  <span className="flex items-center gap-1">
                    <HardDrive className="w-3.5 h-3.5 text-default-400" />
                    {formatFileSize(video.size_bytes)}
                  </span>
                </div>
              </CardBody>

              <CardFooter className="px-5 py-3 bg-default-50/50 dark:bg-default-100/10 border-t border-default-100 dark:border-default-800 flex items-center justify-between">
                <Button
                  as={Link}
                  to={`/videos/${video.id}`}
                  size="sm"
                  color="primary"
                  variant="flat"
                  startContent={<Play className="w-3.5 h-3.5" />}
                  className="font-medium"
                >
                  View Captions
                </Button>

                <Button
                  isIconOnly
                  size="sm"
                  variant="light"
                  color="danger"
                  onClick={() => setDeleteVideoId(video.id)}
                  aria-label="Delete video"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </CardFooter>
            </Card>
          ))}
        </div>
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!deleteVideoId}
        onClose={() => setDeleteVideoId(null)}
        placement="center"
        size="sm"
      >
        <ModalContent>
          <ModalHeader>Confirm Deletion</ModalHeader>
          <ModalBody>
            <p className="text-sm text-default-600">
              Are you sure you want to permanently delete this video, storage objects, and all generated transcripts?
            </p>
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setDeleteVideoId(null)}>
              Cancel
            </Button>
            <Button
              color="danger"
              onPress={handleDelete}
              isLoading={deleting}
            >
              Delete Video
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Upload Modal */}
      <UploadModal
        isOpen={isUploadOpen}
        onClose={() => setIsUploadOpen(false)}
        onSuccess={(newVideo) => {
          setVideos((prev) => [newVideo, ...prev]);
        }}
      />
    </div>
  );
};
