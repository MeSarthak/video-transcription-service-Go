import React from 'react';
import { Chip } from '@heroui/react';
import { CheckCircle2, Clock, AlertTriangle, Loader2 } from 'lucide-react';

interface JobStatusBadgeProps {
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'uploaded' | 'ready';
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
          {status}
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
    case 'pending':
    case 'uploaded':
      return (
        <Chip
          color="warning"
          variant="flat"
          size={size}
          startContent={<Clock className="w-3.5 h-3.5 text-warning ml-1" />}
          className="capitalize font-medium"
        >
          {status}
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
