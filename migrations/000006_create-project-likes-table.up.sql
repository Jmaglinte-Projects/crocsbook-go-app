CREATE TABLE IF NOT EXISTS project_likes (
  project_id VARCHAR(40) NOT NULL,
  project_user_id VARCHAR(40) NOT NULL,
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`project_id`, `project_user_id`)
);