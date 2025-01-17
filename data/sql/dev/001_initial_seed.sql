/* ============================== USER & USER_CREDENTIAL ============================== */
INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "userone@sonamusica.com", "userone", '{ "firstName": "User", "lastName": "One" }', 200
);
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  1, "userone", "userone@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "usertwo@sonamusica.com", "usertwo", '{ "firstName": "User", "lastName": "Two" }', 200
);
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  2, "usertwo", "usertwo@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "staffone@sonamusica.com", "staffone", '{ "firstName": "Staff", "lastName": "One" }', 300
);
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  3, "staffone", "staffone@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "stafftwo@sonamusica.com", "stafftwo", '{ "firstName": "Staff", "lastName": "Two" }', 300
);
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  4, "stafftwo", "stafftwo@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "adminone@sonamusica.com", "adminone", '{ "firstName": "Admin", "lastName": "One" }', 400
);
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  5, "adminone", "adminone@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "admintwo@sonamusica.com", "admintwo", '{ "firstName": "Admin", "lastName": "Two" }', 400
);
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  6, "admintwo", "admintwo@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

/* ============================== STUDENT & TEACHER ============================== */
INSERT INTO student ( user_id ) VALUES ( 1 );
INSERT INTO student ( user_id ) VALUES ( 2 );
INSERT INTO student ( user_id ) VALUES ( 3 );
INSERT INTO student ( user_id ) VALUES ( 4 );

INSERT INTO teacher ( user_id ) VALUES ( 3 );
INSERT INTO teacher ( user_id ) VALUES ( 4 );
INSERT INTO teacher ( user_id ) VALUES ( 5 );
INSERT INTO teacher ( user_id ) VALUES ( 6 );

/* ============================== INSTRUMENT, GRADE, & COURSE ============================== */
INSERT INTO instrument ( name ) VALUES ( "Vocal" );
INSERT INTO instrument ( name ) VALUES ( "Piano" );
INSERT INTO instrument ( name ) VALUES ( "Violin" );
INSERT INTO instrument ( name ) VALUES ( "Cello" );
INSERT INTO instrument ( name ) VALUES ( "Guitar Classic" );
INSERT INTO instrument ( name ) VALUES ( "Music Theory" );
INSERT INTO instrument ( name ) VALUES ( "Guitar Performance" );
INSERT INTO instrument ( name ) VALUES ( "Ukulele" );
INSERT INTO instrument ( name ) VALUES ( "Drum" );
INSERT INTO instrument ( name ) VALUES ( "Flute" );
INSERT INTO instrument ( name ) VALUES ( "Saxophone" );
INSERT INTO instrument ( name ) VALUES ( "Trumpet" );
INSERT INTO instrument ( name ) VALUES ( "Viola" );

INSERT INTO grade ( name ) VALUES ( "Children" );
INSERT INTO grade ( name ) VALUES ( "Adult" );
INSERT INTO grade ( name ) VALUES ( "Group" );
INSERT INTO grade ( name ) VALUES ( "Beginner A" );
INSERT INTO grade ( name ) VALUES ( "Beginner B" );
INSERT INTO grade ( name ) VALUES ( "Grade 1" );
INSERT INTO grade ( name ) VALUES ( "Grade 2" );
INSERT INTO grade ( name ) VALUES ( "Grade 3" );
INSERT INTO grade ( name ) VALUES ( "Grade 4" );
INSERT INTO grade ( name ) VALUES ( "Grade 5" );
INSERT INTO grade ( name ) VALUES ( "Grade 6" );
INSERT INTO grade ( name ) VALUES ( "Grade 7" );
INSERT INTO grade ( name ) VALUES ( "Grade 8" );
INSERT INTO grade ( name ) VALUES ( "Beginner" );
INSERT INTO grade ( name ) VALUES ( "Elementary" );
INSERT INTO grade ( name ) VALUES ( "Intermediate" );
INSERT INTO grade ( name ) VALUES ( "Advanced" );

