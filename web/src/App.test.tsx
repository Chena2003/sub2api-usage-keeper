import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it, vi } from 'vitest';
import App, { AppView, checkSessionAuthenticated, loginAndAuthenticate } from './App';

vi.mock('./pages/UsagePage', () => ({
  UsagePage: () => <main>UsagePage mock</main>,
}));

vi.mock('./pages/LoginPage', () => ({
  LoginPage: () => <main>LoginPage mock</main>,
}));

describe('App', () => {
  it('keeps the default server-rendered shell empty until the session check completes', () => {
    expect(renderToStaticMarkup(<App />)).toBe('');
  });

  it('renders UsagePage when the session is authenticated', () => {
    const html = renderToStaticMarkup(<AppView authenticated={true} onLogin={async () => {}} />);

    expect(html).toContain('UsagePage mock');
    expect(html).not.toContain('LoginPage mock');
  });

  it('renders LoginPage when backend auth is enabled and the session is unauthenticated', () => {
    const html = renderToStaticMarkup(<AppView authenticated={false} onLogin={async () => {}} />);

    expect(html).toContain('LoginPage mock');
    expect(html).not.toContain('UsagePage mock');
  });

  it('treats the no-password backend session as authenticated', async () => {
    await expect(checkSessionAuthenticated(async () => ({ authenticated: true }))).resolves.toBe(true);
  });

  it('marks the UI authenticated after a successful login', async () => {
    const login = vi.fn().mockResolvedValue(undefined);

    await expect(loginAndAuthenticate(login, 'secret')).resolves.toBe(true);
    expect(login).toHaveBeenCalledWith('secret');
  });
});
