package main

import "fmt"

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
}
