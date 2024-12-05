-- Add check to feedbacks to see if I've already handle it
-- We set it to false by default, the existing feedback will get false,
-- and futures feedback will have false without requiring a value
ALTER TABLE achievements ADD "check" BOOLEAN NOT NULL DEFAULT FALSE;
