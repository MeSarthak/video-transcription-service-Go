import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { User, Video, Job, Transcript, AuthTokens, UploadUrlResponse } from '../types';

const API_BASE_URL = '/api/v1';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to attach JWT token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('access_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor for 401 handling & refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      const refreshToken = localStorage.getItem('refresh_token');

      if (refreshToken) {
        try {
          const res = await axios.post<{ tokens: AuthTokens }>(`${API_BASE_URL}/auth/refresh`, {
            refresh_token: refreshToken,
          });
          const newTokens = res.data.tokens;
          localStorage.setItem('access_token', newTokens.access_token);
          localStorage.setItem('refresh_token', newTokens.refresh_token);

          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${newTokens.access_token}`;
          }
          return apiClient(originalRequest);
        } catch {
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

export const api = {
  // Auth
  register: async (email: string, password: string): Promise<{ user: User; tokens: AuthTokens }> => {
    const res = await apiClient.post('/auth/register', { email, password });
    return res.data;
  },
  login: async (email: string, password: string): Promise<{ user: User; tokens: AuthTokens }> => {
    const res = await apiClient.post('/auth/login', { email, password });
    return res.data;
  },
  getMe: async (): Promise<{ user: User }> => {
    const res = await apiClient.get('/auth/me');
    return res.data;
  },

  // Videos
  getVideos: async (): Promise<{ videos: Video[]; total: number }> => {
    const res = await apiClient.get('/videos');
    return res.data;
  },
  getVideo: async (id: string): Promise<Video> => {
    const res = await apiClient.get(`/videos/${id}`);
    return res.data;
  },
  deleteVideo: async (id: string): Promise<{ message: string }> => {
    const res = await apiClient.delete(`/videos/${id}`);
    return res.data;
  },
  requestUploadUrl: async (data: {
    filename: string;
    content_type?: string;
  }): Promise<UploadUrlResponse> => {
    const res = await apiClient.post('/videos/upload-url', data);
    return res.data;
  },
  completeUpload: async (videoId: string): Promise<Video> => {
    const res = await apiClient.post(`/videos/${videoId}/complete-upload`);
    return res.data;
  },
  uploadToS3: async (uploadUrl: string, file: File, onProgress?: (percent: number) => void): Promise<void> => {
    // If uploadUrl starts with /api or is relative, send to backend proxy
    const targetUrl = uploadUrl.startsWith('http') ? uploadUrl : uploadUrl;

    await axios.put(targetUrl, file, {
      headers: {
        'Content-Type': file.type || 'video/mp4',
      },
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total && onProgress) {
          const percent = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(percent);
        }
      },
    });
  },

  // Transcription Jobs
  startTranscription: async (videoId: string, language: string = 'en-US'): Promise<Job> => {
    const res = await apiClient.post(`/videos/${videoId}/transcribe`, { language });
    return res.data;
  },
  getJobStatus: async (videoId: string): Promise<Job> => {
    const res = await apiClient.get(`/videos/${videoId}/transcription/status`);
    return res.data;
  },
  getTranscription: async (videoId: string): Promise<Transcript> => {
    const res = await apiClient.get(`/videos/${videoId}/transcription`);
    return res.data;
  },
  exportTranscript: async (videoId: string, format: 'srt' | 'vtt' | 'txt'): Promise<Blob> => {
    const res = await apiClient.get(`/videos/${videoId}/transcript/export?format=${format}`, {
      responseType: 'blob',
    });
    return res.data;
  },
};
