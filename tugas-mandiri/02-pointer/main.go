package main

import "fmt"

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

func main() {
	// ============================================
	// TUGAS 3 – Pointer
	// ============================================
	fmt.Println("========== TUGAS 3 – POINTER ==========")

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
}
