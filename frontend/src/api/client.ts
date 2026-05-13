import type {
  Employee,
  EmployeeFormData,
  EmployeeListResult,
  EmployeeFilter,
  SalaryRange,
  SalaryByTitle,
  DepartmentStats,
  OrgSummary,
} from '../types/employee';

const BASE = '/api';

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

export const employeeApi = {
  list(filter: EmployeeFilter): Promise<EmployeeListResult> {
    const params = new URLSearchParams();
    params.set('page', String(filter.page));
    params.set('limit', String(filter.limit));
    if (filter.search) params.set('search', filter.search);
    if (filter.country) params.set('country', filter.country);
    if (filter.job_title) params.set('job_title', filter.job_title);
    return request(`${BASE}/employees?${params}`);
  },

  getById(id: number): Promise<Employee> {
    return request(`${BASE}/employees/${id}`);
  },

  create(data: EmployeeFormData): Promise<Employee> {
    return request(`${BASE}/employees`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  update(id: number, data: EmployeeFormData): Promise<Employee> {
    return request(`${BASE}/employees/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  delete(id: number): Promise<void> {
    return request(`${BASE}/employees/${id}`, { method: 'DELETE' });
  },
};

export const insightsApi = {
  salaryRange(country: string): Promise<SalaryRange> {
    return request(`${BASE}/insights/salary-range?country=${encodeURIComponent(country)}`);
  },

  salaryByTitle(country: string, jobTitle: string): Promise<SalaryByTitle> {
    const params = new URLSearchParams({ country, job_title: jobTitle });
    return request(`${BASE}/insights/salary-by-title?${params}`);
  },

  departmentStats(country?: string): Promise<DepartmentStats[]> {
    const params = country ? `?country=${encodeURIComponent(country)}` : '';
    return request(`${BASE}/insights/department-stats${params}`);
  },

  summary(): Promise<OrgSummary> {
    return request(`${BASE}/insights/summary`);
  },
};
