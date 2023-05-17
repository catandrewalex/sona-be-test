CREATE TABLE teacher
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT unsigned NOT NULL UNIQUE,
  FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE student
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT unsigned NOT NULL UNIQUE,
  FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE instrument
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE
);

CREATE TABLE grade
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE
);

CREATE TABLE course
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  default_fee BIGINT NOT NULL,
  default_duration_minute INT NOT NULL,
  instrument_id BIGINT unsigned NOT NULL,
  grade_id BIGINT unsigned NOT NULL,
  -- `course` consists of `instrument` + `grade`, and `course` is importantly referred by other tables, so these 3 should always be coupled
  FOREIGN KEY (instrument_id) REFERENCES instrument(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  FOREIGN KEY (grade_id) REFERENCES grade(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE class
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  default_transport_fee BIGINT NOT NULL,
  teacher_id BIGINT unsigned,
  course_id BIGINT unsigned NOT NULL,
  is_deactivated TINYINT NOT NULL DEFAULT 0,
  -- a `class` may temporarily have no `teacher`
  FOREIGN KEY (teacher_id) REFERENCES teacher(id) ON UPDATE CASCADE ON DELETE SET NULL,
  -- a `class` must migrate to another `course` first before the `course` getting deleted
  FOREIGN KEY (course_id) REFERENCES course(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE student_enrollment
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  student_id BIGINT unsigned NOT NULL,
  class_id BIGINT unsigned NOT NULL,
  -- `student_enrollment` is an entity for many-to-many relationship
  FOREIGN KEY (student_id) REFERENCES student(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (class_id) REFERENCES class(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE enrollment_payment
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  payment_date TIMESTAMP NOT NULL,
  balance_top_up INT NOT NULL,
  value INT NOT NULL,
  value_penalty INT NOT NULL,
  enrollment_id BIGINT unsigned,
  -- `enrollment_payment` stores historical records, and must not be deleted by CASCADE, but allow deletion of the parent entity
  FOREIGN KEY (enrollment_id) REFERENCES student_enrollment(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE student_learning_token
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  quota INT NOT NULL DEFAULT 4,
  quota_bonus INT NOT NULL DEFAULT 0,
  course_fee_value INT NOT NULL,
  transport_fee_value INT NOT NULL,
  last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  enrollment_id BIGINT unsigned,
  -- `student_learning_token` has quota, whose value must be transferrable to another `enrollment` before a student `enrollment` is deleted
  FOREIGN KEY (enrollment_id) REFERENCES student_enrollment(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE teacher_special_fee
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  fee BIGINT NOT NULL,
  teacher_id BIGINT unsigned NOT NULL,
  course_id BIGINT unsigned NOT NULL,
  -- `teacher_special_fee` acts as an additional information, and couples a `teacher` with a `course`. We can simply delete this record by CASCADE
  FOREIGN KEY (teacher_id) REFERENCES teacher(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (course_id) REFERENCES course(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE presence
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  date TIMESTAMP NOT NULL,
  used_student_token_quota FLOAT NOT NULL,
  duration INT NOT NULL,
  class_id BIGINT unsigned,
  teacher_id BIGINT unsigned,
  token_id BIGINT unsigned NOT NULL,
  -- `presence` stores historical records, and must not be deleted by CASCADE, but allow deletion of the parent entity
  FOREIGN KEY (class_id) REFERENCES class(id) ON UPDATE CASCADE ON DELETE SET NULL,
  FOREIGN KEY (teacher_id) REFERENCES teacher(id) ON UPDATE CASCADE ON DELETE SET NULL,
  -- a `presence` must have a `student_learning_token` for calculating `presence` fee. If one wishes to delete a token ID, we force the `presence` to migrate to use another token.
  FOREIGN KEY (token_id) REFERENCES student_learning_token(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE student_attend
(
  student_id BIGINT unsigned NOT NULL,
  presence_id BIGINT unsigned NOT NULL,
  CONSTRAINT pk_student_attend PRIMARY KEY (student_id, presence_id),
  -- `student_attend` is an entity for many-to-many relationship
  FOREIGN KEY (student_id) REFERENCES student(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (presence_id) REFERENCES presence(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE teacher_salary
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  presence_id BIGINT unsigned NOT NULL,
  course_fee_value INT NOT NULL,
  transport_fee_value INT NOT NULL,
  profit_sharing_percentage FLOAT NOT NULL,
  added_at DATE NOT NULL,
  -- `teacher_salary` stores historical records of teacher payment, and must be deleted explicitly
  FOREIGN KEY (presence_id) REFERENCES presence(id) ON UPDATE CASCADE ON DELETE RESTRICT
);
