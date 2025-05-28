-- name: NewUser :one
INSERT INTO users (username, password) VALUES (?, ?) RETURNING id;

-- name: UpdateUserPassword :one
UPDATE users SET
password = sqlc.arg(newPassword), updated_at = unixepoch()
WHERE id = sqlc.arg(userId) RETURNING *;

-- name: GetUsers :many
SELECT * FROM users ORDER BY username;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = ?;

-- name: GetUserByName :one
SELECT * FROM users
WHERE username = ?;
       
-- name: NewChallenge :one
INSERT INTO challenges (title, description, difficulty)
VALUES (?, ?, ?)
RETURNING id;

-- name: GetChallenge :one
SELECT * FROM challenges
WHERE id = ?;

-- name: GetChallenges :many
SELECT * FROM challenges
ORDER BY created_at DESC;

-- name: CountChallenges :one
SELECT COUNT(*) AS count FROM challenges;

-- name: GetChallengesPaginated :many
SELECT sqlc.embed(challenges), IFNULL(ch_t.test_count, 0) AS test_count FROM challenges
LEFT JOIN (
     SELECT count(*) as test_count
     FROM challenge_tests
     GROUP BY challenge_id
     ORDER BY
     test_count DESC
) AS ch_t
ORDER BY
     test_count DESC
LIMIT ? OFFSET ?;

-- name: NewChallengeTest :one
INSERT INTO challenge_tests (challenge_id, input_data, expected_output)
VALUES (?, ?, ?)
RETURNING id;

-- name: GetTestsForChallenge :many
SELECT * FROM challenge_tests
WHERE challenge_id = ?
ORDER BY id;

-- name: GetTestDataForChallenge :many
SELECT input_data as input, expected_output as output
FROM challenge_tests
WHERE challenge_id = ?
ORDER BY length(input_data) ASC;

-- name: NewSolution :one
INSERT INTO solutions (user_id, challenge_id, code, language, status, runtime_info)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: GetSolutionsForChallenge :many
SELECT * FROM solutions
WHERE challenge_id = ?
ORDER BY created_at DESC;

-- name: GetUserSolutions :many
SELECT * FROM solutions
WHERE user_id = ? AND challenge_id = ?
ORDER BY created_at DESC;

-- name: UpdateSolutionStatus :one
UPDATE solutions
SET status = sqlc.arg(newStatus),
    runtime_info = sqlc.arg(runtimeInfo)
WHERE id = sqlc.arg(solutionId)
RETURNING *;

-- name: GetCompletedChallengesForUser :many
SELECT * FROM user_completed_challenges
WHERE user_id = ?
ORDER BY completed_at DESC;
