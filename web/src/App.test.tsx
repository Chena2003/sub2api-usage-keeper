import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it, vi } from 'vitest';
import App from './App';

vi.mock('./pages/UsagePage', () => ({
  UsagePage: () => <main>UsagePage mock</main>,
}));

describe('App', () => {
  it('renders UsagePage directly', () => {
    const html = renderToStaticMarkup(<App />);
    expect(html).toContain('UsagePage mock');
  });
});
