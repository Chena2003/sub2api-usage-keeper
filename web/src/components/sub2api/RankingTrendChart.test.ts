import { describe, expect, it } from 'vitest';
import { assignSeriesColors } from './RankingTrendChart';

describe('assignSeriesColors', () => {
  it('assigns the same color to a name regardless of input order', () => {
    const names = ['alice', 'bob', 'carol', 'dave'];
    const forward = assignSeriesColors(names);
    const shuffled = assignSeriesColors([...names].reverse());

    for (const name of names) {
      expect(shuffled.get(name)).toBe(forward.get(name));
    }
  });

  it('keeps colors stable when the set of names is re-ranked (simulating a backend rank swap)', () => {
    const before = assignSeriesColors(['alice', 'bob', 'carol']);
    // 'bob' and 'carol' swapped rank (i.e. insertion order), names unchanged.
    const after = assignSeriesColors(['alice', 'carol', 'bob']);

    expect(after.get('alice')).toBe(before.get('alice'));
    expect(after.get('bob')).toBe(before.get('bob'));
    expect(after.get('carol')).toBe(before.get('carol'));
  });

  it('assigns distinct colors to distinct names within one dataset', () => {
    const names = ['alice', 'bob', 'carol', 'dave', 'erin'];
    const assignments = assignSeriesColors(names);
    const colors = [...assignments.values()];

    expect(new Set(colors).size).toBe(colors.length);
  });

  it('falls back to the next free color on a hash collision', () => {
    const colors = ['#000000', '#111111'];
    // Force both names to hash to the same index by using a tiny palette.
    const assignments = assignSeriesColors(['name-a', 'name-b'], colors);

    const assigned = [...assignments.values()];
    expect(new Set(assigned).size).toBe(2);
    expect(assigned.every((color) => colors.includes(color))).toBe(true);
  });
});
