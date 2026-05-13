# Salary Manager — Design Document

## Overview

A minimal yet usable salary management tool for an organisation with 10,000 employees.
The primary user persona is an **HR Manager** who needs to manage employee data and
gain salary insights across countries and job titles.

## Architecture

```
┌─────────────────┐         ┌─────────────────────────────────────┐
│  React Frontend │  REST   │          Go Backend                 │
│  (Vite, TS)     │◄───────►│  chi router → handler → service    │
│  Port :5173     │  JSON   │         → repository → SQLite      │
│                 │         │  Port :8080                         │
└─────────────────┘         └─────────────────────────────────────┘
```

### Layered Backend

| Layer      | Responsibility                                      |
|------------|------------------------------------------------------|
| Handler    | HTTP concerns: parse request, validate input, return JSON |
| Service    | Business logic: validation rules, aggregation        |
| Repository | Data access: SQL queries behind an interface         |

Each layer depends only on the layer below it via interfaces, enabling
isolated testing and easy swapping (e.g., replacing SQLite with Postgres
requires only a new repository implementation).

## Technology Choices

| Concern       | Choice                    | Why                                                   |
|---------------|---------------------------|-------------------------------------------------------|
| Go version    | 1.23 via GVM              | Latest stable; GVM for consistent version management  |
| HTTP router   | go-chi/chi                | Lightweight, idiomatic, good middleware ecosystem     |
| Database      | SQLite (mattn/go-sqlite3) | Zero-infrastructure; 10K rows is trivially handled    |
| Frontend      | React + TypeScript + Vite | Rich interactivity for tables, filters, charts        |
| Styling       | Tailwind CSS              | Utility-first, rapid UI development                   |
| Testing       | Go stdlib `testing`       | Fast, no external test framework needed               |

## Data Model

The `Employee` entity captures all required fields plus additional fields
that provide meaningful context for an HR manager:

- **Email**: Unique identity, useful for contact
- **Department**: Organizational context, enables department-level analytics
- **Currency**: Supports multi-country salary comparison
- **JoinDate**: Enables tenure-based insights

## Key Trade-offs

### SQLite vs PostgreSQL
SQLite was chosen as the default for zero-setup simplicity. With 10,000 employees,
read/write performance is excellent. The repository interface pattern means switching
to PostgreSQL requires only a new `repository/postgres/` implementation — no changes
to service or handler layers.

### Separate Frontend & Backend
Frontend and backend are independent projects in separate directories. Benefits:
- Independent deployment (CDN for frontend, server for backend)
- Independent development workflows and tooling
- Clear API contract between the two
- During development, Vite's proxy forwards `/api` calls to Go server

### No Authentication
Omitted to keep scope focused. In production, would add JWT-based auth with
role-based access control (HR Manager vs read-only viewer).

### Currency Handling
We store raw salary amounts with a currency code. Currency conversion is out of
scope but the model supports adding it later.

### Pagination
Server-side pagination with default page size of 20. Essential for 10K records
to keep API responses fast and UI responsive.

## Performance Considerations

- **Database indexes**: On `country`, `job_title`, and `department` for fast
  filtered queries and aggregations
- **Pagination**: All list endpoints paginate server-side
- **Aggregation queries**: Salary insights use SQL aggregation (MIN, MAX, AVG)
  rather than loading all rows into memory
- **SQLite WAL mode**: Enabled for better concurrent read performance

## AI Tools Usage

This project was built with AI assistance (Cursor with Claude). The AI was used for:
- Generating boilerplate code and CRUD implementations
- Writing test cases and table-driven test patterns
- Suggesting API design patterns
- Generating seed data logic

All AI-generated code was reviewed and refined for correctness, idiomatic Go style,
and production quality.
