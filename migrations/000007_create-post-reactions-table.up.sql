CREATE TABLE IF NOT EXISTS post_reactions (
  post_reaction_id VARCHAR(40) NOT NULL,
  post_id VARCHAR(40) NOT NULL,
  user_id VARCHAR(40) NOT NULL,
  reaction_type ENUM("Like", "Heart"),
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`post_reaction_id`)
);
