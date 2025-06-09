# Database Connection Optimization

## Masalah yang Ditemukan

### 1. **Koneksi Database Redundant**
- Beberapa controller membuat koneksi database baru di setiap request menggunakan `gorm.Open()`
- File yang bermasalah:
  - `controllers/Kuis.go` - fungsi `AddKuis()`
  - `controllers/Kelas_Pengguna.go` - fungsi `JoinKelas()`
  - `controllers/HasilKuis.go` - fungsi `SubmitJawaban()` dan `GetHasilKuis()`

### 2. **Tidak Ada Connection Pooling**
- Tidak ada konfigurasi untuk mengoptimalkan penggunaan koneksi database
- Setiap request membuat koneksi baru yang tidak efisien

### 3. **Kredensial Hardcoded**
- Kredensial database di-hardcode dalam kode
- Tidak menggunakan environment variables

## Solusi yang Diimplementasikan

### 1. **Singleton Pattern untuk Database Connection**
- Menggunakan global variable `database.DB` yang diinisialisasi sekali
- Semua controller menggunakan `database.GetDBConnection()` untuk mendapatkan koneksi yang sama

### 2. **Connection Pooling Configuration**
```go
// Konfigurasi connection pool yang optimal
sqlDB.SetMaxOpenConns(25)    // Maksimal 25 koneksi terbuka
sqlDB.SetMaxIdleConns(10)    // Maksimal 10 koneksi idle
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Lifetime koneksi 5 menit
sqlDB.SetConnMaxIdleTime(1 * time.Minute)  // Idle time 1 menit
```

### 3. **Environment Variables**
- Semua konfigurasi database dipindahkan ke file `.env`
- Menggunakan library `github.com/joho/godotenv` untuk load environment variables

### 4. **Prepared Statement Caching**
- Mengaktifkan `PrepareStmt: true` di GORM config untuk cache prepared statements

### 5. **Optimized GORM Configuration**
```go
gormConfig := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Silent), // Mengurangi overhead logging
    PrepareStmt: true, // Enable prepared statement caching
    DisableForeignKeyConstraintWhenMigrating: false,
}
```

## Perubahan File

### 1. **database/database.go**
- ✅ Ditambahkan struct `Config` untuk konfigurasi database
- ✅ Ditambahkan fungsi `GetDatabaseConfig()` untuk membaca environment variables
- ✅ Ditambahkan connection pooling configuration
- ✅ Ditambahkan prepared statement caching
- ✅ Ditambahkan fungsi `InitializeDatabase()` dan `CloseDB()`

### 2. **main.go**
- ✅ Ditambahkan loading environment variables dengan `godotenv.Load()`
- ✅ Menggunakan `database.InitializeDatabase()` untuk inisialisasi
- ✅ Ditambahkan graceful shutdown untuk database connection
- ✅ Port aplikasi menggunakan environment variable

### 3. **controllers/Kuis.go**
- ✅ Menghapus `gorm.Open()` yang redundant
- ✅ Menggunakan `database.GetDBConnection()` untuk mendapatkan koneksi global
- ✅ Membersihkan import yang tidak diperlukan

### 4. **controllers/Kelas_Pengguna.go**
- ✅ Menghapus `gorm.Open()` yang redundant
- ✅ Menggunakan `database.GetDBConnection()` untuk mendapatkan koneksi global
- ✅ Membersihkan import yang tidak diperlukan

### 5. **controllers/HasilKuis.go**
- ✅ Menghapus `gorm.Open()` yang redundant di `SubmitJawaban()` dan `GetHasilKuis()`
- ✅ Menggunakan `database.GetDBConnection()` untuk mendapatkan koneksi global
- ✅ Membersihkan import yang tidak diperlukan

### 6. **Environment Files**
- ✅ Dibuat `.env.example` sebagai template
- ✅ Dibuat `.env` dengan konfigurasi production-ready

## Konfigurasi Environment Variables

```env
# Database Configuration
DB_HOST=hopper.proxy.rlwy.net
DB_PORT=27146
DB_USER=postgres
DB_PASSWORD=yBxKUopLCrVnBCpjpKdADHLGYMTYkKPC
DB_NAME=railway
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Jakarta

# Database Connection Pool Settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=300
DB_CONN_MAX_IDLE_TIME=60

# Application Settings
PORT=8000
ENV=production

# JWT Settings
JWT_SECRET=secret
```

## Manfaat Optimasi

