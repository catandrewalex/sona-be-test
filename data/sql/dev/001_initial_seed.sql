/* ============================== USER & USER_CREDENTIAL ============================== */
INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "userone@sonamusica.com", "userone", '{ "firstName": "User", "lastName": "One" }', 200
);
INSERT INTO user_credential (
  user_id, email, password
) VALUES (
  1, "userone@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "usertwo@sonamusica.com", "usertwo", '{ "firstName": "User", "lastName": "Two" }', 200
);
INSERT INTO user_credential (
  user_id, email, password
) VALUES (
  2, "usertwo@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "staffone@sonamusica.com", "staffone", '{ "firstName": "Staff", "lastName": "One" }', 300
);
INSERT INTO user_credential (
  user_id, email, password
) VALUES (
  3, "staffone@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "stafftwo@sonamusica.com", "stafftwo", '{ "firstName": "Staff", "lastName": "Two" }', 300
);
INSERT INTO user_credential (
  user_id, email, password
) VALUES (
  4, "stafftwo@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "adminone@sonamusica.com", "adminone", '{ "firstName": "Admin", "lastName": "One" }', 400
);
INSERT INTO user_credential (
  user_id, email, password
) VALUES (
  5, "adminone@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
);

INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  "admintwo@sonamusica.com", "admintwo", '{ "firstName": "Admin", "lastName": "Two" }', 400
);
INSERT INTO user_credential (
  user_id, email, password
) VALUES (
  6, "admintwo@sonamusica.com", "$2a$10$ao4yOZxrqv0TOmED.ZeoHOecLzYSgXpAZnAIVOmXAv8CWt1XAG/Lm" -- this equals to "pass"
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

/* ============================== CLASS & STUDENT ENROLLMENT ============================== */
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 0, 1, 1, 0 );
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 0, 2, 13, 0 );
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 100000, 3, 16, 0 );
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 50000, 4, 17, 0 );

INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 1, 1 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 2, 2 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 3, 3 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 4, 4 );

/* ---------- Class with multiple students ---------- */
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 150000, 4, 44, 0 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 1, 5 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 2, 5 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 3, 5 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 4, 5 );

/* ---------- Class with teacher, without student ---------- */
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 0, 4, 3, 0 );

/* ---------- Class without teacher, with student ---------- */
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 0, NULL, 2, 0 );
INSERT INTO student_enrollment ( student_id, class_id ) VALUES ( 2, 7 );

/* ---------- Class without teacher, without student ---------- */
INSERT INTO class ( default_transport_fee, teacher_id, course_id, is_deactivated ) VALUES ( 0, NULL, 36, 1 );
