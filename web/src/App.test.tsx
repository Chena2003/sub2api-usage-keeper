import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it, vi } from 'vitest';
import App from './App';

vi.mock('./pages/UsagePage', () => ({
  UsagePage: () => <main>UsagePage mock</main>,
}));

vi.mock('./pages/LoginPage', () => ({
  LoginPage: () => <main>LoginPage mock</main>,
}));

describe('App', () => {
  it('renders UsagePage without LoginPage', () => {
    const html = renderToStaticMarkup(<App />);

    expect(html).toContain('UsagePage mock');
    expect(html).not.toContain('LoginPage mock');
  });
});
