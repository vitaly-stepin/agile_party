-- Migration: Create rooms table
-- Version: 1
-- Description: Initial schema for room metadata

CREATE TABLE IF NOT EXISTS rooms (
    id VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    voting_system VARCHAR(20) NOT NULL DEFAULT 'dbs_fibo',
    auto_reveal BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