### 1. **Pengurangan Koneksi Database**
- Dari membuat koneksi baru di setiap request → menggunakan 1 koneksi global dengan pooling
- Mengurangi overhead pembuatan dan penutupan koneksi

### 2. **Improved Performance**
- Connection pooling mengurangi latency
- Prepared statement caching mempercepat query execution
- Reduced memory usage

### 3. **Better Resource Management**
- Kontrol yang lebih baik terhadap jumlah koneksi database
- Automatic connection cleanup dengan lifetime dan idle time
- Graceful shutdown untuk mencegah connection leaks

### 4. **Scalability**
- Dapat menangani lebih banyak concurrent requests
- Database tidak akan kewalahan dengan terlalu banyak koneksi
- Konfigurasi yang dapat disesuaikan melalui environment variables

### 5. **Security & Maintainability**
- Kredensial database tidak lagi hardcoded
- Mudah untuk mengubah konfigurasi tanpa rebuild aplikasi
- Environment-specific configuration

## Cara Menjalankan

1. **Install dependencies:**
```bash
go mod tidy
```

2. **Setup environment variables:**
```bash
cp .env.example .env
# Edit .env sesuai dengan konfigurasi database Anda
```

3. **Run aplikasi:**
```bash
go run main.go
```

4. **Untuk seeding database:**
```bash
go run main.go seed
```

## Monitoring

Aplikasi sekarang akan menampilkan log informasi tentang konfigurasi database saat startup:
```
Database connected successfully with 25 max open connections and 10 max idle connections
Server starting on port 8000
```

## Optimasi Endpoint Hasil Kuis

### Masalah Sebelumnya
- Frontend harus memanggil endpoint `/hasil-kuis/:user_id/:kuis_id` untuk setiap kuis
- Jika ada 10 kuis, maka akan ada 10 request ke database
- Ini menyebabkan database cepat mencapai limit dan tidak efisien

### Solusi yang Diimplementasikan

#### 1. **Endpoint Baru yang Efisien**
```go
// GET /hasil-kuis/my-results
// Mengambil SEMUA hasil kuis user dalam 1 request
func GetAllHasilKuisByUser(c *fiber.Ctx) error {
    // Authenticate user
    user, err := Authenticate(c)

    // Get all quiz results with related quiz info in ONE query
    var hasilKuisList []models.Hasil_Kuis
    db.Preload("Kuis").Where("users_id = ?", user.ID).Find(&hasilKuisList)

    return sendResponse(c, fiber.StatusOK, true, "All quiz results retrieved successfully", hasilKuisList)
}
```

#### 2. **Routes yang Dioptimasi**
```go
// OPTIMIZED: Get all quiz results for authenticated user in one request
result.Get("/my-results", controllers.GetAllHasilKuisByUser)

// Admin/Teacher: Get all quiz results for specific user
result.Get("/user/:user_id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.GetHasilKuisByUserID)

// Legacy endpoint (kept for backward compatibility)
result.Get("/:user_id/:kuis_id", controllers.GetHasilKuis)
```

### Perbandingan Performance

#### **Sebelum Optimasi:**
- 10 kuis = 10 request ke database
- 10 koneksi database terpisah
- Latency tinggi karena multiple round trips

#### **Setelah Optimasi:**
- 10 kuis = 1 request ke database
- 1 koneksi database dengan JOIN
- Latency rendah dengan single query

### Cara Penggunaan Frontend

#### **Sebelum (Tidak Efisien):**
```javascript
// BAD: Multiple requests
const results = [];
for (const kuis of kuisList) {
    const response = await fetch(`/hasil-kuis/${userId}/${kuis.id}`);
    results.push(await response.json());
}
```

#### **Setelah (Efisien):**
```javascript
// GOOD: Single request
const response = await fetch('/hasil-kuis/my-results');
const allResults = await response.json();
```

## Backend Code Fixes & Improvements

### Masalah yang Diperbaiki

#### 1. **Model Structure Issues**
- ✅ Menghapus field `id` redundant yang konflik dengan `gorm.Model`
- ✅ `gorm.Model` sudah include `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`

#### 2. **Database Function Naming**
- ✅ Mengubah `GetallTasks()` menjadi `GetKategori()` untuk konsistensi
- ✅ Memperbaiki pemanggilan fungsi di controller

#### 3. **Authentication Consistency**
- ✅ Menambahkan authentication yang hilang di semua GET endpoints
- ✅ Semua controller sekarang menggunakan `Authenticate(c)` secara konsisten

