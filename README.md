# go-backend-101

## run
```bash
make tidy
make run

Endpoint API
Berikut adalah daftar endpoint yang tersedia dalam API ini.

🏥 Health Check
GET /health

Deskripsi: Memeriksa status kesehatan layanan. Berguna untuk memastikan layanan berjalan dengan baik.

👤 Endpoint Pengguna (/v1/users)
Mengambil semua pengguna
GET /v1/users

Deskripsi: Mengembalikan daftar semua pengguna yang tersimpan di database.

Membuat pengguna baru
POST /v1/users

Deskripsi: Membuat satu entri pengguna baru.

Body Request (JSON):

JSON

{
  "name": "string",
  "email": "string"
}
Mengambil satu pengguna
GET /v1/users/{id}

Deskripsi: Mengambil detail satu pengguna berdasarkan id uniknya.

Memperbarui pengguna
PUT /v1/users/{id}

Deskripsi: Memperbarui data pengguna yang sudah ada berdasarkan id-nya.

Body Request (JSON):

JSON

{
  "name": "string",
  "email": "string"
}
Menghapus pengguna
DELETE /v1/users/{id}

Deskripsi: Menghapus data pengguna berdasarkan id-nya.

Contoh Penggunaan (cURL)
Berikut adalah beberapa contoh cara berinteraksi dengan API menggunakan cURL.

1. Memeriksa Status Kesehatan
Bash

curl -s http://localhost:8080/health
2. Membuat Pengguna Baru
Bash

curl -s -X POST http://localhost:8080/v1/users \
 -H "Content-Type: application/json" \
 -d '{"name":"Alea","email":"alea@example.com"}'
3. Melihat Semua Pengguna
Bash

curl -s http://localhost:8080/v1/users

