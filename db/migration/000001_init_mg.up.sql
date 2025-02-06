CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_name VARCHAR(100) NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL
);

CREATE INDEX idx_user_name ON users(user_name);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    allow_comments BOOLEAN NOT NULL,
    author_id INT NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);


CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    author_id INT NOT NULL,
    post_id INT NOT NULL,
    parent_id INT,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE INDEX idx_comment_post_id ON comments(post_id);
CREATE INDEX idx_comment_parent_id ON comments(parent_id);

CREATE OR REPLACE FUNCTION notify_new_comment() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('new_comment', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comment_insert_trigger
AFTER INSERT ON comments
FOR EACH ROW EXECUTE FUNCTION notify_new_comment();