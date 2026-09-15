package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost adalah cost factor untuk hashing password.
// Nilai 12 dipilih karena:
// - Minimal yang diminta tugas: 10
// - 12 memberikan keseimbangan antara keamanan dan kecepatan (~250ms)
// - Setiap kenaikan 1 menggandakan waktu komputasi
const bcryptCost = 12

// dummyHash dipakai oleh VerifyDummyPassword. Di-generate sekali saat
// package dimuat supaya VerifyDummyPassword tetap memakan waktu yang
// sama dengan VerifyPassword sungguhan — mencegah timing attack untuk
// membedakan "user tidak ada" vs "password salah".
var dummyHash []byte

func init() {
	// Hash dari string acak; hasilnya tidak penting, yang penting prosesnya
	// tetap memakan waktu bcrypt.
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-placeholder-never-match"), bcryptCost)
}

// HashPassword meng-hash password menggunakan bcrypt.
func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("gagal hashing password: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword membandingkan password plain dengan hash bcrypt.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// VerifyDummyPassword menjalankan operasi bcrypt yang sama mahalnya
// dengan VerifyPassword, tetapi hasilnya selalu false.
// Gunakan ini ketika user tidak ditemukan agar waktu respons tetap
// konsisten — penyerang tidak bisa membedakan "user tidak ada"
// dari "password salah" berdasarkan selisih waktu respons.
func VerifyDummyPassword(plain string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}

// RandomToken menghasilkan string acak kriptografis sepanjang 32 byte
// yang di-encode ke hex (64 karakter).
// Menggunakan crypto/rand, BUKAN math/rand, untuk keamanan.
func RandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("gagal generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// SHA256Hex menghasilkan hash SHA-256 dari string input, dikembalikan
// sebagai hex string. Dipakai untuk meng-hash refresh token sebelum
// disimpan ke database — karena refresh token sudah acak 32 byte,
// SHA-256 sudah cukup (tidak perlu bcrypt yang lambat).
func SHA256Hex(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
