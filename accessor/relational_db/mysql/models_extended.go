package mysql

// TODO: uncomment this, and adjust all invocations (including all entites, user, teacher, instrument, grade, etc.)
// func (r GetCoursesByIdsRow) ToGetCoursesRow() GetCoursesRow {
// 	return GetCoursesRow{
// 		CourseID:              r.CourseID,
// 		CourseName:            r.CourseName,
// 		DefaultFee:            r.DefaultFee,
// 		DefaultDurationMinute: r.DefaultDurationMinute,
// 	}
// }

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
