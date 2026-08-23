package main

import "fmt"

// Struct Student
type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// GetInfo hanya MEMBACA data, tidak mengubah apa pun,
// sehingga cukup pakai value receiver (Student, bukan *Student).
func (s Student) GetInfo() string {
	return fmt.Sprintf("[%d] %s | Nilai: %.1f | Aktif: %v", s.ID, s.Name, s.Grade, s.IsActive)
}

// UpdateGrade MENGUBAH data asli, jadi wajib pointer receiver (*Student).
// Kalau pakai value receiver, perubahan cuma terjadi di salinan
// dan hilang begitu function selesai.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate dan Deactivate juga mengubah data asli -> pointer receiver.
func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	// ============================================
	// TUGAS 4 – Struct & Method
	// ============================================
	fmt.Println("========== TUGAS 4 – STRUCT & METHOD ==========")

	// Membuat instance Student baru
	s1 := Student{
		ID:       1,
		Name:     "Putri",
		Grade:    0,
		IsActive: true,
	}

	fmt.Println("--- Kondisi awal ---")
	fmt.Println(s1.GetInfo())

	// Uji UpdateGrade
	s1.UpdateGrade(88.5)
	fmt.Println("\n--- Setelah UpdateGrade(88.5) ---")
	fmt.Println(s1.GetInfo())

	// Uji Deactivate
	s1.Deactivate()
	fmt.Println("\n--- Setelah Deactivate() ---")
	fmt.Println(s1.GetInfo())

	// Uji Activate
	s1.Activate()
	fmt.Println("\n--- Setelah Activate() ---")
	fmt.Println(s1.GetInfo())
}
