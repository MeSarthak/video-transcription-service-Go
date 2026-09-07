import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  CardBody,
  CardHeader,
  Input,
  Button,
  Divider,
} from '@heroui/react';
import { Mail, Lock, Video as VideoIcon, ArrowRight, AlertCircle, CheckCircle } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export const LoginPage: React.FC = () => {
  const [isRegister, setIsRegister] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { login, register } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) {
      setError('Please fill in both email and password.');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      if (isRegister) {
        await register(email, password);
      } else {
        await login(email, password);
      }
      navigate('/');
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        (err as Error).message ||
        'Authentication failed.';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[85vh] flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-6">
          <div className="inline-flex p-3 rounded-2xl bg-gradient-to-tr from-primary-600 to-indigo-500 text-white shadow-xl shadow-primary-500/20 mb-3">
            <VideoIcon className="w-8 h-8" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {isRegister ? 'Create an Account' : 'Welcome Back'}
          </h1>
          <p className="text-xs text-default-400 mt-1">
            Sign in to access distributed video transcription and cloud pipelines
          </p>
        </div>

        <Card className="border border-default-200 dark:border-default-800 shadow-xl bg-background/80 backdrop-blur-lg">
          <CardHeader className="pb-0 pt-6 px-6 flex justify-between items-center">
            <span className="text-sm font-semibold uppercase tracking-wider text-primary">
              {isRegister ? 'Registration' : 'Sign In'}
            </span>
            <Button
              size="sm"
              variant="light"
              color="primary"
              onClick={() => {
                setIsRegister(!isRegister);
                setError(null);
              }}
            >
              {isRegister ? 'Have an account? Sign In' : "Don't have an account? Sign Up"}
            </Button>
          </CardHeader>

          <CardBody className="p-6">
            {error && (
              <div className="p-3 mb-4 bg-danger-50 dark:bg-danger-900/20 text-danger border border-danger-200 dark:border-danger-800 rounded-xl text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                <span>{error}</span>
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <Input
                label="Email Address"
                placeholder="developer@example.com"
                type="email"
                value={email}
                onValueChange={setEmail}
                startContent={<Mail className="w-4 h-4 text-default-400" />}
                variant="bordered"
                isRequired
              />

              <Input
                label="Password"
                placeholder="••••••••••••"
                type="password"
                value={password}
                onValueChange={setPassword}
                startContent={<Lock className="w-4 h-4 text-default-400" />}
                variant="bordered"
                isRequired
              />

              <Button
                type="submit"
                color="primary"
                fullWidth
                isLoading={loading}
                endContent={!loading && <ArrowRight className="w-4 h-4" />}
                className="mt-2 font-medium shadow-md shadow-primary/20"
              >
                {isRegister ? 'Create Account' : 'Sign In'}
              </Button>
            </form>

            <div className="mt-6 pt-6 border-t border-default-100 dark:border-default-800 text-center">
              <div className="flex items-center justify-center gap-1.5 text-xs text-default-400">
                <CheckCircle className="w-3.5 h-3.5 text-success" />
                <span>JWT Access + Refresh Token pairs</span>
              </div>
            </div>
          </CardBody>
        </Card>
      </div>
    </div>
  );
};
