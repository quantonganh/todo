CREATE OR REPLACE FUNCTION subtasks (parent_ids int)
RETURNS jsonb[] AS $$
BEGIN
    RETURN (
        SELECT array_agg(row_to_json(t)) FROM
            (
                SELECT id, description, status, due_date, subtasks(id) AS sub_tasks
                FROM task
                WHERE parent_id = parent_ids
            ) t
        );
END;
$$ LANGUAGE plpgsql