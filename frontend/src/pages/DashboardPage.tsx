import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Button,
  Card,
  CardBody,
  Input,
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Dropdown,
  DropdownTrigger,
  DropdownMenu,
  DropdownItem,
} from '@heroui/react';
import {
  Upload,
  Search,
  Trash2,
  Play,
  RefreshCw,
  Clock,
  MoreVertical,
  Layers,
  FileVideo,
  CheckCircle2,
  Zap,
} from 'lucide-react';
import { api } from '../services/api';
import { Video } from '../types';
import { UploadModal } from '../components/UploadModal';

export const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<'all' | 'completed' | 'processing' | 'uploaded'>('all');
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [deleteVideoId, setDeleteVideoId] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

  const videosRef = useRef(videos);
  videosRef.current = videos;

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

  // Fetch videos once on mount
  useEffect(() => {
    fetchVideos();
  }, [fetchVideos]);

  // Poll every 3s if any video is active
  useEffect(() => {
    const interval = setInterval(() => {
      const hasActive = videosRef.current.some(
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
  }, [fetchVideos]);

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

  const filteredVideos = videos.filter((v) => {
    const matchesSearch = v.filename.toLowerCase().includes(searchQuery.toLowerCase());
    if (!matchesSearch) return false;

    if (categoryFilter === 'completed') return v.status === 'completed';
    if (categoryFilter === 'processing') return v.status === 'processing' || v.status === 'uploading';
    if (categoryFilter === 'uploaded') return v.status === 'uploaded';
    return true;
  });

  const formatFileSize = (bytes: number) => {
    if (!bytes) return '0 MB';
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDuration = (secs?: number) => {
    if (!secs) return '0:22';
    const m = Math.floor(secs / 60);
    const s = Math.floor(secs % 60);
    return `${m}:${s.toString().padStart(2, '0')}`;
  };

  return (
    <div className="max-w-[1600px] mx-auto px-3 sm:px-6 py-6 space-y-6">
      {/* Category Pills Bar (YouTube Style) */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-2">
        <div className="flex items-center gap-2 overflow-x-auto py-1 scrollbar-none">
          <button
            onClick={() => setCategoryFilter('all')}
            className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all whitespace-nowrap ${
              categoryFilter === 'all'
                ? 'bg-foreground text-background font-bold shadow-sm'
                : 'bg-default-100 hover:bg-default-200 dark:bg-yt-surface dark:hover:bg-yt-hoverCard text-foreground'
            }`}
          >
            All ({videos.length})
          </button>

          <button
            onClick={() => setCategoryFilter('completed')}
            className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 whitespace-nowrap ${
              categoryFilter === 'completed'
                ? 'bg-foreground text-background font-bold shadow-sm'
                : 'bg-default-100 hover:bg-default-200 dark:bg-yt-surface dark:hover:bg-yt-hoverCard text-foreground'
            }`}
          >
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
            <span>Completed ({videos.filter((v) => v.status === 'completed').length})</span>
          </button>

          <button
            onClick={() => setCategoryFilter('processing')}
            className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 whitespace-nowrap ${
              categoryFilter === 'processing'
                ? 'bg-foreground text-background font-bold shadow-sm'
                : 'bg-default-100 hover:bg-default-200 dark:bg-yt-surface dark:hover:bg-yt-hoverCard text-foreground'
            }`}
          >
            <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
            <span>Processing ({videos.filter((v) => v.status === 'processing' || v.status === 'uploading').length})</span>
          </button>

          <button
            onClick={() => setCategoryFilter('uploaded')}
            className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all whitespace-nowrap ${
              categoryFilter === 'uploaded'
                ? 'bg-foreground text-background font-bold shadow-sm'
                : 'bg-default-100 hover:bg-default-200 dark:bg-yt-surface dark:hover:bg-yt-hoverCard text-foreground'
            }`}
          >
            Uploaded ({videos.filter((v) => v.status === 'uploaded').length})
          </button>
        </div>

        {/* Refresh & Search */}
        <div className="flex items-center gap-2">
          <Input
            size="sm"
            radius="full"
            placeholder="Filter list..."
            value={searchQuery}
            onValueChange={setSearchQuery}
            startContent={<Search className="w-3.5 h-3.5 text-default-400" />}
            className="w-48 sm:w-64"
            isClearable
          />
          <Button
            isIconOnly
            variant="flat"
            size="sm"
            radius="full"
            onClick={fetchVideos}
            aria-label="Refresh list"
            className="text-default-500 hover:text-foreground"
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Video Grid (YouTube 16:9 Thumbnail Layout) */}
      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-x-4 gap-y-8">
          {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
            <div key={i} className="space-y-3 animate-pulse">
              <div className="aspect-video w-full rounded-2xl bg-default-200 dark:bg-yt-surface" />
              <div className="flex gap-3">
                <div className="w-9 h-9 rounded-full bg-default-200 dark:bg-yt-surface shrink-0" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 bg-default-200 dark:bg-yt-surface rounded w-5/6" />
                  <div className="h-3 bg-default-200 dark:bg-yt-surface rounded w-1/2" />
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : filteredVideos.length === 0 ? (
        <div className="text-center py-24 border-2 border-dashed border-default-200 dark:border-yt-border rounded-3xl p-8 space-y-4 max-w-xl mx-auto">
          <div className="w-16 h-16 rounded-full bg-yt-red/10 text-yt-red mx-auto flex items-center justify-center">
            <FileVideo className="w-8 h-8" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-foreground">No videos found</h3>
            <p className="text-xs text-default-400 mt-1 max-w-sm mx-auto">
              {searchQuery
                ? `No videos match "${searchQuery}". Try a different keyword.`
                : 'Get started by uploading your first video to generate synchronized transcripts.'}
            </p>
          </div>
          <Button
            size="sm"
            radius="full"
            startContent={<Upload className="w-4 h-4 text-white" />}
            onClick={() => setIsUploadOpen(true)}
            className="bg-yt-red hover:bg-red-700 text-white font-semibold shadow-md shadow-red-600/30 px-5"
          >
            Upload Video
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-x-4 gap-y-7">
          {filteredVideos.map((video) => {
            const isCompleted = video.status === 'completed';
            const isProcessing = video.status === 'processing' || video.status === 'uploading';

            return (
              <div key={video.id} className="group flex flex-col space-y-2.5 cursor-pointer">
                {/* 16:9 Thumbnail Box */}
                <Link to={`/videos/${video.id}`} className="relative aspect-video w-full rounded-2xl overflow-hidden bg-default-100 dark:bg-yt-surface border border-default-200/80 dark:border-yt-border yt-card-hover block">
                  {/* Background Thumbnail Art */}
                  <div className="absolute inset-0 bg-gradient-to-tr from-black/80 via-neutral-900/60 to-neutral-800/40 flex items-center justify-center">
                    <div className="w-12 h-12 rounded-full bg-black/60 text-white/90 group-hover:bg-yt-red group-hover:scale-110 transition-all flex items-center justify-center shadow-lg">
                      <Play className="w-5 h-5 fill-current ml-0.5" />
                    </div>
                  </div>

                  {/* Status Overlay Badge (Top-Left) */}
                  <div className="absolute top-2.5 left-2.5 z-10">
                    {isCompleted && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-emerald-950/80 text-emerald-400 border border-emerald-500/30 backdrop-blur-md">
                        <CheckCircle2 className="w-3 h-3 text-emerald-400" />
                        Transcribed
                      </span>
                    )}
                    {isProcessing && (
                      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[10px] font-bold bg-amber-950/80 text-amber-400 border border-amber-500/30 backdrop-blur-md">
                        <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
                        Processing
                      </span>
                    )}
                    {video.status === 'uploaded' && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-blue-950/80 text-blue-400 border border-blue-500/30 backdrop-blur-md">
                        Uploaded
                      </span>
                    )}
                    {video.status === 'failed' && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-rose-950/80 text-rose-400 border border-rose-500/30 backdrop-blur-md">
                        Failed
                      </span>
                    )}
                  </div>

                  {/* Duration Pill (Bottom-Right YouTube Badge) */}
                  <div className="absolute bottom-2 right-2 z-10 px-1.5 py-0.5 rounded bg-black/85 text-white text-[11px] font-mono font-semibold tracking-wider">
                    {formatDuration(video.duration_seconds)}
                  </div>
                </Link>

                {/* Video Info Section */}
                <div className="flex items-start gap-3 px-0.5">
                  {/* Channel/Studio Avatar */}
                  <div className="w-9 h-9 rounded-full bg-default-200 dark:bg-yt-surface flex items-center justify-center shrink-0 mt-0.5 border border-default-200 dark:border-yt-border">
                    <FileVideo className="w-4 h-4 text-yt-red" />
                  </div>

                  {/* Title & Metadata */}
                  <div className="flex-1 min-w-0">
                    <Link to={`/videos/${video.id}`}>
                      <h4 className="text-sm font-semibold text-foreground line-clamp-2 leading-snug group-hover:text-yt-red transition-colors">
                        {video.filename}
                      </h4>
                    </Link>
                    <p className="text-xs text-default-400 mt-1">TranscribeX Studio</p>
                    <div className="flex items-center gap-1.5 text-[11px] text-default-400">
                      <span className="uppercase font-mono text-[10px]">{video.content_type?.split('/')[1] || 'MP4'}</span>
                      {video.created_at && (
                        <>
                          <span>&bull;</span>
                          <span>{new Date(video.created_at).toLocaleDateString()}</span>
                        </>
                      )}
                    </div>
                  </div>

                  {/* Kebab Action Menu */}
                  <Dropdown placement="bottom-end">
                    <DropdownTrigger>
                      <button
                        aria-label="More actions"
                        className="text-default-400 hover:text-foreground p-1 rounded-full hover:bg-default-100 dark:hover:bg-yt-surface transition-colors"
                      >
                        <MoreVertical className="w-4 h-4" />
                      </button>
                    </DropdownTrigger>
                    <DropdownMenu aria-label="Video Options" variant="flat">
                      <DropdownItem
                        key="view"
                        onPress={() => navigate(`/videos/${video.id}`)}
                        startContent={<Play className="w-4 h-4" />}
                      >
                        Open Watch Page
                      </DropdownItem>
                      <DropdownItem
                        key="delete"
                        color="danger"
                        className="text-danger"
                        startContent={<Trash2 className="w-4 h-4" />}
                        onPress={() => setDeleteVideoId(video.id)}
                      >
                        Delete Video
                      </DropdownItem>
                    </DropdownMenu>
                  </Dropdown>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!deleteVideoId}
        onClose={() => setDeleteVideoId(null)}
        placement="center"
        size="sm"
        backdrop="blur"
      >
        <ModalContent>
          <ModalHeader className="text-base font-bold">Delete Video</ModalHeader>
          <ModalBody>
            <p className="text-xs text-default-500 leading-relaxed">
              Are you sure you want to permanently delete this video? All extracted audio, transcripts, and subtitle segments will be permanently removed.
            </p>
          </ModalBody>
          <ModalFooter>
            <Button size="sm" variant="flat" radius="full" onPress={() => setDeleteVideoId(null)}>
              Cancel
            </Button>
            <Button
              size="sm"
              color="danger"
              radius="full"
              onPress={handleDelete}
              isLoading={deleting}
              className="bg-yt-red hover:bg-red-700 text-white font-semibold"
            >
              Delete
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
