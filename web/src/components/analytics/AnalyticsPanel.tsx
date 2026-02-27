import { useState, useEffect } from 'react';
import { BarChart3, DollarSign, Zap, Cpu, TrendingUp, Settings2 } from 'lucide-react';
import { analyticsApi, type UsageSummary, type DailySummary, type ModelBreakdown, type UsageRecord } from '../../api/analytics';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Card } from '../ui/Card';

interface AnalyticsPanelProps {
  projectId: string;
}

function formatCost(microcents: number): string {
  const dollars = microcents / 1_000_000;
  if (dollars < 0.01) return `$${(microcents / 10000).toFixed(4)}`;
  return `$${dollars.toFixed(2)}`;
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toString();
}

export function AnalyticsPanel({ projectId }: AnalyticsPanelProps) {
  const [tab, setTab] = useState<'overview' | 'models' | 'history' | 'budget'>('overview');
  const [days, setDays] = useState(30);
  const [summary, setSummary] = useState<UsageSummary | null>(null);
  const [daily, setDaily] = useState<DailySummary[]>([]);
  const [models, setModels] = useState<ModelBreakdown[]>([]);
  const [records, setRecords] = useState<UsageRecord[]>([]);
  const [budgetData, setBudgetData] = useState<{ current_spend: number; monthly_limit: number; over_budget: boolean } | null>(null);
  const [budgetInput, setBudgetInput] = useState('');
  const [alertPct, setAlertPct] = useState('80');

  useEffect(() => { loadAll(); }, [projectId, days]);

  const loadAll = async () => {
    try {
      const [s, d, m, r, b] = await Promise.all([
        analyticsApi.summary(projectId, days),
        analyticsApi.daily(projectId, days),
        analyticsApi.models(projectId, days),
        analyticsApi.recent(projectId, 30),
        analyticsApi.getBudget(projectId),
      ]);
      setSummary(s.data.summary);
      setDaily(d.data.daily || []);
      setModels(m.data.models || []);
      setRecords(r.data.records || []);
      setBudgetData({ current_spend: b.data.current_spend, monthly_limit: b.data.monthly_limit, over_budget: b.data.over_budget });
      if (b.data.budget) {
        setBudgetInput((b.data.budget.monthly_limit_microcents / 1_000_000).toString());
        setAlertPct(b.data.budget.alert_threshold_pct.toString());
      }
    } catch { /* ignore */ }
  };

  const handleSetBudget = async () => {
    const limit = Math.round(parseFloat(budgetInput || '0') * 1_000_000);
    await analyticsApi.setBudget(projectId, limit, parseInt(alertPct) || 80);
    loadAll();
  };

  const maxDailyTokens = Math.max(...daily.map((d) => d.total_tokens), 1);

  const tabs = [
    { id: 'overview' as const, label: 'Overview', icon: BarChart3 },
    { id: 'models' as const, label: 'Models', icon: Cpu },
    { id: 'history' as const, label: 'History', icon: TrendingUp },
    { id: 'budget' as const, label: 'Budget', icon: Settings2 },
  ];

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold flex items-center gap-2">
          <BarChart3 size={18} /> Analytics
        </h3>
        <select
          value={days}
          onChange={(e) => setDays(parseInt(e.target.value))}
          className="px-2 py-1 text-xs bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)]"
        >
          <option value={7}>7 days</option>
          <option value={14}>14 days</option>
          <option value={30}>30 days</option>
          <option value={90}>90 days</option>
        </select>
      </div>

      {/* Sub-tabs */}
      <div className="flex gap-1 bg-[var(--color-bg)] rounded-lg p-1">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex-1 text-xs py-1.5 px-2 rounded-md transition-colors flex items-center justify-center gap-1 ${
              tab === t.id ? 'bg-[var(--color-bg-tertiary)] text-[var(--color-text)]' : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
            }`}
          >
            <t.icon size={12} /> {t.label}
          </button>
        ))}
      </div>

      {/* Overview */}
      {tab === 'overview' && summary && (
        <div className="space-y-4">
          {/* Stat Cards */}
          <div className="grid grid-cols-2 gap-3">
            <Card>
              <div className="flex items-center gap-2 mb-1">
                <Zap size={14} className="text-yellow-400" />
                <span className="text-[10px] text-[var(--color-text-secondary)] uppercase">Tokens Used</span>
              </div>
              <div className="text-lg font-bold">{formatTokens(summary.total_tokens)}</div>
            </Card>
            <Card>
              <div className="flex items-center gap-2 mb-1">
                <DollarSign size={14} className="text-green-400" />
                <span className="text-[10px] text-[var(--color-text-secondary)] uppercase">Est. Cost</span>
              </div>
              <div className="text-lg font-bold">{formatCost(summary.total_cost_microcents)}</div>
            </Card>
            <Card>
              <div className="flex items-center gap-2 mb-1">
                <BarChart3 size={14} className="text-blue-400" />
                <span className="text-[10px] text-[var(--color-text-secondary)] uppercase">Invocations</span>
              </div>
              <div className="text-lg font-bold">{summary.total_invocations}</div>
            </Card>
            <Card>
              <div className="flex items-center gap-2 mb-1">
                <Cpu size={14} className="text-purple-400" />
                <span className="text-[10px] text-[var(--color-text-secondary)] uppercase">Tool Calls</span>
              </div>
              <div className="text-lg font-bold">{summary.tool_calls_total}</div>
            </Card>
          </div>

          {/* Daily Usage Bar Chart */}
          {daily.length > 0 && (
            <Card>
              <p className="text-xs text-[var(--color-text-secondary)] mb-3">Daily Token Usage</p>
              <div className="flex items-end gap-[2px] h-32">
                {daily.map((d, i) => {
                  const pct = (d.total_tokens / maxDailyTokens) * 100;
                  return (
                    <div
                      key={i}
                      className="flex-1 bg-blue-500/60 rounded-t hover:bg-blue-500/80 transition-colors relative group"
                      style={{ height: `${Math.max(pct, 2)}%` }}
                      title={`${d.date}: ${formatTokens(d.total_tokens)} tokens (${formatCost(d.cost_microcents)})`}
                    >
                      <div className="absolute bottom-full left-1/2 -translate-x-1/2 hidden group-hover:block bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] rounded px-2 py-1 text-[10px] whitespace-nowrap z-10 mb-1">
                        {d.date}<br />
                        {formatTokens(d.total_tokens)} tok &middot; {formatCost(d.cost_microcents)}
                      </div>
                    </div>
                  );
                })}
              </div>
            </Card>
          )}

          {/* Budget Alert */}
          {budgetData && budgetData.monthly_limit > 0 && (
            <Card>
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-[var(--color-text-secondary)]">Monthly Budget</span>
                <span className={`text-xs font-medium ${budgetData.over_budget ? 'text-red-400' : 'text-green-400'}`}>
                  {formatCost(budgetData.current_spend)} / {formatCost(budgetData.monthly_limit)}
                </span>
              </div>
              <div className="w-full h-2 bg-[var(--color-bg)] rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all ${budgetData.over_budget ? 'bg-red-500' : 'bg-green-500'}`}
                  style={{ width: `${Math.min((budgetData.current_spend / budgetData.monthly_limit) * 100, 100)}%` }}
                />
              </div>
            </Card>
          )}
        </div>
      )}

      {/* Models Tab */}
      {tab === 'models' && (
        <div className="space-y-3">
          {models.length === 0 ? (
            <div className="text-center py-8 text-[var(--color-text-secondary)] text-sm">No model usage data yet.</div>
          ) : (
            models.map((m, i) => (
              <Card key={i}>
                <div className="flex items-center justify-between">
                  <div>
                    <span className="text-sm font-medium">{m.model}</span>
                    <span className="text-[10px] text-[var(--color-text-secondary)] ml-2">{m.provider}</span>
                  </div>
                  <span className="text-sm font-mono text-green-400">{formatCost(m.cost_microcents)}</span>
                </div>
                <div className="flex gap-4 mt-1 text-[10px] text-[var(--color-text-secondary)]">
                  <span>{formatTokens(m.total_tokens)} tokens</span>
                  <span>{m.invocations} calls</span>
                </div>
                {/* proportion bar */}
                <div className="w-full h-1.5 bg-[var(--color-bg)] rounded-full mt-2 overflow-hidden">
                  <div
                    className="h-full bg-blue-500/60 rounded-full"
                    style={{ width: `${Math.max((m.cost_microcents / Math.max(models[0].cost_microcents, 1)) * 100, 3)}%` }}
                  />
                </div>
              </Card>
            ))
          )}
        </div>
      )}

      {/* History Tab */}
      {tab === 'history' && (
        <div className="space-y-2">
          {records.length === 0 ? (
            <div className="text-center py-8 text-[var(--color-text-secondary)] text-sm">No usage records yet.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-[var(--color-text-secondary)] border-b border-[var(--color-border)]">
                    <th className="text-left pb-2">Model</th>
                    <th className="text-right pb-2">In</th>
                    <th className="text-right pb-2">Out</th>
                    <th className="text-right pb-2">Cost</th>
                    <th className="text-right pb-2">Tools</th>
                    <th className="text-right pb-2">Time</th>
                  </tr>
                </thead>
                <tbody>
                  {records.map((rec) => (
                    <tr key={rec.id} className="border-b border-[var(--color-border)]">
                      <td className="py-1.5 font-mono">{rec.model}</td>
                      <td className="text-right py-1.5">{formatTokens(rec.input_tokens)}</td>
                      <td className="text-right py-1.5">{formatTokens(rec.output_tokens)}</td>
                      <td className="text-right py-1.5 text-green-400">{formatCost(rec.cost_microcents)}</td>
                      <td className="text-right py-1.5">{rec.tool_calls_count || '-'}</td>
                      <td className="text-right py-1.5 text-[var(--color-text-secondary)]">
                        {new Date(rec.created_at).toLocaleTimeString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Budget Tab */}
      {tab === 'budget' && (
        <div className="space-y-4">
          <Card>
            <h4 className="text-sm font-medium mb-3">Monthly Cost Budget</h4>
            <div className="space-y-3">
              <Input
                label="Monthly Limit ($)"
                type="number"
                value={budgetInput}
                onChange={(e) => setBudgetInput(e.target.value)}
                placeholder="10.00"
              />
              <Input
                label="Alert Threshold (%)"
                type="number"
                value={alertPct}
                onChange={(e) => setAlertPct(e.target.value)}
                placeholder="80"
              />
              <Button onClick={handleSetBudget} size="sm" className="w-full">
                Save Budget
              </Button>
            </div>
          </Card>

          {budgetData && budgetData.monthly_limit > 0 && (
            <Card>
              <h4 className="text-sm font-medium mb-2">Current Month</h4>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">Spent</span>
                  <span className="font-mono">{formatCost(budgetData.current_spend)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">Limit</span>
                  <span className="font-mono">{formatCost(budgetData.monthly_limit)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">Remaining</span>
                  <span className={`font-mono ${budgetData.over_budget ? 'text-red-400' : 'text-green-400'}`}>
                    {formatCost(Math.max(budgetData.monthly_limit - budgetData.current_spend, 0))}
                  </span>
                </div>
              </div>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}
