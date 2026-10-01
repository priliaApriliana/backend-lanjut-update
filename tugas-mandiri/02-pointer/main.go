package main

import "fmt"

// =========================================================================
// FUNCTION DENGAN POINTER
// =========================================================================

// swap menukar nilai dua integer melalui pointer.
// Parameter a dan b bertipe *int (pointer yang menyimpan alamat memori integer).
// Operator dereference (*) digunakan untuk membaca dan menimpa nilai langsung pada alamat memori tersebut.
func swap(a, b *int) {
	*a, *b = *b, *a
}

// updateSlice menambahkan item baru ke slice melalui pointer ke slice (*[]string).
// Di Go, slice header terdiri dari: pointer ke backing array, length, dan capacity.
// Fungsi append() dapat merealokasi array baru jika kapasitas penuh dan memperbarui length/capacity.
// Jika dilewatkan tanpa pointer (*[]string), perubahan header tersebut hanya terjadi pada salinan lokal.
// Dengan pointer *[]string, pemanggil akan melihat item baru yang ditambahkan secara permanen.
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

// =========================================================================
// PERBANDINGAN: PASS BY VALUE VS PASS BY POINTER
// =========================================================================

// swapGagal (Pass by Value):
// Menerima SALINAN dari nilai integer. Menukar salinan di dalam fungsi
// TIDAK memengaruhi variabel asli milik pemanggil.
func swapGagal(a, b int) {
	a, b = b, a
}

// tambahByValue (Pass by Value):
// Hanya mengubah salinan lokal x. Variabel asli di pemanggil tidak terpengaruh.
func tambahByValue(x int) {
	x = x + 10
}

// tambahByPointer (Pass by Pointer):
// Menerima alamat memori. Dereference *x langsung memodifikasi variabel asli.
func tambahByPointer(x *int) {
	*x = *x + 10
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("TUGAS MANDIRI 3: POINTER")
	fmt.Println("==================================================")

	// 1. Uji function swap(a, b *int)
	x, y := 10, 20
	fmt.Println("\n--- 1. UJI FUNGSI swap(a, b *int) ---")
	fmt.Printf("Sebelum swap  : x = %d, y = %d\n", x, y)
	swap(&x, &y)
	fmt.Printf("Setelah swap   : x = %d, y = %d\n", x, y)

	// 2. Uji function updateSlice(s *[]string, newItem string)
	daftarMatkul := []string{"Algoritma", "Basis Data"}
	fmt.Println("\n--- 2. UJI FUNGSI updateSlice(s *[]string, newItem string) ---")
	fmt.Printf("Slice awal    : %v (panjang: %d)\n", daftarMatkul, len(daftarMatkul))
	updateSlice(&daftarMatkul, "Backend Lanjut")
	fmt.Printf("Setelah update: %v (panjang: %d)\n", daftarMatkul, len(daftarMatkul))

	// 3. Perbandingan Pass by Value vs Pass by Pointer
	fmt.Println("\n--- 3. PERBANDINGAN PASS BY VALUE VS PASS BY POINTER ---")

	// Kasus A: Swap nilai
	p, q := 100, 200
	fmt.Printf("[A] Sebelum swapGagal (by value)  : p = %d, q = %d\n", p, q)
	swapGagal(p, q)
	fmt.Printf("    Setelah swapGagal (by value)  : p = %d, q = %d (TIDAK BERUBAH)\n", p, q)

	fmt.Printf("    Sebelum swap (by pointer)     : p = %d, q = %d\n", p, q)
	swap(&p, &q)
	fmt.Printf("    Setelah swap (by pointer)     : p = %d, q = %d (BERHASIL DITUKAR)\n", p, q)

	// Kasus B: Manipulasi nilai integer
	val := 50
	fmt.Printf("\n[B] Nilai awal angka              : %d\n", val)
	tambahByValue(val)
	fmt.Printf("    Setelah tambahByValue (value) : %d (TETAP, tidak terpengaruh)\n", val)
	tambahByPointer(&val)
	fmt.Printf("    Setelah tambahByPointer(ptr)  : %d (BERUBAH, nilai asli dimutasi)\n", val)

	fmt.Println("==================================================")
}
