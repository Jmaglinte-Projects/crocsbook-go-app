CREATE TABLE chat_threads (
  chat_thread_id CHAR(36) PRIMARY KEY,
  type ENUM('direct', 'group') NOT NULL DEFAULT 'direct',
  title VARCHAR(255) NULL,

  last_message_id CHAR(36) NULL,
  last_message_at DATETIME NULL,

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NULL
);

CREATE TABLE chat_thread_participants (
  chat_thread_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,

  archived_at DATETIME NULL,
  muted_at DATETIME NULL,

  last_read_message_id CHAR(36) NULL,
  last_read_at DATETIME NULL,

  joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  left_at DATETIME NULL,

  PRIMARY KEY (chat_thread_id, user_id),

  INDEX idx_participants_user_archived (user_id, archived_at),
  INDEX idx_participants_user_last_read (user_id, last_read_at),

  FOREIGN KEY (chat_thread_id) REFERENCES chat_threads(chat_thread_id)
);

CREATE TABLE chat_messages (
  chat_message_id CHAR(36) PRIMARY KEY,
  chat_thread_id CHAR(36) NOT NULL,
  sender_user_id CHAR(36) NOT NULL,

  message_type ENUM('text', 'image', 'attachment') NOT NULL DEFAULT 'text',
  body TEXT NULL,

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NULL,
  deleted_at DATETIME NULL,

  INDEX idx_messages_thread_created (chat_thread_id, created_at),
  INDEX idx_messages_sender (sender_user_id),

  FOREIGN KEY (chat_thread_id) REFERENCES chat_threads(chat_thread_id)
);

CREATE TABLE chat_message_attachments (
  chat_message_attachment_id CHAR(36) PRIMARY KEY,
  chat_message_id CHAR(36) NOT NULL,

  file_name VARCHAR(255) NOT NULL,
  file_type VARCHAR(100) NULL,
  file_size BIGINT NULL,

  r2_bucket VARCHAR(255) NOT NULL,
  r2_object_key VARCHAR(500) NOT NULL,
  public_url TEXT NULL,

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_attachments_message (chat_message_id),

  FOREIGN KEY (chat_message_id) REFERENCES chat_messages(chat_message_id)
);
