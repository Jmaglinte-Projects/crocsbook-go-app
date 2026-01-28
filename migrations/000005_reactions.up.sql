CREATE TABLE IF NOT EXISTS medias (
  reaction_id VARCHAR(40) NOT NULL,
  reaction_project_id VARCHAR(40) NOT NULL,

  type ENUM("Like"),
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`reaction_id`, `reaction_project_id`)
);