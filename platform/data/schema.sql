-- Table: Users
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash BLOB NOT NULL, -- bcrypt-hashed password
	is_admin INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- Table: Challenges
CREATE TABLE challenges (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    difficulty INTEGER NOT NULL CHECK(difficulty BETWEEN 0 AND 3),
    test_count INTEGER NOT NULL DEFAULT 0, -- To avoid expensive subqueries
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- Table: Challenge_Tests
CREATE TABLE challenge_tests (
    id INTEGER PRIMARY KEY,
    challenge_id INTEGER NOT NULL,
    input_data TEXT NOT NULL,
    expected_output TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE
);

-- Add index to challenge_tests for better performance on other queries
CREATE INDEX idx_challenge_tests_challenge_id ON challenge_tests(challenge_id);

CREATE INDEX idx_challenges_test_count ON challenges(test_count);

-- Trigger to update test_count after inserting a challenge test
CREATE TRIGGER after_insert_challenge_test
AFTER INSERT ON challenge_tests
FOR EACH ROW
BEGIN
    UPDATE challenges
    SET test_count = test_count + 1
    WHERE id = NEW.challenge_id;
END;

-- Trigger to update test_count after deleting a challenge test
CREATE TRIGGER after_delete_challenge_test
AFTER DELETE ON challenge_tests
FOR EACH ROW
BEGIN
    UPDATE challenges
    SET test_count = test_count - 1
    WHERE id = OLD.challenge_id;
END;

-- Trigger to update test_count after updating a challenge test (specifically if challenge_id changes)
CREATE TRIGGER after_update_challenge_test
AFTER UPDATE ON challenge_tests
FOR EACH ROW
WHEN OLD.challenge_id != NEW.challenge_id
BEGIN
    -- Decrement count for the old challenge
    UPDATE challenges
    SET test_count = test_count - 1
    WHERE id = OLD.challenge_id;

    -- Increment count for the new challenge
    UPDATE challenges
    SET test_count = test_count + 1
    WHERE id = NEW.challenge_id;
END;

-- Table: Solutions (Submissions)
CREATE TABLE solutions (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    challenge_id INTEGER NOT NULL,
    code TEXT NOT NULL,
    language TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'accepted', 'wrong_answer', 'runtime_error')),
    runtime_info TEXT,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE
);

-- Table: User_Completed_Challenges (Many-to-Many linking)
CREATE TABLE user_completed_challenges (
    user_id INTEGER NOT NULL,
    challenge_id INTEGER NOT NULL,
    completed_at INTEGER NOT NULL DEFAULT (unixepoch()),
    best_solution_id INTEGER,
    PRIMARY KEY (user_id, challenge_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE,
    FOREIGN KEY (best_solution_id) REFERENCES solutions(id) ON DELETE SET NULL
);
