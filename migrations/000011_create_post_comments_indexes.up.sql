CREATE INDEX idx_post_comments_post_created
ON post_comments (post_id, created_time);