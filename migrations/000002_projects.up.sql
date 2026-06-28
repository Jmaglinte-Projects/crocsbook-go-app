CREATE TABLE IF NOT EXISTS projects (
  project_id VARCHAR(40) NOT NULL,
  project_user_id VARCHAR(40) NOT NULL,

  name VARCHAR(200) NOT NULL,
  description TEXT,
  thumbnail_key TEXT,
  location VARCHAR(200),
  cost BIGINT,
  start_date DATE,
  completion_date DATE,
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_time TIMESTAMP,
  PRIMARY KEY (`project_id`, `project_user_id`)
);