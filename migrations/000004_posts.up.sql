CREATE TABLE IF NOT EXISTS medias (
  post_id VARCHAR(40) NOT NULL,
  post_project_id VARCHAR(40) NOT NULL,

  content TEXT,
  visibility ENUM("Public", "Private"),
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_time TIMESTAMP,
  PRIMARY KEY (`post_id`, `post_project_id`)
);