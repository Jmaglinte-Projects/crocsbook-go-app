CREATE TABLE IF NOT EXISTS post_comments (
  comment_id VARCHAR(40) NOT NULL,
  post_id VARCHAR(40) NOT NULL,
  user_id VARCHAR(40) NOT NULL,
  parent_comment_id VARCHAR(40) NULL,
  content TEXT NOT NULL,
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_time TIMESTAMP NULL,
  PRIMARY KEY (comment_id)
);

