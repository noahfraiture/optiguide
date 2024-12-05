-- Add check to feedbacks to see if I've already handle it
-- We set it to false by default, the existing feedback will get false,
-- and futures feedback will have false without requiring a value
-- I previously put achievements, which is an error, it should be feedbacks
ALTER TABLE achievements
DROP "check";

ALTER TABLE feedbacks ADD "check" BOOLEAN NOT NULL DEFAULT FALSE;
