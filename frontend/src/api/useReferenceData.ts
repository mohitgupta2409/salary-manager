import { useState, useEffect } from 'react';
import { countryApi, departmentApi, jobTitleApi } from './client';
import type { Country, Department, JobTitle } from '../types/employee';

export interface ReferenceData {
  countries: Country[];
  departments: Department[];
  jobTitles: JobTitle[];
  loading: boolean;
  error: string | null;
  reload: () => void;
}

export function useReferenceData(): ReferenceData {
  const [countries, setCountries] = useState<Country[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [jobTitles, setJobTitles] = useState<JobTitle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [version, setVersion] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    Promise.all([countryApi.list(), departmentApi.list(), jobTitleApi.list()])
      .then(([c, d, j]) => {
        if (cancelled) return;
        setCountries(c);
        setDepartments(d);
        setJobTitles(j);
      })
      .catch(err => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : 'Failed to load reference data');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [version]);

  return {
    countries,
    departments,
    jobTitles,
    loading,
    error,
    reload: () => setVersion(v => v + 1),
  };
}
