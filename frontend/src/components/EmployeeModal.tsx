import { useState, useEffect, useMemo } from 'react';
import { X } from 'lucide-react';
import type { Country, Department, Employee, EmployeeFormData, JobTitle } from '../types/employee';

interface Props {
  employee?: Employee | null;
  countries: Country[];
  departments: Department[];
  jobTitles: JobTitle[];
  onSave: (data: EmployeeFormData) => Promise<void>;
  onClose: () => void;
}

const initial: EmployeeFormData = {
  first_name: '',
  last_name: '',
  email: '',
  job_title_id: 0,
  country_id: 0,
  salary: 0,
  address: '',
  join_date: '',
};

export default function EmployeeModal({
  employee, countries, departments, jobTitles, onSave, onClose,
}: Props) {
  const [form, setForm] = useState<EmployeeFormData>(initial);
  const [departmentID, setDepartmentID] = useState<number>(0);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  // Filter job titles by selected department
  const filteredJobTitles = useMemo(() => {
    if (!departmentID) return jobTitles;
    return jobTitles.filter(jt => jt.department_id === departmentID);
  }, [jobTitles, departmentID]);

  // Compute currency from selected country (for display only)
  const selectedCurrency = useMemo(() => {
    const c = countries.find(c => c.id === form.country_id);
    return c?.currency || '';
  }, [countries, form.country_id]);

  useEffect(() => {
    if (employee) {
      // The API returns full_name (joined server-side) and looks up
      // job_title/country by name; resolve back to the FK ids the form
      // needs by matching against the reference data. Splitting on the
      // first space keeps multi-word last names intact.
      const [firstName = '', ...lastParts] = (employee.full_name || '').split(' ');
      const jt = jobTitles.find(j => j.name === employee.job_title);
      const country = countries.find(c => c.name === employee.country);

      setForm({
        first_name: firstName,
        last_name: lastParts.join(' '),
        email: employee.email,
        job_title_id: jt?.id ?? 0,
        country_id: country?.id ?? 0,
        salary: employee.salary,
        address: employee.address || '',
        join_date: employee.join_date.split('T')[0],
      });
      if (jt) setDepartmentID(jt.department_id);
    }
  }, [employee, jobTitles, countries]);

  const handleDepartmentChange = (deptID: number) => {
    setDepartmentID(deptID);
    // Clear job_title_id if it's no longer valid for the new department
    const stillValid = jobTitles.some(jt => jt.id === form.job_title_id && jt.department_id === deptID);
    if (!stillValid) {
      setForm(f => ({ ...f, job_title_id: 0 }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSaving(true);
    try {
      await onSave({ ...form, join_date: form.join_date + 'T00:00:00Z' });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const inputClass = 'block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 outline-none disabled:bg-gray-50 disabled:text-gray-400';
  const labelClass = 'block text-sm font-medium text-gray-700 mb-1';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between p-5 border-b">
          <h2 className="text-lg font-semibold text-gray-900">
            {employee ? 'Edit Employee' : 'Add Employee'}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {error && (
            <div className="bg-red-50 text-red-700 text-sm px-4 py-3 rounded-lg">{error}</div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>First Name *</label>
              <input className={inputClass} value={form.first_name}
                onChange={e => setForm(f => ({ ...f, first_name: e.target.value }))}
                required />
            </div>
            <div>
              <label className={labelClass}>Last Name *</label>
              <input className={inputClass} value={form.last_name}
                onChange={e => setForm(f => ({ ...f, last_name: e.target.value }))}
                required />
            </div>
          </div>

          <div>
            <label className={labelClass}>Email *</label>
            <input className={inputClass} type="email" value={form.email}
              onChange={e => setForm(f => ({ ...f, email: e.target.value }))}
              required />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Department *</label>
              <select className={inputClass} value={departmentID || ''}
                onChange={e => handleDepartmentChange(Number(e.target.value))}
                required>
                <option value="">Select department...</option>
                {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div>
              <label className={labelClass}>Job Title *</label>
              <select className={inputClass} value={form.job_title_id || ''}
                onChange={e => setForm(f => ({ ...f, job_title_id: Number(e.target.value) }))}
                disabled={!departmentID}
                required>
                <option value="">{departmentID ? 'Select title...' : 'Select dept first'}</option>
                {filteredJobTitles.map(jt => <option key={jt.id} value={jt.id}>{jt.name}</option>)}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Country *</label>
              <select className={inputClass} value={form.country_id || ''}
                onChange={e => setForm(f => ({ ...f, country_id: Number(e.target.value) }))}
                required>
                <option value="">Select country...</option>
                {countries.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div>
              <label className={labelClass}>Join Date *</label>
              <input className={inputClass} type="date" value={form.join_date}
                onChange={e => setForm(f => ({ ...f, join_date: e.target.value }))}
                required />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Salary *</label>
              <input className={inputClass} type="number" min="0" step="100"
                value={form.salary || ''}
                onChange={e => setForm(f => ({ ...f, salary: Number(e.target.value) }))}
                required />
            </div>
            <div>
              <label className={labelClass}>Currency</label>
              <input className={inputClass} value={selectedCurrency} disabled
                placeholder="Auto from country"
                title="Currency is derived from the selected country" />
            </div>
          </div>

          <div>
            <label className={labelClass}>Address</label>
            <input className={inputClass} value={form.address}
              onChange={e => setForm(f => ({ ...f, address: e.target.value }))}
              placeholder="Optional" />
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t">
            <button type="button" onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">
              Cancel
            </button>
            <button type="submit" disabled={saving}
              className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 disabled:opacity-50">
              {saving ? 'Saving...' : employee ? 'Update' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
