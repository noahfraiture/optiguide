ALTER TABLE achievements ADD "check" BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE feedbacks
DROP "check";
