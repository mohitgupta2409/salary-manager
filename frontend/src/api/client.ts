import type {
  Country,
  Department,
  JobTitle,
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
    if (filter.country_id) params.set('country_id', String(filter.country_id));
    if (filter.job_title_id) params.set('job_title_id', String(filter.job_title_id));
    if (filter.department_id) params.set('department_id', String(filter.department_id));
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

export const countryApi = {
  list(): Promise<Country[]> {
    return request(`${BASE}/countries`);
  },
};

export const departmentApi = {
  list(): Promise<Department[]> {
    return request(`${BASE}/departments`);
  },
};

export const jobTitleApi = {
  list(departmentId?: number): Promise<JobTitle[]> {
    const params = new URLSearchParams();
    if (departmentId) params.set('department_id', String(departmentId));
    const qs = params.toString();
    return request(`${BASE}/job-titles${qs ? '?' + qs : ''}`);
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
