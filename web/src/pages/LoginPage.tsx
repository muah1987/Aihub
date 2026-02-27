import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { LoginForm } from '../components/auth/LoginForm';
import { RegisterForm } from '../components/auth/RegisterForm';
import { useAuth } from '../hooks/useAuth';
import { Cpu } from 'lucide-react';

export function LoginPage() {
  const [isRegister, setIsRegister] = useState(false);
  const { isAuthenticated } = useAuth();

  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-2 mb-2">
            <Cpu size={32} className="text-[var(--color-primary)]" />
            <h1 className="text-2xl font-bold">Aihub</h1>
          </div>
          <p className="text-sm text-[var(--color-text-secondary)]">
            AI-Native Web CLI Platform
          </p>
        </div>

        <div className="bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-xl p-6">
          <h2 className="text-lg font-semibold mb-4">
            {isRegister ? 'Create Account' : 'Welcome Back'}
          </h2>
          {isRegister ? (
            <RegisterForm onSwitchToLogin={() => setIsRegister(false)} />
          ) : (
            <LoginForm onSwitchToRegister={() => setIsRegister(true)} />
          )}
        </div>
      </div>
    </div>
  );
}
