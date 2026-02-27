import { useState } from 'react';
import { AlertTriangle } from 'lucide-react';
import { authApi } from '../../api/auth';
import { Button } from '../ui/Button';

export function EmailVerificationBanner() {
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleResend = async () => {
    setLoading(true);
    try {
      await authApi.resendVerification();
      setSent(true);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-yellow-500/10 border-b border-yellow-500/30 px-4 py-2 flex items-center justify-between gap-3">
      <div className="flex items-center gap-2 text-yellow-400 text-sm">
        <AlertTriangle size={16} />
        <span>Please verify your email address</span>
      </div>
      {sent ? (
        <span className="text-xs text-[var(--color-text-secondary)]">Verification email sent!</span>
      ) : (
        <Button size="sm" variant="ghost" onClick={handleResend} loading={loading}>
          Resend
        </Button>
      )}
    </div>
  );
}
