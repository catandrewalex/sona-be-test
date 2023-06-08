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
