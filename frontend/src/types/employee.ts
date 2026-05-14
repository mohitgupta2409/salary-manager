export interface Country {
  id: number;
  name: string;
  code: string;
  currency: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Department {
  id: number;
  name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface JobTitle {
  id: number;
  name: string;
  department_id: number;
  department?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Employee {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  job_title_id: number;
  country_id: number;
  salary: number;
  address?: string;
  join_date: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;

  // Denormalized fields populated by the backend via JOINs
  job_title?: string;
  department?: string;
  country?: string;
  currency?: string;
}

export interface EmployeeFormData {
  first_name: string;
  last_name: string;
  email: string;
  job_title_id: number;
  country_id: number;
  salary: number;
  address: string;
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
  country_id?: number;
  job_title_id?: number;
  department_id?: number;
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
