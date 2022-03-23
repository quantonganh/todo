CREATE TABLE IF NOT EXISTS task (
    id SERIAL PRIMARY KEY,
    description VARCHAR(255) NOT NULL,
    status TEXT CHECK (status IN ('new', 'in-progress', 'completed', 'overdue')) DEFAULT 'new',
    due_date TIMESTAMPTZ,
    parent_id INT REFERENCES task(id)
);