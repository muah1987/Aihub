import { useState, FormEvent } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { useAuth } from '../../hooks/useAuth';
import { useAuthStore } from '../../store/authStore';
import { TwoFactorForm } from './TwoFactorForm';

interface LoginFormProps {
  onSwitchToRegister: () => void;
}

export function LoginForm({ onSwitchToRegister }: LoginFormProps) {
  const { login, loading, error } = useAuth();
  const pendingTwoFactorToken = useAuthStore((s) => s.pendingTwoFactorToken);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    try {
      await login(email, password);
    } catch {
      // Error handled by store
    }
  };

  if (pendingTwoFactorToken) {
    return <TwoFactorForm />;
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        label="Email"
        type="email"
        placeholder="you@example.com"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
      />
      <Input
        label="Password"
        type="password"
        placeholder="Min 8 characters"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        required
      />
      {error && <p className="text-sm text-[var(--color-error)]">{error}</p>}
      <Button type="submit" loading={loading} className="w-full">
        Sign In
      </Button>
      <p className="text-sm text-center text-[var(--color-text-secondary)]">
        Don't have an account?{' '}
        <button
          type="button"
          onClick={onSwitchToRegister}
          className="text-[var(--color-primary)] hover:underline"
        >
          Register
        </button>
      </p>
    </form>
  );
}
