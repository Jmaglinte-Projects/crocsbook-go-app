CREATE INDEX idx_post_reactions_post_user
ON post_reactions (post_id, user_id);