CREATE TRIGGER `slt_start_penalty_date_before_update` BEFORE
UPDATE
    ON `student_learning_token` FOR EACH ROW
SET
    NEW.penalty_start_at = CASE
        WHEN OLD.quota >= 0 AND NEW.quota < 0 THEN CURRENT_TIMESTAMP
        WHEN OLD.quota < 0 AND NEW.quota >= 0 THEN NULL
        ELSE OLD.penalty_start_at
    END;

CREATE TRIGGER `slt_start_penalty_date_before_insert` BEFORE
INSERT
    ON `student_learning_token` FOR EACH ROW
SET
    NEW.penalty_start_at = CASE
        WHEN NEW.quota < 0 THEN CURRENT_TIMESTAMP
        ELSE NULL
    END;