/* ---------- Vocal ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 375000, 30, 1, 1 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 450000, 45, 1, 2 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 375000, 60, 1, 3 );
/* ---------- Piano ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 2, 4 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 2, 5 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 375000, 30, 2, 6 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 425000, 45, 2, 7 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 475000, 45, 2, 8 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 525000, 45, 2, 9 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 2, 10 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 750000, 60, 2, 11 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 2, 12 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 2, 13 );
/* ---------- Violin ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 3, 4 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 3, 5 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 375000, 30, 3, 6 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 425000, 45, 3, 7 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 475000, 45, 3, 8 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 525000, 45, 3, 9 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 3, 10 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 750000, 60, 3, 11 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 3, 12 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 3, 13 );
/* ---------- Cello ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 4, 4 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 4, 5 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 375000, 30, 4, 6 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 425000, 45, 4, 7 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 475000, 45, 4, 8 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 525000, 45, 4, 9 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 4, 10 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 750000, 60, 4, 11 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 4, 12 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 4, 13 );
/* ---------- Guitar Classic ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 5, 4 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 5, 5 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 375000, 30, 5, 6 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 425000, 45, 5, 7 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 475000, 45, 5, 8 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 525000, 45, 5, 9 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 5, 10 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 750000, 60, 5, 11 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 5, 12 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 5, 13 );
/* ---------- Music Theory ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 250000, 30, 6, 6 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 250000, 30, 6, 7 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 325000, 45, 6, 8 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 325000, 45, 6, 9 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 475000, 60, 6, 10 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 475000, 60, 6, 11 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 600000, 60, 6, 12 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 600000, 60, 6, 13 );
/* ---------- Guitar Performance ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 7, 14 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 450000, 45, 7, 15 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 7, 16 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 7, 17 );
/* ---------- Ukulele ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 325000, 30, 8, 14 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 400000, 45, 8, 15 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 500000, 45, 8, 16 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 60, 8, 17 );
/* ---------- Drum ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 9, 14 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 450000, 45, 9, 15 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 9, 16 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 9, 17 );
/* ---------- Flute ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 10, 14 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 725000, 45, 10, 15 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 45, 10, 16 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 950000, 60, 10, 17 );
/* ---------- Saxophone ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 30, 11, 14 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 45, 11, 15 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 45, 11, 16 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 60, 11, 17 );
/* ---------- Trumpet ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 30, 12, 14 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 45, 12, 15 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 45, 12, 16 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 0, 60, 12, 17 );
/* ---------- Viola ---------- */
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 13, 4 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 350000, 30, 13, 5 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 375000, 30, 13, 6 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 425000, 45, 13, 7 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 475000, 45, 13, 8 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 525000, 45, 13, 9 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 650000, 45, 13, 10 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 750000, 60, 13, 11 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 13, 12 );
INSERT INTO course ( default_fee, default_duration_minute, instrument_id, grade_id ) VALUES ( 850000, 60, 13, 13 );

/* ============================== CLASS & STUDENT ENROLLMENT ============================== */
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 0, 1, 1, 0, 0 );
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 0, 2, 13, 0, 0 );
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 100000, 3, 16, 0, 0 );
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 50000, 4, 17, 0, 0 );

INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 1, 1 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 2, 2 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 3, 3 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 4, 4 );

/* ---------- Class with multiple students ---------- */
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 150000, 4, 44, 0, 0 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 1, 5 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 2, 5 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 3, 5 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 4, 5 );

/* ---------- Class with teacher, without student ---------- */
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 0, 4, 3, 0, 0 );

/* ---------- Class without teacher, with student ---------- */
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 0, NULL, 2, 0, 0 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 2, 7 );

/* ---------- Class without teacher, without student ---------- */
INSERT INTO class ( transport_fee, teacher_id, course_id, auto_owe_attendance_token, is_deactivated ) VALUES ( 0, NULL, 36, 0, 1 );

/* ============================== TEACHER_SPECIAL_FEE ============================== */
INSERT INTO teacher_special_fee ( fee, teacher_id, course_id ) VALUES ( 575000, 3, 1 );
INSERT INTO teacher_special_fee ( fee, teacher_id, course_id ) VALUES ( 650000, 3, 2 );
INSERT INTO teacher_special_fee ( fee, teacher_id, course_id ) VALUES ( 575000, 3, 3 );

/* ============================== ENROLLMENT_PAYMENT ============================== */
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-07-01 07:00:00', 4, 300000, 0, 0, 0, 1
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-07-01 07:00:00', 4, 750000, 0, 0, 0, 2
);

INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-08-01 07:00:00', 4, 375000, 0, 20000, 0, 1
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-08-01 07:00:00', 4, 850000, 0, 0, 0, 2
);

INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-10-15 22:03:49', 4, 375000, 0, 40000, 0, 1
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-10-01 09:13:28', 4, 850000, 0, 20000, 0, 2
);


INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-08-28 14:14:22', 4, 250000, 37500, 0, 0, 5
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-08-28 14:24:41', 4, 250000, 37500, 0, 0, 6
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-08-28 14:34:55', 4, 250000, 37500, 0, 0, 7
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-08-28 14:45:06', 4, 250000, 37500, 0, 0, 8
);

INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-09-25 17:04:22', 4, 250000, 37500, 0, 0, 5
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-09-25 17:14:41', 4, 250000, 37500, 0, 0, 6
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-09-25 17:24:55', 4, 250000, 37500, 0, 0, 7
);
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    '2023-09-25 17:35:06', 4, 250000, 37500, 0, 0, 8
);

/* ============================== STUDENT_LEARNING_TOKEN ============================== */
INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    0, 75000, 0, '2023-07-01 07:00:00', '2023-07-01 07:00:00', 1
);

INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    0, 187500, 0, '2023-07-01 08:00:00', '2023-07-01 08:00:00', 2
);

INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    5, 93750, 0, '2023-08-01 07:00:00', '2023-08-01 07:00:00', 1
);

INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    4, 212500, 0, '2023-08-01 08:00:00', '2023-08-01 08:00:00', 2
);

INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    1, 62500, 9375, '2023-08-28 21:04:23', '2023-08-28 21:04:23', 5
);

INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    1, 62500, 9375, '2023-08-28 21:04:42', '2023-08-28 21:04:42', 6
);

INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    1, 62500, 9375, '2023-08-28 21:04:57', '2023-08-28 21:04:57', 7
);

INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    1, 62500, 9375, '2023-08-28 21:05:07', '2023-08-28 21:05:07', 8
);

/* ============================== ATTENDANCE & STUDENT_ATTEND ============================== */
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-01 09:00:00', 1, 30, 'voice placement technique 1', 0, 1, 1, 1, 1 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-08 09:10:00', 1, 30, 'voice placement technique 2', 0, 1, 1, 1, 1 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-15 08:55:00', 1, 30, 'voice placement technique 3', 0, 1, 1, 1, 1 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-23 08:55:00', 1, 30, 'voice placement technique 4', 0, 1, 1, 1, 1 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-30 09:00:00', 1, 30, 'breath control technique 1', 0, 1, 1, 1, 1 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-07 09:10:00', 1, 30, 'breath control technique 2', 0, 1, 1, 1, 3 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-14 08:55:00', 1, 30, 'breath control technique 3', 0, 1, 1, 1, 3 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-02 10:00:00', 1, 60, 'ABRSM grade 8 scales', 0, 2, 2, 2, 4 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-09 11:10:00', 1, 60, 'ABRSM grade 8 arpegios', 0, 2, 2, 2, 4 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-16 09:55:00', 1, 60, 'fantaise improptu', 0, 2, 2, 2, 4 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-07-23 09:35:20', 1, 60, 'fantaise improptu', 0, 2, 2, 2, 4 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-09-30 10:00:00', 1, 60, 'fantaise improptu: phrasing', 0, 2, 2, 2, 4 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-07 11:10:00', 1, 60, 'fantaise improptu: phrasing', 0, 2, 2, 2, 4 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-14 09:55:00', 1, 60, 'fantaise improptu: structure', 0, 2, 2, 2, 4 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-21 09:35:20', 1, 60, 'fantaise improptu: structure', 0, 2, 2, 2, 4 );


INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-03 05:00:00', 1, 30, 'time signature & key signature introduction', 0, 5, 4, 1, 5 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-03 05:00:00', 1, 30, 'time signature & key signature introduction', 0, 5, 4, 2, 6 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-03 05:00:00', 1, 30, 'time signature & key signature introduction', 0, 5, 4, 3, 7 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-03 05:00:00', 1, 30, 'time signature & key signature introduction', 0, 5, 4, 4, 8 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-10 05:00:00', 1, 30, 'notes value & bar counting', 0, 5, 4, 1, 5 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-10 05:00:00', 1, 30, 'notes value & bar counting', 0, 5, 4, 2, 6 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-10 05:00:00', 1, 30, 'notes value & bar counting', 0, 5, 4, 3, 7 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-10 05:00:00', 1, 30, 'notes value & bar counting', 0, 5, 4, 4, 8 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-17 05:00:00', 1, 30, 'scale & interval', 0, 5, 4, 1, 5 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-17 05:00:00', 1, 30, 'scale & interval', 0, 5, 4, 2, 6 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-17 05:00:00', 1, 30, 'scale & interval', 0, 5, 4, 3, 7 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-17 05:00:00', 1, 30, 'scale & interval', 0, 5, 4, 4, 8 );


INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-25 06:00:00', 1, 30, 'intervals & chords 1', 0, 5, 4, 1, 5 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-25 06:00:00', 1, 30, 'intervals & chords 1', 0, 5, 4, 2, 6 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-25 06:00:00', 1, 30, 'intervals & chords 1', 0, 5, 4, 3, 7 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-10-25 06:00:00', 1, 30, 'intervals & chords 1', 0, 5, 4, 4, 8 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-02 06:00:00', 1, 30, 'intervals & chords 2', 0, 5, 4, 1, 5 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-02 06:00:00', 1, 30, 'intervals & chords 2', 0, 5, 4, 2, 6 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-02 06:00:00', 1, 30, 'intervals & chords 2', 0, 5, 4, 3, 7 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-02 06:00:00', 1, 30, 'intervals & chords 2', 0, 5, 4, 4, 8 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-09 06:00:00', 1, 30, 'triad', 0, 5, 4, 1, 5 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-09 06:00:00', 1, 30, 'triad', 0, 5, 4, 2, 6 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-09 06:00:00', 1, 30, 'triad', 0, 5, 4, 3, 7 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-09 06:00:00', 1, 30, 'triad', 0, 5, 4, 4, 8 );

INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-16 06:00:00', 1, 30, 'triad 2 & cadence', 0, 5, 4, 1, 5 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-16 06:00:00', 1, 30, 'triad 2 & cadence', 0, 5, 4, 2, 6 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-16 06:00:00', 1, 30, 'triad 2 & cadence', 0, 5, 4, 3, 7 );
INSERT INTO attendance ( date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id ) VALUES ( '2023-11-16 06:00:00', 1, 30, 'triad 2 & cadence', 0, 5, 4, 4, 8 );
