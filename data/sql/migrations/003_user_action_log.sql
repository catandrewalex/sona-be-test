CREATE TABLE user_action_log (
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  date DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  user_id BIGINT unsigned,
  FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE SET NULL,
  privilege_type INT NOT NULL,
  endpoint VARCHAR(64) NOT NULL,
  method VARCHAR(24) NOT NULL,
  status_code SMALLINT unsigned NOT NULL,
  request_body TEXT NOT NULL
);
