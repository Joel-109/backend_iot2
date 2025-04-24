-- name: GetSensorValues :many
SELECT *
FROM SensorValues 
ORDER BY created_at DESC
LIMIT ?;

-- name: GetLastSensorValue :one
SELECT *
FROM SensorValues 
ORDER BY created_at DESC
LIMIT 1;

-- name: InsertSensorValues :exec
INSERT INTO SensorValues (temperature,gas,flame)
VALUES (?,?,?);

-- name: GetRisks :many
SELECT * 
FROM Risk
ORDER BY created_at DESC
LIMIT ?;

-- name: InsertRisk :exec
INSERT INTO Risk (risk)
VALUES (?);

-- name: GetLastRisk :one
SELECT * 
FROM Risk
ORDER BY created_at DESC
LIMIT 1;