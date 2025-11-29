-- name: NewUser :one
INSERT INTO users (username, password_hash) VALUES (?, ?) RETURNING id;

-- name: UpdateUserPassword :one
UPDATE users SET
password_hash = sqlc.arg(newPassword), updated_at = unixepoch()
WHERE id = sqlc.arg(userId) RETURNING *;

-- name: UpdateUserAdminStatus :one
UPDATE users SET
is_admin = sqlc.arg(newStatus), updated_at = unixepoch()
WHERE id = sqlc.arg(userId) RETURNING *;

-- name: GetUsers :many
SELECT * FROM users ORDER BY username;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = ?;

-- name: GetUserByName :one
SELECT * FROM users
WHERE username = ?;

-- name: GetUsersByStatus :many
SELECT * FROM users
WHERE is_admin = ?
ORDER BY username;

-- name: CountUsersByStatus :one
SELECT COUNT(*) as count FROM users
WHERE is_admin = sqlc.arg(status);

-- name: GetUsersPaginated :many
SELECT * FROM users
ORDER BY username ASC
LIMIT ? OFFSET ?;

-- name: GetUsersPaginatedFiltered :many
SELECT * FROM users
WHERE username LIKE '%' || sqlc.arg(search) || '%'
ORDER BY username ASC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*) as count FROM users;

-- name: CountUsersFiltered :one
SELECT COUNT(*) as count FROM users
WHERE username LIKE '%' || sqlc.arg(search) || '%';

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
       
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
SELECT * FROM challenges
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

-- name: GetUserStats :one
SELECT 
    u.username,
    u.created_at,
    u.is_admin,
    COUNT(DISTINCT ucc.challenge_id) as challenges_completed,
    COUNT(DISTINCT s.id) as total_solutions,
    COUNT(DISTINCT CASE WHEN s.status = 'accepted' THEN s.id END) as accepted_solutions,
    COUNT(DISTINCT CASE WHEN s.status = 'wrong_answer' THEN s.id END) as wrong_answers,
    COUNT(DISTINCT CASE WHEN s.status = 'runtime_error' THEN s.id END) as runtime_errors
FROM users u
LEFT JOIN user_completed_challenges ucc ON u.id = ucc.user_id
LEFT JOIN solutions s ON u.id = s.user_id
WHERE u.username = ?
GROUP BY u.id, u.username, u.created_at, u.is_admin;

-- name: GetUserCompletedChallengesWithDetails :many
SELECT 
    c.id as challenge_id,
    c.title,
    c.difficulty,
    c.test_count,
    ucc.completed_at,
    s.language,
    s.runtime_info,
    s.created_at as solution_created_at
FROM user_completed_challenges ucc
JOIN challenges c ON ucc.challenge_id = c.id
LEFT JOIN solutions s ON ucc.best_solution_id = s.id
JOIN users u ON ucc.user_id = u.id
WHERE u.username = ?
ORDER BY ucc.completed_at DESC;

-- name: GetUserRecentSubmissions :many
SELECT 
    s.id,
    s.challenge_id,
    c.title as challenge_title,
    s.language,
    s.status,
    s.runtime_info,
    s.created_at
FROM solutions s
JOIN challenges c ON s.challenge_id = c.id
JOIN users u ON s.user_id = u.id
WHERE u.username = ?
ORDER BY s.created_at DESC
LIMIT ?;

-- name: GetUserDifficultyBreakdown :many
SELECT 
    c.difficulty,
    COUNT(DISTINCT ucc.challenge_id) as completed_count
FROM user_completed_challenges ucc
JOIN challenges c ON ucc.challenge_id = c.id
JOIN users u ON ucc.user_id = u.id
WHERE u.username = ?
GROUP BY c.difficulty
ORDER BY c.difficulty;

-- name: GetUserLanguageStats :many
SELECT 
    s.language,
    COUNT(*) as submission_count,
    COUNT(CASE WHEN s.status = 'accepted' THEN 1 END) as accepted_count
FROM solutions s
JOIN users u ON s.user_id = u.id
WHERE u.username = ?
GROUP BY s.language
ORDER BY submission_count DESC;

-- name: GetProblemsForAdmin :many
SELECT 
    c.id,
    c.title,
    c.description,
    c.difficulty,
    c.test_count,
    c.created_at,
    c.updated_at,
    COUNT(DISTINCT s.user_id) as solver_count,
    COUNT(DISTINCT s.id) as submission_count
FROM challenges c
LEFT JOIN solutions s ON c.id = s.challenge_id AND s.status = 'accepted'
GROUP BY c.id, c.title, c.description, c.difficulty, c.test_count, c.created_at, c.updated_at
ORDER BY c.created_at DESC;

-- name: UpdateChallenge :one
UPDATE challenges 
SET title = ?, description = ?, difficulty = ?, updated_at = unixepoch()
WHERE id = ?
RETURNING *;

-- name: DeleteChallenge :exec
DELETE FROM challenges WHERE id = ?;

-- name: GetChallengeWithTests :one
SELECT 
    c.id,
    c.title,
    c.description,
    c.difficulty,
    c.test_count,
    c.created_at,
    c.updated_at
FROM challenges c
WHERE c.id = ?;

-- name: UpdateChallengeTest :one
UPDATE challenge_tests 
SET input_data = ?, expected_output = ?
WHERE id = ?
RETURNING *;

-- name: DeleteChallengeTest :exec
DELETE FROM challenge_tests WHERE id = ?;

-- name: GetAdminStats :one
SELECT
    (SELECT COUNT(*) FROM challenges) as total_challenges,
    (SELECT COUNT(*) FROM users) as total_users,
    (SELECT COUNT(*) FROM solutions) as total_submissions,
    (SELECT COUNT(*) FROM solutions WHERE status = 'accepted') as accepted_submissions;
