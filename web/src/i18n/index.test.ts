import { describe, expect, it } from 'vitest';
import i18n, { SUPPORTED_LANGUAGES } from './index';

const flattenKeys = (value: unknown, prefix = ''): string[] => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return [prefix];
  return Object.entries(value).flatMap(([key, child]) => {
    const path = prefix ? `${prefix}.${key}` : key;
    return flattenKeys(child, path);
  });
};

describe('i18n resources', () => {
  it('keeps every supported language aligned with English keys', () => {
    const englishKeys = flattenKeys(i18n.getResourceBundle('en', 'translation')).sort();

    for (const language of SUPPORTED_LANGUAGES) {
      expect(flattenKeys(i18n.getResourceBundle(language, 'translation')).sort()).toEqual(englishKeys);
    }
  });

  it('defines Sub2API redesign keys in every language', async () => {
    const keys = [
      'usage_stats.tab_ranking',
      'usage_stats.tab_quotas',
      'usage_stats.active_users',
      'usage_stats.quota_risk_summary',
      'usage_stats.quota_risk_count',
      'usage_stats.ranking_dimension',
      'usage_stats.ranking_user',
      'usage_stats.ranking_api_key',
      'usage_stats.ranking_model',
      'usage_stats.ranking_account',
      'usage_stats.ranking_name',
      'usage_stats.ranking_share',
      'usage_stats.account',
      'usage_stats.status',
      'usage_stats.five_hour_window',
      'usage_stats.weekly_window',
      'usage_stats.refresh_time',
      'usage_stats.unknown',
      'usage_stats.user',
      'usage_stats.api_key',
    ];

    for (const language of SUPPORTED_LANGUAGES) {
      await i18n.changeLanguage(language);

      for (const key of keys) {
        expect(i18n.t(key)).not.toBe(key);
      }
    }
  });
});
