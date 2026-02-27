import { useState, FormEvent } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { useAuthStore } from '../../store/authStore';

export function TwoFactorForm() {
  const { verifyTwoFactor, loading, error } = useAuthStore();
  const [code, setCode] = useState('');

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    try {
      await verifyTwoFactor(code);
    } catch {
      // Error handled by store
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="text-center mb-4">
        <h3 className="text-lg font-semibold">Two-Factor Authentication</h3>
        <p className="text-sm text-[var(--color-text-secondary)] mt-1">
          Enter the code from your authenticator app
        </p>
      </div>
      <Input
        label="Authentication Code"
        type="text"
        placeholder="000000"
        value={code}
        onChange={(e) => setCode(e.target.value)}
        maxLength={8}
        required
        autoFocus
      />
      {error && <p className="text-sm text-[var(--color-error)]">{error}</p>}
      <Button type="submit" loading={loading} className="w-full">
        Verify
      </Button>
      <p className="text-xs text-center text-[var(--color-text-secondary)]">
        You can also enter a backup code
      </p>
    </form>
  );
}
