CREATE TABLE teams (
                       team_name TEXT PRIMARY KEY
);

CREATE TABLE users (
                       user_id TEXT PRIMARY KEY,
                       username TEXT NOT NULL,
                       team_name TEXT REFERENCES teams(team_name),
                       is_active BOOLEAN DEFAULT FALSE
);

CREATE TABLE pull_requests (
                               pull_request_id TEXT PRIMARY KEY,
                               pull_request_name TEXT NOT NULL,
                               author_id TEXT REFERENCES users(user_id),
                               status TEXT CHECK (status IN ('OPEN', 'MERGED')),
                               created_at TIMESTAMP,
                               merged_at TIMESTAMP
);

CREATE TABLE reviewers (
                              pull_request_id TEXT REFERENCES pull_requests(pull_request_id),
                              reviewer_id TEXT REFERENCES users(user_id),
                              PRIMARY KEY (pull_request_id, reviewer_id)
);
