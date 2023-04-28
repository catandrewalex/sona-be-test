CREATE TABLE users (
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  email VARCHAR(64) NOT NULL UNIQUE,
  -- this can be filled with anything, JSON encoded
  user_detail json NOT NULL DEFAULT '{}',
  privilege_type INT NOT NULL DEFAULT 0,
  is_deactivated TINYINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_credentials (
  user_id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  email VARCHAR(64) NOT NULL UNIQUE,
  password CHAR(64) NOT NULL
);
