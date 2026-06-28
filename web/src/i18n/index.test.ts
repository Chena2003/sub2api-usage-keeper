import { describe, expect, it } from 'vitest';
import i18n, { SUPPORTED_LANGUAGES } from './index';

const flattenKeys = (value: unknown, prefix = ''): string[] => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return [prefix];
  return Object.entries(value).flatMap(([key, child]) => {
    const path = prefix ? `${prefix}.${key}` : key;
    return flattenKeys(child, path);
  });
};

const sub2ApiRenderedKeys = [
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
  'usage_stats.sub2api_account',
  'usage_stats.sub2api_active_users',
  'usage_stats.sub2api_api_key',
  'usage_stats.sub2api_avg_duration',
  'usage_stats.sub2api_status_active',
  'usage_stats.sub2api_status_paused',
  'usage_stats.sub2api_status_inactive',
  'usage_stats.sub2api_status_rate_limited',
  'usage_stats.sub2api_status_temp_unschedulable',
  'usage_stats.sub2api_status_overloaded',
  'usage_stats.sub2api_status_error',
  'usage_stats.sub2api_cache_tokens',
  'usage_stats.sub2api_consumed',
  'usage_stats.sub2api_cost',
  'usage_stats.sub2api_dimension_user',
  'usage_stats.sub2api_dimension_api_key',
  'usage_stats.sub2api_dimension_model',
  'usage_stats.sub2api_dimension_account',
  'usage_stats.sub2api_events_eyebrow',
  'usage_stats.sub2api_events_hint',
  'usage_stats.sub2api_events_title',
  'usage_stats.sub2api_five_hour_window',
  'usage_stats.sub2api_input_tokens',
  'usage_stats.sub2api_last_used',
  'usage_stats.sub2api_model',
  'usage_stats.sub2api_models_eyebrow',
  'usage_stats.sub2api_name',
  'usage_stats.sub2api_no_data',
  'usage_stats.sub2api_no_quota_window',
  'usage_stats.sub2api_output_tokens',
  'usage_stats.sub2api_overview_title',
  'usage_stats.sub2api_ratio',
  'usage_stats.sub2api_plan',
  'usage_stats.sub2api_provider',
  'usage_stats.sub2api_available_accounts',
  'usage_stats.sub2api_quotas_eyebrow',
  'usage_stats.sub2api_quotas_hint',
  'usage_stats.sub2api_quotas_title',
  'usage_stats.sub2api_rank',
  'usage_stats.sub2api_ranking_dimension',
  'usage_stats.sub2api_ranking_eyebrow',
  'usage_stats.sub2api_ranking_hint',
  'usage_stats.sub2api_ranking_title',
  'usage_stats.sub2api_refresh_at',
  'usage_stats.sub2api_requests',
  'usage_stats.sub2api_share',
  'usage_stats.sub2api_status',
  'usage_stats.sub2api_timestamp',
  'usage_stats.sub2api_tokens',
  'usage_stats.sub2api_top_models',
  'usage_stats.sub2api_top_models_hint',
  'usage_stats.sub2api_total_cost',
  'usage_stats.sub2api_total_requests',
  'usage_stats.sub2api_total_tokens',
  'usage_stats.sub2api_user',
  'usage_stats.sub2api_weekly_window',
];

describe('i18n resources', () => {
  it('keeps every supported language aligned with English keys', () => {
    const englishKeys = flattenKeys(i18n.getResourceBundle('en', 'translation')).sort();

    for (const language of SUPPORTED_LANGUAGES) {
      expect(flattenKeys(i18n.getResourceBundle(language, 'translation')).sort()).toEqual(englishKeys);
    }
  });

  it('defines Sub2API redesign keys in every language', async () => {
    for (const language of SUPPORTED_LANGUAGES) {
      await i18n.changeLanguage(language);

      for (const key of sub2ApiRenderedKeys) {
        expect(i18n.t(key)).not.toBe(key);
      }
    }
  });

  it('resolves rendered Sub2API available accounts label in every language', async () => {
    for (const language of SUPPORTED_LANGUAGES) {
      await i18n.changeLanguage(language);

      const text = i18n.t('usage_stats.sub2api_available_accounts');

      expect(text).not.toBe('usage_stats.sub2api_available_accounts');
      expect(text.length).toBeGreaterThan(0);
    }
  });
});
