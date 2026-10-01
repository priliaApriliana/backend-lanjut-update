package main

import "fmt"

// Student merepresentasikan data mahasiswa
type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// GetInfo mengembalikan informasi lengkap student dalam satu baris teks.
// RECEIVER: Value Receiver (s Student)
// ALASAN:
// Method ini hanya membaca data (read-only) untuk diformat menjadi teks,
// tanpa mengubah state atau memodifikasi field apa pun pada struct.
// Penggunaan value receiver aman, idiomatik di Go, dan menjamin struct asli tidak termutasi.
func (s Student) GetInfo() string {
	statusStr := "Aktif"
	if !s.IsActive {
		statusStr = "Tidak Aktif"
	}
	return fmt.Sprintf("ID: %d | Nama: %-10s | Grade: %.2f | Status: %s", s.ID, s.Name, s.Grade, statusStr)
}

// UpdateGrade memperbarui nilai (Grade) mahasiswa.
// RECEIVER: Pointer Receiver (s *Student)
// ALASAN:
// Method ini memodifikasi state internal struct (field Grade).
// Jika menggunakan value receiver, Go akan menduplikasi objek Student (pass-by-value),
// sehingga perubahan nilai hanya terjadi pada salinan lokal di dalam method dan hilang setelah selesai.
// Dengan pointer receiver, mutasi langsung diterapkan pada alamat memori objek pemanggil.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate mengubah status IsActive menjadi true.
// RECEIVER: Pointer Receiver (s *Student)
// ALASAN:
// Method ini memutasi state objek asli (mengubah field IsActive menjadi true).
// Wajib menggunakan pointer receiver agar perubahan tersimpan permanen pada instance pemanggil.
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate mengubah status IsActive menjadi false.
// RECEIVER: Pointer Receiver (s *Student)
// ALASAN:
// Method ini memutasi state objek asli (mengubah field IsActive menjadi false).
// Wajib menggunakan pointer receiver agar perubahan tersimpan permanen pada instance pemanggil.
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("TUGAS MANDIRI 4: STRUCT & METHOD STUDENT")
	fmt.Println("==================================================")

	// 1. Inisialisasi struct Student baru
	mhs := Student{
		ID:       101,
		Name:     "Putri",
		Grade:    75.0,
		IsActive: true,
	}

	fmt.Println("\n--- [1] Kondisi Awal Mahasiswa ---")
	fmt.Println(mhs.GetInfo())

	// 2. Menguji method UpdateGrade (pointer receiver)
	fmt.Println("\n--- [2] Memperbarui Nilai Mahasiswa: UpdateGrade(91.5) ---")
	mhs.UpdateGrade(91.5)
	fmt.Println(mhs.GetInfo())

	// 3. Menguji method Deactivate (pointer receiver)
	fmt.Println("\n--- [3] Menonaktifkan Mahasiswa: Deactivate() ---")
	mhs.Deactivate()
	fmt.Println(mhs.GetInfo())

	// 4. Menguji method Activate (pointer receiver)
	fmt.Println("\n--- [4] Mengaktifkan Kembali Mahasiswa: Activate() ---")
	mhs.Activate()
	fmt.Println(mhs.GetInfo())

	fmt.Println("==================================================")
}
