import React from 'react';
import { Chip } from '@heroui/react';
import { CheckCircle2, Clock, AlertTriangle, Loader2, UploadCloud } from 'lucide-react';

interface JobStatusBadgeProps {
  status: string;
  size?: 'sm' | 'md' | 'lg';
}

export const JobStatusBadge: React.FC<JobStatusBadgeProps> = ({ status, size = 'sm' }) => {
  switch (status) {
    case 'completed':
    case 'ready':
      return (
        <Chip
          color="success"
          variant="flat"
          size={size}
          startContent={<CheckCircle2 className="w-3.5 h-3.5 text-success ml-1" />}
          className="capitalize font-medium"
        >
          Completed
        </Chip>
      );
    case 'processing':
      return (
        <Chip
          color="primary"
          variant="flat"
          size={size}
          startContent={<Loader2 className="w-3.5 h-3.5 text-primary animate-spin ml-1" />}
          className="capitalize font-medium"
        >
          Processing
        </Chip>
      );
    case 'queued':
    case 'uploaded':
    case 'pending':
      return (
        <Chip
          color="warning"
          variant="flat"
          size={size}
          startContent={<Clock className="w-3.5 h-3.5 text-warning ml-1" />}
          className="capitalize font-medium"
        >
          {status === 'queued' ? 'Queued' : 'Uploaded'}
        </Chip>
      );
    case 'uploading':
      return (
        <Chip
          color="primary"
          variant="flat"
          size={size}
          startContent={<UploadCloud className="w-3.5 h-3.5 text-primary ml-1" />}
          className="capitalize font-medium"
        >
          Uploading
        </Chip>
      );
    case 'failed':
      return (
        <Chip
          color="danger"
          variant="flat"
          size={size}
          startContent={<AlertTriangle className="w-3.5 h-3.5 text-danger ml-1" />}
          className="capitalize font-medium"
        >
          Failed
        </Chip>
      );
    default:
      return (
        <Chip variant="flat" size={size} className="capitalize">
          {status}
        </Chip>
      );
  }
};
