package mysql

// ============================== TEACHER ==============================
func (r GetTeacherByIdRow) ToGetTeachersRow() GetTeachersRow {
	return GetTeachersRow(r)
}
func (r GetTeachersByIdsRow) ToGetTeachersRow() GetTeachersRow {
	return GetTeachersRow(r)
}

// ============================== STUDENT ==============================
func (r GetStudentByIdRow) ToGetStudentsRow() GetStudentsRow {
	return GetStudentsRow(r)
}
func (r GetStudentsByIdsRow) ToGetStudentsRow() GetStudentsRow {
	return GetStudentsRow(r)
}

// ============================== COURSE ==============================

func (r GetCourseByIdRow) ToGetCoursesRow() GetCoursesRow {
	return GetCoursesRow(r)
}
func (r GetCoursesByIdsRow) ToGetCoursesRow() GetCoursesRow {
	return GetCoursesRow(r)
}

// ============================== CLASS ==============================

func (r GetClassByIdRow) ToGetClassesRow() GetClassesRow {
	return GetClassesRow(r)
}
func (r GetClassesByIdsRow) ToGetClassesRow() GetClassesRow {
	return GetClassesRow(r)
}

// ============================== STUDENT_ENROLLMENT ==============================

func (r GetStudentEnrollmentByIdRow) ToGetStudentEnrollmentsRow() GetStudentEnrollmentsRow {
	return GetStudentEnrollmentsRow(r)
}

// ============================== TEACHER_SPECIAL_FEE ==============================

func (r GetTeacherSpecialFeeByIdRow) ToGetTeacherSpecialFeesRow() GetTeacherSpecialFeesRow {
	return GetTeacherSpecialFeesRow(r)
}
func (r GetTeacherSpecialFeesByIdsRow) ToGetTeacherSpecialFeesRow() GetTeacherSpecialFeesRow {
	return GetTeacherSpecialFeesRow(r)
}
func (r GetTeacherSpecialFeesByTeacherIdRow) ToGetTeacherSpecialFeesRow() GetTeacherSpecialFeesRow {
	return GetTeacherSpecialFeesRow(r)
}

// ============================== ENROLLMENT_PAYMENT ==============================

func (r GetEnrollmentPaymentByIdRow) ToGetEnrollmentPaymentsRow() GetEnrollmentPaymentsRow {
	return GetEnrollmentPaymentsRow(r)
}
func (r GetEnrollmentPaymentsByIdsRow) ToGetEnrollmentPaymentsRow() GetEnrollmentPaymentsRow {
	return GetEnrollmentPaymentsRow(r)
}
func (r GetEnrollmentPaymentsDescendingDateRow) ToGetEnrollmentPaymentsRow() GetEnrollmentPaymentsRow {
	return GetEnrollmentPaymentsRow(r)
}

// ============================== STUDENT_LEARNING_TOKEN ==============================

func (r GetStudentLearningTokenByIdRow) ToGetStudentLearningTokensRow() GetStudentLearningTokensRow {
	return GetStudentLearningTokensRow(r)
}
func (r GetStudentLearningTokensByIdsRow) ToGetStudentLearningTokensRow() GetStudentLearningTokensRow {
	return GetStudentLearningTokensRow(r)
}

// ============================== PRESENCE ==============================

func (r GetPresenceByIdRow) ToGetPresencesRow() GetPresencesRow {
	return GetPresencesRow(r)
}
func (r GetPresencesByIdsRow) ToGetPresencesRow() GetPresencesRow {
	return GetPresencesRow(r)
}
