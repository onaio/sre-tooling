package audit

import "testing"

func TestCompareGrades(t *testing.T) {
	gradeTest1 := CompareGrades("B", "A+")
	if !gradeTest1 {
		t.Errorf("CompareGrades(\"B\", \"A+\") = %t; want true", gradeTest1)
	}

	gradeTest2 := CompareGrades("A", "C")
	if gradeTest2 {
		t.Errorf("CompareGrades(\"A\", \"C\") = %t; want false", gradeTest2)
	}

	gradeTest3 := CompareGrades("A", "A")
	if gradeTest3 {
		t.Errorf("CompareGrades(\"A\", \"A\") = %t; want false", gradeTest3)
	}

	gradeTest4 := CompareGrades("B+", "A-")
	if !gradeTest4 {
		t.Errorf("CompareGrades(\"B+\", \"A-\") = %t; want true", gradeTest4)
	}
}
