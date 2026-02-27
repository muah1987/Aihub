import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';
import { Button } from '../components/ui/Button';

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [errorMessage, setErrorMessage] = useState('');

  useEffect(() => {
    const token = searchParams.get('token');
    if (!token) {
      setStatus('error');
      setErrorMessage('No verification token provided');
      return;
    }

    authApi.verifyEmail(token)
      .then(() => setStatus('success'))
      .catch((err) => {
        setStatus('error');
        setErrorMessage(err?.response?.data?.error || 'Verification failed');
      });
  }, [searchParams]);

  return (
    <div className="min-h-screen bg-[var(--color-bg)] flex items-center justify-center p-4">
      <div className="bg-[var(--color-bg-secondary)] rounded-2xl p-8 w-full max-w-md text-center space-y-4">
        {status === 'loading' && (
          <>
            <Loader2 size={48} className="animate-spin mx-auto text-[var(--color-primary)]" />
            <p className="text-[var(--color-text-secondary)]">Verifying your email...</p>
          </>
        )}
        {status === 'success' && (
          <>
            <CheckCircle size={48} className="mx-auto text-[var(--color-success)]" />
            <h2 className="text-xl font-bold">Email Verified!</h2>
            <p className="text-[var(--color-text-secondary)]">Your email has been successfully verified.</p>
            <Button onClick={() => navigate('/')} className="w-full">Go to Dashboard</Button>
          </>
        )}
        {status === 'error' && (
          <>
            <XCircle size={48} className="mx-auto text-[var(--color-error)]" />
            <h2 className="text-xl font-bold">Verification Failed</h2>
            <p className="text-[var(--color-text-secondary)]">{errorMessage}</p>
            <Button onClick={() => navigate('/login')} className="w-full">Go to Login</Button>
          </>
        )}
      </div>
    </div>
  );
}
