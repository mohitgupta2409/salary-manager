export interface Employee {
  id: number;
  full_name: string;
  email: string;
  job_title: string;
  department: string;
  country: string;
  salary: number;
  currency: string;
  join_date: string;
  created_at: string;
  updated_at: string;
}

export interface EmployeeFormData {
  full_name: string;
  email: string;
  job_title: string;
  department: string;
  country: string;
  salary: number;
  currency: string;
  join_date: string;
}

export interface EmployeeListResult {
  employees: Employee[];
  total: number;
  page: number;
  limit: number;
}

export interface EmployeeFilter {
  search?: string;
  country?: string;
  job_title?: string;
  page: number;
  limit: number;
}

export interface SalaryRange {
  country: string;
  min: number;
  max: number;
  average: number;
  median: number;
  count: number;
}

export interface SalaryByTitle {
  country: string;
  job_title: string;
  average: number;
  min: number;
  max: number;
  count: number;
}

export interface DepartmentStats {
  department: string;
  average_salary: number;
  min_salary: number;
  max_salary: number;
  employee_count: number;
}

export interface CountryHeadcount {
  country: string;
  employee_count: number;
  average_salary: number;
}

export interface OrgSummary {
  total_employees: number;
  average_salary: number;
  total_countries: number;
  total_departments: number;
  country_breakdown: CountryHeadcount[];
}
