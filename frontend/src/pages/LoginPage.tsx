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
    <div className="min-h-[80vh] flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-6">
        <div className="text-center space-y-2">
          <div className="w-14 h-10 rounded-xl bg-yt-red text-white flex items-center justify-center shadow-lg shadow-red-600/30 mx-auto">
            <VideoIcon className="w-6 h-6 fill-white ml-0.5" />
          </div>
          <h1 className="text-2xl font-extrabold tracking-tight text-foreground">
            {isRegister ? 'Create your Studio Account' : 'Sign in to TranscribeX'}
          </h1>
          <p className="text-xs text-default-400">
            Distributed asynchronous speech recognition & media processing
          </p>
        </div>

        <Card className="border border-default-200 dark:border-yt-border shadow-2xl bg-background dark:bg-yt-dark rounded-3xl">
          <CardHeader className="pb-0 pt-6 px-6 flex justify-between items-center border-b border-default-100 dark:border-yt-border pb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-yt-red">
              {isRegister ? 'Registration' : 'Authentication'}
            </span>
            <Button
              size="sm"
              variant="light"
              radius="full"
              className="text-xs font-semibold text-default-400 hover:text-foreground"
              onClick={() => {
                setIsRegister(!isRegister);
                setError(null);
              }}
            >
              {isRegister ? 'Have an account? Sign In' : "New? Create Account"}
            </Button>
          </CardHeader>

          <CardBody className="p-6 space-y-4">
            {error && (
              <div className="p-3 bg-danger-500/10 text-danger border border-danger-500/30 rounded-2xl text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
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
                radius="lg"
                isRequired
                classNames={{
                  inputWrapper: 'border-default-200 dark:border-yt-border focus-within:!border-yt-red',
                }}
              />

              <Input
                label="Password"
                placeholder="••••••••••••"
                type="password"
                value={password}
                onValueChange={setPassword}
                startContent={<Lock className="w-4 h-4 text-default-400" />}
                variant="bordered"
                radius="lg"
                isRequired
                classNames={{
                  inputWrapper: 'border-default-200 dark:border-yt-border focus-within:!border-yt-red',
                }}
              />

              <Button
                type="submit"
                fullWidth
                radius="full"
                isLoading={loading}
                endContent={!loading && <ArrowRight className="w-4 h-4" />}
                className="mt-2 bg-yt-red hover:bg-red-700 text-white font-semibold text-sm shadow-md shadow-red-600/30 h-11"
              >
                {isRegister ? 'Create Account' : 'Sign In'}
              </Button>
            </form>

            <div className="pt-4 border-t border-default-100 dark:border-yt-border text-center">
              <div className="flex items-center justify-center gap-1.5 text-xs text-default-400">
                <CheckCircle className="w-3.5 h-3.5 text-emerald-500" />
                <span>HMAC-SHA256 Cryptographic Sessions</span>
              </div>
            </div>
          </CardBody>
        </Card>
      </div>
    </div>
  );
};
