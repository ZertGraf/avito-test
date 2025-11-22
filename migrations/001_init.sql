-- 001_init.sql

-- Write your migrate up statements here
CREATE TABLE teams (
                       team_name VARCHAR(255) PRIMARY KEY,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
                       user_id VARCHAR(255) PRIMARY KEY,
                       username VARCHAR(255) NOT NULL,
                       team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name),
                       is_active BOOLEAN NOT NULL DEFAULT true,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_team ON users(team_name);
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = true;

CREATE TABLE pull_requests (
                               pull_request_id VARCHAR(255) PRIMARY KEY,
                               pull_request_name VARCHAR(255) NOT NULL,
                               author_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
                               status VARCHAR(20) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               merged_at TIMESTAMPTZ
);

CREATE TABLE pr_reviewers (
                              pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id),
                              reviewer_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
                              assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);

---- create above / drop below ----

DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

