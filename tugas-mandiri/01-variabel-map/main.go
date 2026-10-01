package main

import "fmt"

func main() {
	// =========================================================================
	// TUGAS MANDIRI 2: VARIABEL & STRUKTUR DATA
	// =========================================================================

	// 1. Deklarasi lima variabel dengan tipe berbeda:
	// - string : teks/karakter
	// - int    : bilangan bulat
	// - float64: bilangan desimal
	// - bool   : nilai kebenaran (true/false)
	// - slice  : daftar dinamis ([]string)
	var nama string = "Putri"
	var umur int = 20
	var ipk float64 = 3.85
	var isActive bool = true
	hobi := []string{"Membaca", "Coding", "Berenang"}

	fmt.Println("==================================================")
	fmt.Println("1. DEMONSTRASI LIMA VARIABEL BERBEDA TIPE")
	fmt.Println("==================================================")
	fmt.Printf("Nama (string)   : %s\n", nama)
	fmt.Printf("Umur (int)      : %d tahun\n", umur)
	fmt.Printf("IPK (float64)   : %.2f\n", ipk)
	fmt.Printf("Aktif (bool)    : %t\n", isActive)
	fmt.Printf("Hobi (slice)    : %v (len: %d, cap: %d)\n", hobi, len(hobi), cap(hobi))

	// 2. Map data mahasiswa (Key: Nama [string], Value: Nilai [float64])
	fmt.Println("\n==================================================")
	fmt.Println("2. STRUKTUR DATA MAP MAHASISWA")
	fmt.Println("==================================================")
	dataMahasiswa := make(map[string]float64)

	// a. Operasi Menambah Data ke Map
	dataMahasiswa["Putri"] = 92.5
	dataMahasiswa["Angelyna"] = 87.0
	dataMahasiswa["Adel"] = 78.5
	dataMahasiswa["Budi"] = 82.0

	fmt.Println("[A] Setelah menambah 4 data mahasiswa:")
	for k, v := range dataMahasiswa {
		fmt.Printf("    - %-10s : %.1f\n", k, v)
	}

	// b. Operasi Membaca dengan Pengecekan Keberadaan (comma-ok idiom)
	fmt.Println("\n[B] Pengecekan Keberadaan Kunci (comma-ok idiom):")

	// Kasus 1: Kunci ada
	if nilai, ada := dataMahasiswa["Putri"]; ada {
		fmt.Printf("    - Mahasiswa 'Putri' DITEMUKAN dengan nilai: %.1f\n", nilai)
	} else {
		fmt.Println("    - Mahasiswa 'Putri' TIDAK DITEMUKAN.")
	}

	// Kasus 2: Kunci tidak ada
	if nilai, ada := dataMahasiswa["Cledyan"]; ada {
		fmt.Printf("    - Mahasiswa 'Cledyan' DITEMUKAN dengan nilai: %.1f\n", nilai)
	} else {
		fmt.Println("    - Mahasiswa 'Cledyan' TIDAK DITEMUKAN dalam map.")
	}

	// c. Operasi Menghapus Data dari Map
	fmt.Println("\n[C] Menghapus data mahasiswa 'Adel':")
	delete(dataMahasiswa, "Adel")
	fmt.Printf("    - Apakah 'Adel' masih ada? ")
	if _, ada := dataMahasiswa["Adel"]; ada {
		fmt.Println("Masih ada")
	} else {
		fmt.Println("Sudah berhasil dihapus.")
	}

	// d. Operasi Menelusuri (Iterasi) Seluruh Isi Map
	fmt.Println("\n[D] Menelusuri seluruh data mahasiswa yang tersisa:")
	nomor := 1
	for namaMhs, nilaiMhs := range dataMahasiswa {
		fmt.Printf("    %d. %-10s -> Nilai: %.1f\n", nomor, namaMhs, nilaiMhs)
		nomor++
	}
	fmt.Println("==================================================")
}
