const API_BASE = "/api";

export interface Probe {
  id: string;
  type: string;
  target: string;
  file_path: string;
  status: string;
  created_at: string;
}

export interface Provider {
  name: string;
  base_url: string;
  api_key: string;
  models: string[];
  default_model: string;
}

export interface Config {
  providers: Record<string, Provider>;
  default: string;
}

export interface Finding {
  id: string;
  probe_id: string;
  text: string;
  severity: string;
  completed: boolean;
}

async function request<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

// Probes
export const getProbes = () => request<Probe[]>("/probes");
export const getProbe = (id: string) => request<Probe>(`/probes/${id}`);

// Config
export const getConfig = () => request<Config>("/config");
export const updateConfig = (config: Config) =>
  request<Config>("/config", {
    method: "PUT",
    body: JSON.stringify(config),
  });

// Findings
export const toggleFinding = (id: string) =>
  request<Finding>(`/findings/${id}`, { method: "PATCH" });

// File tree (placeholder)
export const getFileTree = (id: string) =>
  request<string[]>(`/file-tree/${id}`);
