# Salary Manager

A full-stack salary management tool for an organization with 10,000 employees. Built with Go (backend) and React (frontend).

## Quick Start

### Prerequisites

- **Go 1.22+** (via GVM: `gvm use go1.23`)
- **Node.js 18+** and npm
- **GCC** (required for go-sqlite3 CGo compilation)

### 1. Seed the Database

```bash
cd backend
go run ./cmd/seed/main.go
```

This creates `salary-manager.db` with 10,000 realistic employee records (~0.5s).

### 2. Start the Backend

```bash
cd backend
go run ./main.go
```

Backend runs on **http://localhost:8080**.

### 3. Start the Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on **http://localhost:5173** with API proxy to the backend.

### 4. Run Tests

```bash
cd backend
go test ./... -v -count=1
```

50 tests across repository, service, and handler layers — all fast and deterministic.

## Project Structure

```
salary-manager/
├── backend/              # Go REST API server
│   ├── main.go           # Entry point
│   ├── internal/
│   │   ├── model/        # Domain types (Employee, filters, insights)
│   │   ├── repository/   # Repository interface + SQLite implementation
│   │   ├── service/      # Business logic and validation
│   │   └── handler/      # HTTP handlers (chi router)
│   └── cmd/seed/         # Data seeder CLI for 10K employees
├── frontend/             # React + TypeScript + Vite SPA
│   └── src/
│       ├── api/          # Typed API client
│       ├── components/   # Reusable UI (Layout, EmployeeModal)
│       ├── pages/        # EmployeesPage, DashboardPage
│       └── types/        # TypeScript interfaces
├── DESIGN.md             # Architecture decisions and trade-offs
└── Makefile              # Top-level build commands
```

## Features

### Employee Management (CRUD)
- Paginated employee list (20 per page, 10K records)
- Search by name or email
- Filter by country and job title
- Add, edit, and delete employees via modal forms
- Input validation (required fields, valid email, non-negative salary)

### Salary Insights API
- `GET /api/insights/summary` — org-wide KPIs
- `GET /api/insights/salary-range?country=X` — min/max/avg/median by country
- `GET /api/insights/salary-by-title?country=X&job_title=Y` — salary for title in country
- `GET /api/insights/department-stats?country=X` — salary breakdown by department

### Dashboard
- KPI cards: total employees, average salary, countries, departments
- Bar chart: headcount by country
- Pie chart: employee distribution
- Horizontal bar chart: average salary by department
- Country drill-down filter for all charts

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/employees?page=&limit=&search=&country=&job_title=` | List employees (paginated) |
| GET | `/api/employees/:id` | Get employee by ID |
| POST | `/api/employees` | Create employee |
| PUT | `/api/employees/:id` | Update employee |
| DELETE | `/api/employees/:id` | Delete employee |
| GET | `/api/insights/summary` | Organization summary |
| GET | `/api/insights/salary-range?country=` | Salary range by country |
| GET | `/api/insights/salary-by-title?country=&job_title=` | Salary by title |
| GET | `/api/insights/department-stats?country=` | Department statistics |
| GET | `/api/health` | Health check |

## Testing

The test suite covers all three backend layers:

- **Repository tests** (15): Use in-memory SQLite — real SQL execution, zero mocks
- **Service tests** (22): Mock repository — test validation, normalization, error handling
- **Handler tests** (13): httptest — test HTTP status codes, JSON serialization, routing

All tests are:
- Fast (< 5 seconds total)
- Deterministic (no external dependencies)
- Table-driven where applicable

## Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| Backend | Go 1.23 | Fast, strongly typed, great stdlib |
| Router | chi | Lightweight, idiomatic Go router |
| Database | SQLite | Zero-infra, handles 10K rows trivially |
| Frontend | React + TypeScript | Rich interactivity for data-heavy UI |
| Build | Vite | Fast dev server, HMR, production builds |
| Styling | Tailwind CSS | Utility-first, rapid UI development |
| Charts | Recharts | React-native charting library |
