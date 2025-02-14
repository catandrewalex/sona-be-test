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
  default_fee INT NOT NULL,
  default_duration_minute INT NOT NULL,
  instrument_id BIGINT unsigned NOT NULL,
  grade_id BIGINT unsigned NOT NULL,
  -- `course` consists of `instrument` + `grade`, and `course` is importantly referred by other tables, so these 3 should always be coupled
  FOREIGN KEY (instrument_id) REFERENCES instrument(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  FOREIGN KEY (grade_id) REFERENCES grade(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  UNIQUE KEY `instrument_id--grade_id` (`instrument_id`, `grade_id`)
);

CREATE TABLE class
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  transport_fee INT NOT NULL,
  teacher_id BIGINT unsigned,
  course_id BIGINT unsigned NOT NULL,
  -- auto_owe_attendance_token determines whether a newly added `attendance` will automatically be assigned to the latest `student_learning_token` (or automatically create one) when the remaining quota is <= 0.
  -- the resulting `student_learning_token` quota will be negative, thus the term "owing".
  auto_owe_attendance_token TINYINT NOT NULL DEFAULT 1,
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
  is_deleted TINYINT NOT NULL DEFAULT 0,
  -- `student_enrollment` is an entity for many-to-many relationship
  FOREIGN KEY (student_id) REFERENCES student(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (class_id) REFERENCES class(id) ON UPDATE CASCADE ON DELETE CASCADE,
  UNIQUE KEY `student_id--class_id` (`student_id`, `class_id`)
);

CREATE TABLE enrollment_payment
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  payment_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  balance_top_up INT NOT NULL,
  balance_bonus INT NOT NULL DEFAULT 0,
  course_fee_value INT NOT NULL,
  transport_fee_value INT NOT NULL,
  penalty_fee_value INT NOT NULL,
  discount_fee_value INT NOT NULL DEFAULT 0,
  enrollment_id BIGINT unsigned,
  -- `enrollment_payment` stores historical records, and must not be deleted by CASCADE, but allow deletion of the parent entity
  FOREIGN KEY (enrollment_id) REFERENCES student_enrollment(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE student_learning_token
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  quota FLOAT NOT NULL DEFAULT 4,
  course_fee_quarter_value INT NOT NULL,
  transport_fee_quarter_value INT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  enrollment_id BIGINT unsigned NOT NULL,
  -- `student_learning_token` has quota, whose value must be transferrable to another `enrollment` before a student `enrollment` is deleted
  FOREIGN KEY (enrollment_id) REFERENCES student_enrollment(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE teacher_special_fee
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  fee INT NOT NULL,
  teacher_id BIGINT unsigned NOT NULL,
  course_id BIGINT unsigned NOT NULL,
  -- `teacher_special_fee` acts as an additional information, and couples a `teacher` with a `course`. We can simply delete this record by CASCADE
  FOREIGN KEY (teacher_id) REFERENCES teacher(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (course_id) REFERENCES course(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE attendance
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  used_student_token_quota FLOAT NOT NULL,
  duration INT NOT NULL,
  note VARCHAR(255) NOT NULL DEFAULT '',
  is_paid TINYINT NOT NULL DEFAULT 0,
  class_id BIGINT unsigned NOT NULL,
  teacher_id BIGINT unsigned NOT NULL,
  student_id BIGINT unsigned NOT NULL,
  -- an `attendance` may have null `student_learning_token` when the class is NOT on "auto_owe_attendance_token". That mode allows adding "dangling" `attendance`, whose `student_learning_token` will be assigned later.
  -- this is mostly useful in the period of course level up, or price change. In this scenario, some `attendance`'s `student_learning_token` are preferred to be assigned manually, instead of automatically.
  -- therefore, we need to be able to insert the `attendance` having no `student_learning_token`.
  token_id BIGINT unsigned,
  -- `attendance` stores historical records, and requires all existing used `attendance` to be deleted before deleting the parent entities.
  FOREIGN KEY (class_id) REFERENCES class(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  FOREIGN KEY (teacher_id) REFERENCES teacher(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  FOREIGN KEY (student_id) REFERENCES student(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  -- an `attendance` normally will have a `student_learning_token` for calculating `attendance` fee. If one wishes to delete a token ID, we force the `attendance` to migrate to use another token.
  FOREIGN KEY (token_id) REFERENCES student_learning_token(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  UNIQUE KEY `class_id--student_id--date` (`class_id`, `student_id`, `date`)
);

CREATE TABLE teacher_payment
(
  id BIGINT unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  attendance_id BIGINT unsigned NOT NULL UNIQUE,
  paid_course_fee_value INT NOT NULL,
  paid_transport_fee_value INT NOT NULL,
  added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- `teacher_payment` stores historical records of teacher payment, and must be deleted explicitly
  FOREIGN KEY (attendance_id) REFERENCES attendance(id) ON UPDATE CASCADE ON DELETE RESTRICT
);
