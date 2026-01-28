CREATE TABLE IF NOT EXISTS medias (
  media_id VARCHAR(40) NOT NULL,
  media_project_id VARCHAR(40) NOT NULL,

  url TEXT,
  type ENUM("Image", "Video"),
  created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`media_id`, `media_project_id`)
);