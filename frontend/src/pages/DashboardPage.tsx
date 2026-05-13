import { useState, useEffect } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, Legend,
} from 'recharts';
import { Users, Globe, Building2, DollarSign } from 'lucide-react';
import { insightsApi } from '../api/client';
import type { OrgSummary, DepartmentStats, SalaryRange } from '../types/employee';

const COLORS = ['#4f46e5', '#06b6d4', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316'];

function formatCompact(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toFixed(0);
}

export default function DashboardPage() {
  const [summary, setSummary] = useState<OrgSummary | null>(null);
  const [deptStats, setDeptStats] = useState<DepartmentStats[]>([]);
  const [selectedCountry, setSelectedCountry] = useState('');
  const [countryRange, setCountryRange] = useState<SalaryRange | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const [s, d] = await Promise.all([
          insightsApi.summary(),
          insightsApi.departmentStats(),
        ]);
        setSummary(s);
        setDeptStats(d);
      } catch (err) {
        console.error('Failed to load dashboard', err);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  useEffect(() => {
    if (!selectedCountry) {
      setCountryRange(null);
      insightsApi.departmentStats().then(setDeptStats).catch(console.error);
      return;
    }
    Promise.all([
      insightsApi.salaryRange(selectedCountry),
      insightsApi.departmentStats(selectedCountry),
    ]).then(([range, dept]) => {
      setCountryRange(range);
      setDeptStats(dept);
    }).catch(console.error);
  }, [selectedCountry]);

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-gray-400">Loading dashboard...</div>;
  }

  if (!summary) return null;

  const kpis = [
    { label: 'Total Employees', value: summary.total_employees.toLocaleString(), icon: Users, color: 'bg-indigo-50 text-indigo-600' },
    { label: 'Average Salary', value: `$${formatCompact(summary.average_salary)}`, icon: DollarSign, color: 'bg-green-50 text-green-600' },
    { label: 'Countries', value: summary.total_countries, icon: Globe, color: 'bg-cyan-50 text-cyan-600' },
    { label: 'Departments', value: summary.total_departments, icon: Building2, color: 'bg-amber-50 text-amber-600' },
  ];

  const headcountData = summary.country_breakdown.map(c => ({
    name: c.country.length > 12 ? c.country.substring(0, 10) + '...' : c.country,
    fullName: c.country,
    employees: c.employee_count,
    avgSalary: Math.round(c.average_salary),
  }));

  const deptChartData = deptStats.map(d => ({
    name: d.department.length > 12 ? d.department.substring(0, 10) + '...' : d.department,
    fullName: d.department,
    average: Math.round(d.average_salary),
    count: d.employee_count,
  }));

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Salary Insights</h1>
          <p className="text-sm text-gray-500 mt-1">Organization-wide salary analytics</p>
        </div>
        <select value={selectedCountry} onChange={e => setSelectedCountry(e.target.value)}
          className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 outline-none">
          <option value="">All Countries</option>
          {summary.country_breakdown.map(c => (
            <option key={c.country} value={c.country}>{c.country}</option>
          ))}
        </select>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {kpis.map(({ label, value, icon: Icon, color }) => (
          <div key={label} className="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3">
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${color}`}>
                <Icon className="w-5 h-5" />
              </div>
              <div>
                <p className="text-sm text-gray-500">{label}</p>
                <p className="text-xl font-bold text-gray-900">{value}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Country Salary Range (shown when country selected) */}
      {countryRange && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            Salary Range — {countryRange.country}
          </h2>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            {[
              { label: 'Minimum', value: countryRange.min },
              { label: 'Maximum', value: countryRange.max },
              { label: 'Average', value: countryRange.average },
              { label: 'Median', value: countryRange.median },
            ].map(({ label, value }) => (
              <div key={label} className="text-center p-3 bg-gray-50 rounded-lg">
                <p className="text-xs text-gray-500 mb-1">{label}</p>
                <p className="text-lg font-bold text-gray-900">{formatCompact(value)}</p>
              </div>
            ))}
          </div>
          <p className="text-xs text-gray-400 mt-3 text-center">{countryRange.count} employees</p>
        </div>
      )}

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Headcount by Country */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Headcount by Country</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={headcountData} margin={{ top: 5, right: 20, bottom: 60, left: 10 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="name" angle={-45} textAnchor="end" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} />
              <Tooltip
                formatter={(value) => [Number(value).toLocaleString(), 'Employees']}
                labelFormatter={(_label, payload) =>
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  (payload?.[0] as any)?.payload?.fullName || String(_label)
                }
              />
              <Bar dataKey="employees" fill="#4f46e5" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Headcount Pie */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Employee Distribution</h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie data={headcountData} dataKey="employees" nameKey="fullName"
                cx="50%" cy="50%" outerRadius={100}
                label={({ name, percent }: { name?: string; percent?: number }) =>
                  `${name ?? ''} ${((percent ?? 0) * 100).toFixed(0)}%`
                }
                labelLine={{ strokeWidth: 1 }} fontSize={11}>
                {headcountData.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip formatter={(value) => [Number(value).toLocaleString(), 'Employees']} />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Department Stats */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          Average Salary by Department {selectedCountry && `— ${selectedCountry}`}
        </h2>
        <ResponsiveContainer width="100%" height={350}>
          <BarChart data={deptChartData} layout="vertical" margin={{ top: 5, right: 30, bottom: 5, left: 100 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
            <XAxis type="number" tick={{ fontSize: 11 }}
              tickFormatter={(v) => formatCompact(v)} />
            <YAxis type="category" dataKey="name" tick={{ fontSize: 11 }} width={95} />
            <Tooltip
              formatter={(value) => [Number(value).toLocaleString(), 'Avg Salary']}
              labelFormatter={(_label, payload) =>
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                (payload?.[0] as any)?.payload?.fullName || String(_label)
              }
            />
            <Bar dataKey="average" fill="#06b6d4" radius={[0, 4, 4, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
