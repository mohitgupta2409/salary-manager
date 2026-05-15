CREATE TABLE IF NOT EXISTS countries (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL UNIQUE,
    code       TEXT    NOT NULL UNIQUE,
    currency   TEXT    NOT NULL,
    is_active  INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS departments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL UNIQUE,
    is_active  INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_titles (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    department_id INTEGER NOT NULL REFERENCES departments(id),
    is_active     INTEGER NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, department_id)
);

CREATE TABLE IF NOT EXISTS employees (
    id            INTEGER  PRIMARY KEY AUTOINCREMENT,
    first_name    TEXT     NOT NULL,
    last_name     TEXT     NOT NULL,
    email         TEXT     NOT NULL UNIQUE,
    job_title_id  INTEGER  NOT NULL REFERENCES job_titles(id),
    country_id    INTEGER  NOT NULL REFERENCES countries(id),
    salary        REAL     NOT NULL CHECK(salary >= 0),
    address       TEXT,
    join_date     DATETIME NOT NULL,
    is_active     INTEGER  NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_employees_country    ON employees(country_id);
CREATE INDEX IF NOT EXISTS idx_employees_job_title  ON employees(job_title_id);
CREATE INDEX IF NOT EXISTS idx_employees_email      ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_active     ON employees(is_active);
CREATE INDEX IF NOT EXISTS idx_job_titles_dept      ON job_titles(department_id);
