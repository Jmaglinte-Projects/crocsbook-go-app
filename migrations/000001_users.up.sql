CREATE TABLE IF NOT EXISTS users (
  user_id VARCHAR(40) NOT NULL,
  email VARCHAR(50) NOT NULL DEFAULT '',
  gender VARCHAR(20) NOT NULL DEFAULT '', 
  profile_url TEXT,
  nickname VARCHAR(20),
  username VARCHAR(20),
  password VARCHAR(255),
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_time TIMESTAMP,
  PRIMARY KEY (`user_id`)
);