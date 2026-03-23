-- name: GetLatestOTP :one
SELECT *
FROM otp_requests
WHERE email = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateOTPRequest :exec
INSERT INTO otp_requests (email, otp_hash, expires_at, attempts, max_attempts, verified)
VALUES ($1, $2, $3, 0, $4, FALSE);

-- name: GetOTPRequestByEmail :one
SELECT *
FROM otp_requests
WHERE email = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateOTPRequest :exec
UPDATE otp_requests
SET attempts = $1,
    verified = $2,
    expires_at = $3,
    otp_hash = $4
WHERE email = $5;

-- name: DeleteOTPRequest :exec
DELETE FROM otp_requests
WHERE email = $1;

-- name: MarkOTPAsVerified :exec
UPDATE otp_requests
SET verified = TRUE
WHERE email = $1;

-- name: IncrementOTPAttempts :exec
UPDATE otp_requests
SET attempts = attempts + 1
WHERE email = $1;


