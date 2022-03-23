CREATE OR REPLACE FUNCTION mark_parent_task_completed()
RETURNS TRIGGER AS $$
DECLARE
    r RECORD;
BEGIN
FOR r IN SELECT * FROM task WHERE parent_id = OLD.parent_id
    LOOP
        IF r.status != 'completed' THEN
            RETURN NULL;
        END IF;
    END LOOP;
    UPDATE task SET status = 'completed' WHERE id = OLD.parent_id;
    RETURN NEW;
END
$$
LANGUAGE plpgsql;

CREATE TRIGGER mark_parent_task_completed
    AFTER UPDATE ON task
    FOR EACH ROW
    EXECUTE FUNCTION mark_parent_task_completed();

