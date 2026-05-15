import { useState, useEffect, useCallback } from 'react';
import { Search, Plus, Pencil, Trash2, ChevronLeft, ChevronRight } from 'lucide-react';
import { employeeApi } from '../api/client';
import { useReferenceData } from '../api/useReferenceData';
import type { Employee, EmployeeListResult } from '../types/employee';
import EmployeeModal from '../components/EmployeeModal';

const PAGE_SIZE = 20;

export default function EmployeesPage() {
  const ref = useReferenceData();

  const [data, setData] = useState<EmployeeListResult | null>(null);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [countryID, setCountryID] = useState<number>(0);
  const [departmentID, setDepartmentID] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<number | null>(null);
  const [error, setError] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await employeeApi.list({
        page, limit: PAGE_SIZE, search,
        country_id: countryID || undefined,
        department_id: departmentID || undefined,
      });
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load employees');
    } finally {
      setLoading(false);
    }
  }, [page, search, countryID, departmentID]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    setPage(1);
  }, [search, countryID, departmentID]);

  const handleCreate = async (formData: Parameters<typeof employeeApi.create>[0]) => {
    await employeeApi.create(formData);
    fetchData();
  };

  const handleUpdate = async (formData: Parameters<typeof employeeApi.create>[0]) => {
    if (!editingEmployee) return;
    await employeeApi.update(editingEmployee.id, formData);
    fetchData();
  };

  const handleDelete = async (id: number) => {
    await employeeApi.delete(id);
    setDeleteConfirm(null);
    fetchData();
  };

  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 0;

  const formatSalary = (salary: number, currency?: string) => {
    if (!currency) return salary.toLocaleString();
    try {
      return new Intl.NumberFormat('en-US', {
        style: 'currency', currency, maximumFractionDigits: 0,
      }).format(salary);
    } catch {
      return `${currency} ${salary.toLocaleString()}`;
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Employees</h1>
          <p className="text-sm text-gray-500 mt-1">
            {data ? `${data.total.toLocaleString()} active employees` : 'Loading...'}
          </p>
        </div>
        <button
          onClick={() => { setEditingEmployee(null); setModalOpen(true); }}
          disabled={ref.loading}
          className="flex items-center gap-2 px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm font-medium shadow-sm disabled:opacity-50">
          <Plus className="w-4 h-4" /> Add Employee
        </button>
      </div>

      {ref.error && (
        <div className="bg-red-50 text-red-700 text-sm px-4 py-3 rounded-lg mb-4">
          Failed to load reference data: {ref.error}
        </div>
      )}
      {error && (
        <div className="bg-red-50 text-red-700 text-sm px-4 py-3 rounded-lg mb-4">{error}</div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
        <div className="flex flex-wrap gap-3">
          <div className="relative flex-1 min-w-[200px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search by name or email..."
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 outline-none"
            />
          </div>
          <select value={countryID} onChange={e => setCountryID(Number(e.target.value))}
            className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 outline-none">
            <option value={0}>All Countries</option>
            {ref.countries.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
          <select value={departmentID} onChange={e => setDepartmentID(Number(e.target.value))}
            className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 outline-none">
            <option value={0}>All Departments</option>
            {ref.departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="text-left px-4 py-3 font-medium text-gray-600">Name</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Email</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Job Title</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Department</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Country</th>
                <th className="text-right px-4 py-3 font-medium text-gray-600">Salary</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Joined</th>
                <th className="text-right px-4 py-3 font-medium text-gray-600">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {loading && (
                <tr><td colSpan={8} className="px-4 py-12 text-center text-gray-400">Loading...</td></tr>
              )}
              {!loading && data?.employees?.length === 0 && (
                <tr><td colSpan={8} className="px-4 py-12 text-center text-gray-400">No employees found</td></tr>
              )}
              {!loading && data?.employees?.map(emp => (
                <tr key={emp.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 font-medium text-gray-900">{emp.full_name}</td>
                  <td className="px-4 py-3 text-gray-500">{emp.email}</td>
                  <td className="px-4 py-3 text-gray-700">{emp.job_title}</td>
                  <td className="px-4 py-3">
                    <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-700">
                      {emp.department}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-700">{emp.country}</td>
                  <td className="px-4 py-3 text-right font-mono text-gray-900">
                    {formatSalary(emp.salary, emp.currency)}
                  </td>
                  <td className="px-4 py-3 text-gray-500">
                    {new Date(emp.join_date).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex justify-end gap-1">
                      <button onClick={() => { setEditingEmployee(emp); setModalOpen(true); }}
                        className="p-1.5 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                        title="Edit">
                        <Pencil className="w-4 h-4" />
                      </button>
                      <button onClick={() => setDeleteConfirm(emp.id)}
                        className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                        title="Delete">
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200 bg-gray-50">
            <p className="text-sm text-gray-600">Page {page} of {totalPages}</p>
            <div className="flex gap-1">
              <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}
                className="p-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-white disabled:opacity-40 disabled:cursor-not-allowed">
                <ChevronLeft className="w-4 h-4" />
              </button>
              <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}
                className="p-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-white disabled:opacity-40 disabled:cursor-not-allowed">
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>

      {modalOpen && (
        <EmployeeModal
          employee={editingEmployee}
          countries={ref.countries}
          departments={ref.departments}
          jobTitles={ref.jobTitles}
          onSave={editingEmployee ? handleUpdate : handleCreate}
          onClose={() => { setModalOpen(false); setEditingEmployee(null); }}
        />
      )}

      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
          <div className="bg-white rounded-xl shadow-2xl p-6 max-w-sm mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Delete Employee</h3>
            <p className="text-sm text-gray-600 mb-6">
              This will mark the employee as inactive. They will no longer appear in lists or analytics, but their record is preserved.
            </p>
            <div className="flex justify-end gap-3">
              <button onClick={() => setDeleteConfirm(null)}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">
                Cancel
              </button>
              <button onClick={() => handleDelete(deleteConfirm)}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700">
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
