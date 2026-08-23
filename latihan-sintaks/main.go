package main

import "fmt"

// ============================================
// Struct & Method (dari Tugas 4)
// ============================================

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

// ============================================
// Function Pointer (dari Tugas 3)
// ============================================

// tukarNilai menukar nilai dua integer lewat pointer
func tukarNilai(a, b *int) {
	*a, *b = *b, *a
}

// updateSlice menambahkan item baru ke slice lewat pointer.
// Perlu *[]string karena append() bisa mengembalikan slice
// header baru kalau kapasitas lama sudah penuh — tanpa pointer,
// perubahan itu tidak akan terlihat oleh pemanggil function.
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

// --- Pembanding: pass by value vs pass by pointer ---
func tambahSepuluhByValue(x int) {
	x = x + 10 // hanya mengubah salinan lokal
}

func tambahSepuluhByPointer(x *int) {
	*x = *x + 10 // mengubah nilai asli lewat alamatnya
}

// ============================================
// MAIN
// ============================================
func main() {
	// ============================================
	// TUGAS 2 – Variabel & Struktur Data
	// ============================================
	fmt.Println("========== TUGAS 2 – VARIABEL & MAP ==========")

	// 1. Deklarasi 5 variabel dengan tipe berbeda
	var namaMhs string = "Putri"
	var umurMhs int = 20
	var ipk float64 = 3.75
	var aktif bool = true
	hobi := []string{"membaca", "coding", "berenang"}

	fmt.Println("--- Variabel ---")
	fmt.Println("Nama :", namaMhs)
	fmt.Println("Umur :", umurMhs)
	fmt.Println("IPK  :", ipk)
	fmt.Println("Aktif:", aktif)
	fmt.Println("Hobi :", hobi)

	// 2. Map data mahasiswa (nama = kunci, nilai = isi)
	nilaiMahasiswa := make(map[string]float64)

	// Menambah 
	nilaiMahasiswa["Putri"] = 90
	nilaiMahasiswa["Angelyna"] = 85
	nilaiMahasiswa["Adel"] = 78

	fmt.Println("\n--- Setelah menambah data ---")
	fmt.Println(nilaiMahasiswa)

	// Membaca dengan pengecekan keberadaan 
	if nilaiPutri, ada := nilaiMahasiswa["Putri"]; ada {
		fmt.Println("\nNilai Putri ditemukan:", nilaiPutri)
	} else {
		fmt.Println("\nPutri belum punya nilai")
	}

	if _, ada := nilaiMahasiswa["Cledyan"]; !ada {
		fmt.Println("Cledyan belum punya nilai (belum ada di map)")
	}

	// Menghapus 
	delete(nilaiMahasiswa, "Adel")
	fmt.Println("\n--- Setelah Adel dihapus ---")
	fmt.Println(nilaiMahasiswa)

	// Menelusuri seluruh isi map 
	fmt.Println("\n--- Menelusuri seluruh isi map ---")
	for namaMhs, nilaiMhs := range nilaiMahasiswa {
		fmt.Printf("%s: %.1f\n", namaMhs, nilaiMhs)
	}

	// ============================================
	// TUGAS 3 – Pointer
	// ============================================
	fmt.Println("\n========== TUGAS 3 – POINTER ==========")

	// Bukti tukarNilai() / swap()
	x, y := 1, 2
	fmt.Println("--- swap() ---")
	fmt.Println("Sebelum swap:", x, y)
	tukarNilai(&x, &y)
	fmt.Println("Setelah swap :", x, y)

	// Bukti updateSlice()
	daftarBuah := []string{"apel", "jeruk"}
	fmt.Println("\n--- updateSlice() ---")
	fmt.Println("Sebelum:", daftarBuah)
	updateSlice(&daftarBuah, "mangga")
	fmt.Println("Setelah:", daftarBuah)

	// Perbandingan pass by value vs pass by pointer
	fmt.Println("\n--- Pass by value vs pass by pointer ---")

	angka1 := 5
	tambahSepuluhByValue(angka1)
	fmt.Println("Pass by value   -> angka1 tetap  :", angka1) // tetap 5

	angka2 := 5
	tambahSepuluhByPointer(&angka2)
	fmt.Println("Pass by pointer -> angka2 berubah:", angka2) // jadi 15

	// Penjelasan singkat:
	// - Pass by value mengirim SALINAN nilai ke function.
	//   Perubahan di dalam function tidak memengaruhi variabel asli.
	// - Pass by pointer mengirim ALAMAT variabel asli.
	//   Perubahan lewat *x langsung mengubah nilai di alamat itu juga.

	// ============================================
	// TUGAS 4 – Struct & Method
	// ============================================
	fmt.Println("\n========== TUGAS 4 – STRUCT & METHOD ==========")

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