#### 4. **Response Standardization**
- ✅ Menggunakan `sendResponse()` dan `handleError()` helper functions
- ✅ Memperbaiki response di `GetKelasByUserID` untuk konsistensi
- ✅ Mengoptimasi query dengan `Preload()` untuk mengurangi database calls

#### 5. **Message Consistency**
- ✅ Mengubah "Task" menjadi "Category" di kategori endpoints
- ✅ Pesan response yang lebih deskriptif dan konsisten

### File yang Diperbaiki

#### **models/models.go**
```go
// BEFORE: Redundant id field
type Users struct {
    gorm.Model
    id       uint   `gorm:"primaryKey"` // ❌ Redundant
    Name     string `json:"name"`
    // ...
}

// AFTER: Clean model structure
type Users struct {
    gorm.Model  // Already includes ID, CreatedAt, UpdatedAt, DeletedAt
    Name     string `json:"name"`
    // ...
}
```

#### **database/kategori.go**
```go
// BEFORE: Inconsistent function name
func GetallTasks() ([]models.Kategori_Soal, error) // ❌

// AFTER: Consistent naming
func GetKategori() ([]models.Kategori_Soal, error) // ✅
```

#### **controllers/kategori.go**
```go
// BEFORE: Missing authentication
func GetKategori(c *fiber.Ctx) error {
    // No authentication ❌
    result, err := database.GetKategori()
    // ...
}

// AFTER: Proper authentication
func GetKategori(c *fiber.Ctx) error {
    _, err := Authenticate(c) // ✅
    if err != nil {
        return err
    }
    result, err := database.GetKategori()
    // ...
}
```

#### **controllers/Kelas_Pengguna.go**
```go
// BEFORE: Inefficient queries + inconsistent response
func GetKelasByUserID(c *fiber.Ctx) error {
    // Multiple separate queries ❌
    for _, kp := range kelasPengguna {
        var kelas models.Kelas
        err := database.DB.Where("id = ?", kp.Kelas_id).First(&kelas).Error
        // ...
    }
    // Custom response format ❌
    return c.Status(fiber.StatusOK).JSON(fiber.Map{...})
}

// AFTER: Optimized with Preload + consistent response
func GetKelasByUserID(c *fiber.Ctx) error {
    // Single query with Preload ✅
    var kelasPengguna []models.Kelas_Pengguna
    if err := db.Preload("Kelas").Where("users_id = ?", user.ID).Find(&kelasPengguna).Error; err != nil {
        return handleError(c, err, "Failed to get user classes")
    }
    // Consistent response helper ✅
    return sendResponse(c, fiber.StatusOK, true, "User classes retrieved successfully", kelasList)
}
```

### Performance Improvements

#### **Database Query Optimization**
- ✅ Menggunakan `Preload("Kelas")` untuk mengurangi N+1 query problem
- ✅ Single query dengan JOIN vs multiple separate queries

#### **Connection Management**
- ✅ Konsisten menggunakan `database.GetDBConnection()` di semua controller
- ✅ Tidak ada lagi koneksi database redundant

### Security Improvements

#### **Authentication Coverage**
- ✅ Semua endpoints sekarang memiliki authentication
- ✅ Konsisten menggunakan JWT token validation
- ✅ Proper error handling untuk unauthorized access

### Code Quality Improvements

#### **Error Handling**
- ✅ Konsisten menggunakan `handleError()` helper
- ✅ Proper error messages yang deskriptif

#### **Response Format**
- ✅ Semua response menggunakan format yang sama:
```json
{
  "success": true/false,
  "message": "Descriptive message",
  "data": actual_data
}
```

## Rekomendasi Lanjutan

1. **Database Monitoring**: Implementasikan monitoring untuk connection pool metrics
2. **Health Check**: Tambahkan endpoint health check untuk database connection
3. **Circuit Breaker**: Implementasikan circuit breaker pattern untuk database failures
4. **Read Replicas**: Pertimbangkan menggunakan read replicas untuk query read-heavy
5. **Caching**: Implementasikan Redis atau in-memory caching untuk data yang sering diakses
6. **Pagination**: Tambahkan pagination untuk hasil kuis jika data sangat besar
7. **Indexing**: Pastikan ada index pada kolom users_id di tabel hasil_kuis
8. **Input Validation**: Tambahkan validation middleware untuk request body
9. **Rate Limiting**: Implementasikan rate limiting untuk mencegah abuse
10. **Logging**: Tambahkan structured logging untuk monitoring dan debugging
