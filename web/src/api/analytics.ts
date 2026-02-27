import client from './client';

export interface UsageSummary {
  total_tokens: number;
  total_cost_microcents: number;
  total_invocations: number;
  tool_calls_total: number;
}

export interface DailySummary {
  date: string;
  total_tokens: number;
  input_tokens: number;
  output_tokens: number;
  cost_microcents: number;
  invocations: number;
}

export interface ModelBreakdown {
  model: string;
  provider: string;
  total_tokens: number;
  cost_microcents: number;
  invocations: number;
}

export interface UsageRecord {
  id: string;
  user_id: string;
  project_id: string;
  agent_id?: string;
  provider: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  cost_microcents: number;
  tool_calls_count: number;
  created_at: string;
}

export interface CostBudget {
  id: string;
  project_id: string;
  monthly_limit_microcents: number;
  alert_threshold_pct: number;
  created_at: string;
  updated_at: string;
}

export const analyticsApi = {
  summary: (projectId: string, days = 30) =>
    client.get<{ summary: UsageSummary }>(`/projects/${projectId}/analytics/summary?days=${days}`),

  daily: (projectId: string, days = 30) =>
    client.get<{ daily: DailySummary[] }>(`/projects/${projectId}/analytics/daily?days=${days}`),

  models: (projectId: string, days = 30) =>
    client.get<{ models: ModelBreakdown[] }>(`/projects/${projectId}/analytics/models?days=${days}`),

  recent: (projectId: string, limit = 50) =>
    client.get<{ records: UsageRecord[] }>(`/projects/${projectId}/analytics/recent?limit=${limit}`),

  getBudget: (projectId: string) =>
    client.get<{ budget: CostBudget | null; current_spend: number; monthly_limit: number; over_budget: boolean }>(
      `/projects/${projectId}/analytics/budget`,
    ),

  setBudget: (projectId: string, monthlyLimitMicrocents: number, alertThresholdPct?: number) =>
    client.post<{ budget: CostBudget }>(`/projects/${projectId}/analytics/budget`, {
      monthly_limit_microcents: monthlyLimitMicrocents,
      alert_threshold_pct: alertThresholdPct || 80,
    }),
};
