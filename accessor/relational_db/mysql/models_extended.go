package mysql

// ============================== TEACHER ==============================
func (r GetTeacherByIdRow) ToGetTeachersRow() GetTeachersRow {
	return GetTeachersRow{
		ID:            r.ID,
		UserID:        r.UserID,
		Username:      r.Username,
		Email:         r.Email,
		UserDetail:    r.UserDetail,
		PrivilegeType: r.PrivilegeType,
		IsDeactivated: r.IsDeactivated,
		CreatedAt:     r.CreatedAt,
	}
}
func (r GetTeachersByIdsRow) ToGetTeachersRow() GetTeachersRow {
	return GetTeachersRow{
		ID:            r.ID,
		UserID:        r.UserID,
		Username:      r.Username,
		Email:         r.Email,
		UserDetail:    r.UserDetail,
		PrivilegeType: r.PrivilegeType,
		IsDeactivated: r.IsDeactivated,
		CreatedAt:     r.CreatedAt,
	}
}

// ============================== STUDENT ==============================
func (r GetStudentByIdRow) ToGetStudentsRow() GetStudentsRow {
	return GetStudentsRow{
		ID:            r.ID,
		UserID:        r.UserID,
		Username:      r.Username,
		Email:         r.Email,
		UserDetail:    r.UserDetail,
		PrivilegeType: r.PrivilegeType,
		IsDeactivated: r.IsDeactivated,
		CreatedAt:     r.CreatedAt,
	}
}
func (r GetStudentsByIdsRow) ToGetStudentsRow() GetStudentsRow {
	return GetStudentsRow{
		ID:            r.ID,
		UserID:        r.UserID,
		Username:      r.Username,
		Email:         r.Email,
		UserDetail:    r.UserDetail,
		PrivilegeType: r.PrivilegeType,
		IsDeactivated: r.IsDeactivated,
		CreatedAt:     r.CreatedAt,
	}
}

// ============================== COURSE ==============================

func (r GetCourseByIdRow) ToGetCoursesRow() GetCoursesRow {
	return GetCoursesRow{
		CourseID:              r.CourseID,
		CourseName:            r.CourseName,
		DefaultFee:            r.DefaultFee,
		DefaultDurationMinute: r.DefaultDurationMinute,
	}
}
func (r GetCoursesByIdsRow) ToGetCoursesRow() GetCoursesRow {
	return GetCoursesRow{
		CourseID:              r.CourseID,
		CourseName:            r.CourseName,
		DefaultFee:            r.DefaultFee,
		DefaultDurationMinute: r.DefaultDurationMinute,
	}
}

// ============================== CLASS ==============================

func (r GetClassByIdRow) ToGetClassesRow() GetClassesRow {
	return GetClassesRow{
		ClassID:               r.ClassID,
		TransportFee:          r.TransportFee,
		IsDeactivated:         r.IsDeactivated,
		CourseID:              r.CourseID,
		TeacherID:             r.TeacherID,
		StudentID:             r.StudentID,
		EnrollmentID:          r.EnrollmentID,
		TeacherUsername:       r.TeacherUsername,
		TeacherDetail:         r.TeacherDetail,
		CourseName:            r.CourseName,
		StudentUsername:       r.StudentUsername,
		StudentDetail:         r.StudentDetail,
		DefaultFee:            r.DefaultFee,
		DefaultDurationMinute: r.DefaultDurationMinute,
	}
}
func (r GetClassesByIdsRow) ToGetClassesRow() GetClassesRow {
	return GetClassesRow{
		ClassID:               r.ClassID,
		TransportFee:          r.TransportFee,
		IsDeactivated:         r.IsDeactivated,
		CourseID:              r.CourseID,
		TeacherID:             r.TeacherID,
		StudentID:             r.StudentID,
		EnrollmentID:          r.EnrollmentID,
		TeacherUsername:       r.TeacherUsername,
		TeacherDetail:         r.TeacherDetail,
		CourseName:            r.CourseName,
		StudentUsername:       r.StudentUsername,
		StudentDetail:         r.StudentDetail,
		DefaultFee:            r.DefaultFee,
		DefaultDurationMinute: r.DefaultDurationMinute,
	}
}
