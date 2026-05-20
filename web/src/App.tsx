import { useEffect, useState } from 'react';
import './index.css';
import { getSession, login } from './lib/api';
import type { AuthSessionResponse } from './lib/types';
import { LoginPage } from './pages/LoginPage';
import { UsagePage } from './pages/UsagePage';

type SessionLoader = (signal?: AbortSignal) => Promise<AuthSessionResponse>;
type LoginRequest = (password: string) => Promise<void>;

interface AppViewProps {
  authenticated: boolean | null;
  error?: string;
  loading?: boolean;
  onLogin: (password: string) => Promise<void>;
}

export async function checkSessionAuthenticated(loadSession: SessionLoader = getSession, signal?: AbortSignal): Promise<boolean> {
  const session = await loadSession(signal);
  return session.authenticated;
}

export async function loginAndAuthenticate(loginRequest: LoginRequest = login, password: string): Promise<boolean> {
  await loginRequest(password);
  return true;
}

export function AppView({ authenticated, error = '', loading = false, onLogin }: AppViewProps) {
  if (authenticated === null) {
    return null;
  }
  if (authenticated) {
    return <UsagePage />;
  }
  return <LoginPage loading={loading} error={error} onSubmit={onLogin} />;
}

function App() {
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState('');

  useEffect(() => {
    const controller = new AbortController();

    void checkSessionAuthenticated(getSession, controller.signal)
      .then(setAuthenticated)
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return;
        }
        setAuthenticated(false);
      });

    return () => controller.abort();
  }, []);

  const handleLogin = async (password: string) => {
    setLoginLoading(true);
    setLoginError('');
    try {
      setAuthenticated(await loginAndAuthenticate(login, password));
    } catch (error) {
      setLoginError(error instanceof Error ? error.message : 'Login failed');
    } finally {
      setLoginLoading(false);
    }
  };

  return <AppView authenticated={authenticated} loading={loginLoading} error={loginError} onLogin={handleLogin} />;
}

export default App;
