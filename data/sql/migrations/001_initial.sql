CREATE TABLE user (
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  email VARCHAR(64) UNIQUE,
  -- this can be filled with anything, JSON encoded
  user_detail json NOT NULL DEFAULT (JSON_OBJECT()),
  privilege_type INT NOT NULL DEFAULT 0,
  is_deactivated TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_credential (
  user_id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
  username VARCHAR(50) NOT NULL UNIQUE,
  email VARCHAR(64) UNIQUE,
  password CHAR(64) NOT NULL
);
