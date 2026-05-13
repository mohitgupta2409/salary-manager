import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import type { Employee, EmployeeFormData } from '../types/employee';

interface Props {
  employee?: Employee | null;
  onSave: (data: EmployeeFormData) => Promise<void>;
  onClose: () => void;
}

const COUNTRIES = [
  'United States', 'India', 'United Kingdom', 'Germany', 'Canada',
  'Australia', 'France', 'Japan', 'Singapore',
];

const CURRENCY_MAP: Record<string, string> = {
  'United States': 'USD', 'India': 'INR', 'United Kingdom': 'GBP',
  'Germany': 'EUR', 'Canada': 'CAD', 'Australia': 'AUD',
  'France': 'EUR', 'Japan': 'JPY', 'Singapore': 'SGD',
};

const DEPARTMENTS = [
  'Engineering', 'Product', 'Design', 'Data Science', 'Marketing',
  'Sales', 'Human Resources', 'Finance', 'Operations',
  'Customer Success', 'Legal', 'IT Infrastructure',
];

const initial: EmployeeFormData = {
  full_name: '', email: '', job_title: '', department: '',
  country: '', salary: 0, currency: 'USD', join_date: '',
};

export default function EmployeeModal({ employee, onSave, onClose }: Props) {
  const [form, setForm] = useState<EmployeeFormData>(initial);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (employee) {
      setForm({
        full_name: employee.full_name,
        email: employee.email,
        job_title: employee.job_title,
        department: employee.department,
        country: employee.country,
        salary: employee.salary,
        currency: employee.currency,
        join_date: employee.join_date.split('T')[0],
      });
    }
  }, [employee]);

  const handleCountryChange = (country: string) => {
    setForm(f => ({ ...f, country, currency: CURRENCY_MAP[country] || f.currency }));
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

  const inputClass = 'block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 outline-none';
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
              <label className={labelClass}>Full Name *</label>
              <input className={inputClass} value={form.full_name}
                onChange={e => setForm(f => ({ ...f, full_name: e.target.value }))}
                required />
            </div>
            <div>
              <label className={labelClass}>Email *</label>
              <input className={inputClass} type="email" value={form.email}
                onChange={e => setForm(f => ({ ...f, email: e.target.value }))}
                required />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Job Title *</label>
              <input className={inputClass} value={form.job_title}
                onChange={e => setForm(f => ({ ...f, job_title: e.target.value }))}
                required />
            </div>
            <div>
              <label className={labelClass}>Department *</label>
              <select className={inputClass} value={form.department}
                onChange={e => setForm(f => ({ ...f, department: e.target.value }))}
                required>
                <option value="">Select...</option>
                {DEPARTMENTS.map(d => <option key={d} value={d}>{d}</option>)}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Country *</label>
              <select className={inputClass} value={form.country}
                onChange={e => handleCountryChange(e.target.value)}
                required>
                <option value="">Select...</option>
                {COUNTRIES.map(c => <option key={c} value={c}>{c}</option>)}
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
              <input className={inputClass} value={form.currency} readOnly
                title="Auto-set based on country" />
            </div>
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
