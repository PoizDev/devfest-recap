<p align="center">
  <img src="https://i.hizliresim.com/7ndpq83.png" alt="GDG Bursa Logo" height="80">
</p>

<h1 align="center">GDG Bursa — DevFest Recap</h1>

<p align="center">
  <code>recap.devfestbursa.com</code> için geliştirilmiş, yüksek performanslı bir katılımcı etkileşimi, karekod tarama, liderlik tablosu ve kişiselleştirilmiş "Wrapped" analiz platformudur. Etkinlik boyunca stantları gezen, oturumlara ve atölyelere katılan katılımcıların QR kodları taratarak puan toplamasını, sıralamada yarışmasını ve günün sonunda kendilerine özel istatistik kartları (dominant kategoriler, favori salonlar, kazanılan rozetler) üretilmesini sağlar.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Gin-v1.12.0-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Gin">
  <img src="https://img.shields.io/badge/PostgreSQL-15+-316192?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Redis-7.0+-DC382D?style=for-the-badge&logo=redis&logoColor=white" alt="Redis">
  <img src="https://img.shields.io/badge/Docker-Distroless-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

---

## 🌟 Öne Çıkan Özellikler

- ⚡ **Yüksek Performans & Önbellekleme:** Katılımcıların "Wrapped" (Recap) analizleri yoğun yük altında veri tabanını yormamak adına Redis üzerinde **60 saniyelik TTL** ile önbelleğe alınır.
- 🏆 **Gerçek Zamanlı Liderlik Tablosu (Leaderboard):** Katılımcı puanları anlık olarak Redis **Sorted Set (ZSet)** yapısında güncellenir. Redis'in pasif olduğu senaryolarda sistem otomatik olarak veri tabanı yedeğine (**fallback**) geçer.
- 🔄 **Otomatik DevTV Arka Plan Senkronizasyonu:** Tanımlanan `DEVTV_DSN` üzerinden ana etkinlik veritabanı (DevTV) ile entegre çalışır. Her 30 saniyede bir konuşmacıları, salonları ve zaman dilimlerini otomatik çekerek kategori tespiti yapar ve sisteme dahil eder.
- 🛡️ **Gelişmiş Güvenlik ve Hız Sınırı (Rate Limiting):** Brute-force ve spam tarama girişimlerini önlemek için IP tabanlı **Redis Sliding Window Rate Limiter** devrededir.
- 🔑 **Çift Katmanlı Kimlik Doğrulama:** JWT token doğrulaması hem HTTP Header hem de güvenli **HttpOnly çerez (Cookie)** üzerinden gerçekleştirilir.
- 🐳 **Güvenli Konteyner Mimarisi (Distroless):** Production ortamı için shell içermeyen, izole edilmiş `nonroot` kullanıcısı ile çalışan ultra hafif ve güvenli Docker imajı barındırır.

---

## 🛠️ Kullanılan Teknolojiler

| Teknoloji | Versiyon | Kullanım Amacı |
|-----------|----------|----------------|
| **Go** | 1.26.3 | Yüksek performanslı ana uygulama dili |
| **Gin Gonic** | 1.12.0 | HTTP REST API framework |
| **PostgreSQL** | 15 (Alpine) | İlişkisel veri tabanı (Kullanıcılar, seanslar ve taramalar) |
| **Redis** | 7.0 (Alpine) | Önbellek, Liderlik Tablosu ve Hız Sınırlandırıcı |
| **GORM** | 1.31.1 | Veri tabanı ORM kütüphanesi (JSON serializer desteğiyle) |
| **JWT** | v5 | Güvenli kullanıcı kimlik doğrulaması |
| **Zap Logger** | 1.28.0 | Yapılandırılmış yüksek performanslı log mekanizması |
| **Lumberjack** | v2 | Log dosyalarının otomatik döndürülmesi (Rotation) |

---

## 📁 Proje Yapısı

```text
├── .air.toml             # Go için hot-reload geliştirme yapılandırması
├── compose.yaml          # Docker Compose (App, Postgres, Redis) servisleri
├── conf.yaml             # Uygulama genel konfigürasyonu (Port, DB Limitleri, Hız sınırları)
├── Dockerfile            # Multi-stage & Distroless Docker derleme reçetesi
├── main.go               # Uygulama giriş noktası ve rota tanımlamaları
├── config/               # Yapılandırma yükleme ve Zap Logger ayarları
├── controllers/          # HTTP isteklerini karşılayan API denetleyicileri
├── db/                   # DB bağlantıları, Redis entegrasyonu, migration ve seed işlemleri
├── middlewares/          # Auth, Admin kontrolü, Rate Limiter ve QR kod üreteç katmanları
├── models/               # GORM veri modelleri (Users, Session, Scan)
├── services/             # İş mantığı, Liderlik tablosu yönetimi ve DevTV senkronizasyonu
├── utils/                # Yardımcı fonksiyonlar (bcrypt şifre hashleme vb.)
└── logs/                 # Uygulama çalıştırma günlükleri
```

---

## 🚀 Kurulum ve Çalıştırma

### A) Docker Compose ile Kurulum (Önerilen)

Sistemi veri tabanları ve bağımlılıklarıyla birlikte tek komutla production-ready çalıştırmak için Docker Compose en pratik yöntemdir.

#### 1. Ortam Değişkenlerini Tanımlayın

Projenin ana dizininde bir `.env` dosyası oluşturun ve aşağıdaki değişkenleri kendinize göre güncelleyin:

```dotenv
# DATABASE YAPILANDIRMASI
DB_USER=postgres
DB_PASSWORD=your_secure_db_password
DB_NAME=devfest_recap

# GO UYGULAMASI DSN BAĞLANTISI (Docker içi)
dsn=host=db user=postgres password=your_secure_db_password dbname=devfest_recap port=5432 sslmode=disable

# REDIS YAPILANDIRMASI
REDIS_ADDR=redis:6379
REDIS_PASSWORD=

# GÜVENLİK
JWT_SECRET=bursanin_en_guvenli_ve_uzun_jwt_secret_anahtari_1923!

# OTOMATİK OLUŞTURULACAK ADMIN HESABI
DEFAULT_ADMIN_USER=admin
DEFAULT_ADMIN_PASS=AdminSifreniz123!
DEFAULT_ADMIN_MAIL=admin@devfestbursa.com

# DEVTV ENTEGRASYONU (Opsiyonel)
# DEVTV_DSN=host=host.docker.internal user=devtv password=devtv_db_password dbname=devtv port=5432 sslmode=disable
```

#### 2. Servisleri Başlatın

Aşağıdaki komut yardımıyla PostgreSQL, Redis ve Go API servislerini arka planda ayağa kaldırabilirsiniz:

```bash
docker compose up -d
```

> [!NOTE]
> `app` servisi, PostgreSQL ve Redis servislerinin tam olarak hazır ve sağlıklı duruma gelmesini (`service_healthy` durum kontrolü) otomatik olarak bekler ve ardından başlatılır.

#### 3. Logları Takip Edin

```bash
docker compose logs -f app
```

#### 4. Kapatmak İçin

```bash
docker compose down
```

---

### B) Standart Yerel Kurulum (Go & Yerel Servisler)

Uygulamayı kendi bilgisayarınızda Go komutlarıyla yerel olarak koşturmak istiyorsanız:

#### 1. Bağımlılıkları İndirin

```bash
go mod download
go mod tidy
```

#### 2. Konfigürasyonu Güncelleyin

`conf.yaml` dosyasını yerel servislerinize göre yapılandırın:

```yaml
server: 
  port: ":1923"
  shutdown_timeout: "30s"
  active_level: "dev"
  env_path: ".env"
  recap_mode: false # Etkinlik bitmeden recap sonuçlarını açmak için true yapın

database:
  max_idle_conns: 25
  max_open_conns: 45
  conn_max_lifetime: "5m"
  conn_max_idle_time: "1m"
  conn_timeout: "15s"
  ping_period: "30s"

redis:
  host: "localhost"
  port: 6379
  db: 0
  pool_size: 10
  min_idle_conns: 3

jwt:
  expiration: 24h

rate_limit:
  rpm: 100
  burst_size: 10
  window_duration: 1m
```

#### 3. `.env` Dosyanızı Yerel Adreslere Göre Ayarlayın

Eğer lokalinizdeki servisleri kullanacaksanız `.env` dosyasındaki `dsn` ve `REDIS_ADDR` adreslerini güncelleyin:

```dotenv
dsn=host=localhost user=postgres password=your_secure_db_password dbname=devfest_recap port=3924 sslmode=disable
REDIS_ADDR=localhost:8389
# Lütfen DEFAULT_ADMIN_PASS değerini tanımlamayı unutmayın!
```

#### 4. Uygulamayı Başlatın

```bash
go run main.go
```

---

## 🗺️ Veri Modelleri ve İlişkileri

Sistemin kalbinde yer alan üç ana veri modelinin (`Users`, `Session`, `Scan`) yapısı ve aralarındaki ilişkiler şu şekildedir:

```
                  ┌─────────────────────────────────────────┐
                  │             DEVTV DATABASE              │
                  ├─────────────────────────────────────────┤
                  │ 1. facilitators (Konuşmacılar)          │
                  │ 2. workshops (Salonlar/Atölyeler)       │
                  │ 3. workshop_time_slots (Zaman Dilimleri)│
                  └───────────────────┬─────────────────────┘
                                      │
                                      ▼ [Arka Plan Oto-Senkronizasyon]
                  ┌─────────────────────────────────────────┐
                  │                SESSIONS                 │
                  ├─────────────────────────────────────────┤
                  │ id (PK)         : uint                  │
                  │ name            : string                │
                  │ description     : text                  │
                  │ type            : string                │
                  │ category        : string                │
                  │ points          : int                   │
                  │ qr_code_key (UQ): string                │
                  │ tags            : []string              │
                  │ room_name       : string                │
                  │ speaker_name    : string                │
                  │ speaker_title   : string                │
                  └───────────────────┬─────────────────────┘
                                      │ 1
                                      │
                                      │ 1..* (Çoka Çok Bağlantı)
                                      ▼
                  ┌─────────────────────────────────────────┐
                  │                  SCANS                  │
                  ├─────────────────────────────────────────┤
                  │ id (PK)         : uint                  │
                  │ user_id (FK)    : uint       ───────┐   │
                  │ session_id (FK) : uint       ─────┐ │   │
                  │ scanned_at      : time.Time       │ │   │
                  └───────────────────────────────────┼─┼───┘
                                                      │ │
                                      *..1 (İlişki)   │ │
                                 ┌────────────────────┘ │
                                 ▼                      │
                  ┌─────────────────────────────────────────┐
                  │                  USERS                  │
                  ├─────────────────────────────────────────┤
                  │ id (PK)         : uint                  │
                  │ username        : string                │
                  │ full_name       : string                │
                  │ mail (Unique)   : string                │
                  │ password        : string                │
                  │ role            : string                │
                  └─────────────────────────────────────────┘
```

---

## 📡 API Dokümantasyonu

### **Base URL**

```text
http://localhost:1923
```

---

### 1. Kimlik Doğrulama ve Hesap Yönetimi (Auth)

#### **Katılımcı Kaydı (Register)**

Kullanıcıların sisteme kaydolmasını sağlar. Güvenlik nedeniyle üyelik aşamasında `admin` rolü verilmesi engellenmiştir, tüm yeni kayıtlar `user` rolüyle başlar.

- **URL:** `/api/auth/register`
- **Method:** `POST`
- **Headers:** `Content-Type: application/json`
- **Request Body:**

```json
{
  "username": "coder_bursa",
  "full_name": "Ahmet Yılmaz",
  "mail": "ahmet@devfestbursa.com",
  "password": "guvenlisifre123"
}
```

- **Response (201 Created):**

```json
{
  "message": "Kayıt başarılı",
  "user_id": 2
}
```

#### **Katılımcı Girişi (Login)**

Kullanıcı adı ve şifre doğrulanarak JWT token üretilir. Bu token hem JSON body içinde döndürülür hem de güvenli bir **HttpOnly** çerez olarak tarayıcıya eklenir.

- **URL:** `/api/auth/login`
- **Method:** `POST`
- **Headers:** `Content-Type: application/json`
- **Request Body:**

```json
{
  "username": "coder_bursa",
  "password": "guvenlisifre123"
}
```

- **Response (200 OK):**

```json
{
  "message": "Giriş başarılı",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MiwidXNlcm5hbWUiOiJjb2Rlcl9idXJzYSIsIm1haWwiOiJhaG1ldEBkZXZmZXN0YnVyc2EuY29tIiwicm9sZSI6InVzZXIiLCJleHAiOjE3ODE5NzEyMDB9..."
}
```

- **Response Cookies:**

```text
Auth=eyJhbGciOiJI...; Path=/; Max-Age=604800; HttpOnly; SameSite=Lax
```

#### **Çıkış İşlemi (Logout)**

Çerezde saklanan `Auth` JWT token'ını geçersiz kılarak oturumu sonlandırır.

- **URL:** `/api/auth/logout`
- **Method:** `POST`
- **Response (200 OK):**

```json
{
  "message": "Çıkış başarılı"
}
```

---

### 2. Katılımcı Etkileşim API'leri (Giriş Yetkisi Gerekir)

> [!IMPORTANT]
> Bu gruptaki endpoint'ler için isteğin Header kısmında `Authorization: Bearer <token>` bulunmalı veya Cookie'lerinde geçerli bir `Auth=<token>` değeri yer almalıdır.

#### **QR Kod Tarama (Scan)**

Katılımcıların stantlarda veya salonlarda gördükleri karekodları okutmasını sağlar. Taramayla seansın puanı katılımcıya eklenir, liderlik tablosu güncellenir ve katılımcının recap önbelleği silinir. Her QR kod yalnızca bir kez okutulabilir.

- **URL:** `/api/scan`
- **Method:** `POST`
- **Request Body:**

```json
{
  "qr_code_key": "devtv_slot_12"
}
```

- **Response (201 Created):**

```json
{
  "message": "Seans başarıyla okutuldu!",
  "scan": {
    "id": 1,
    "user_id": 2,
    "session_id": 5,
    "scanned_at": "2026-05-23T17:02:15.123+03:00",
    "session": {
      "id": 5,
      "name": "Go ile Microservice Mimarisi",
      "description": "Büyük ölçekli sistemlerde Go diliyle mikroservis tasarımı.",
      "type": "session",
      "category": "backend",
      "points": 15,
      "qr_code_key": "devtv_slot_12",
      "tags": ["go", "microservices", "grpc"],
      "room_name": "Salih Tozan Salonu",
      "speaker_name": "Caner Hızlı",
      "speaker_title": "Senior Backend Architect"
    }
  }
}
```

- **Hata Senaryosu (400 Bad Request - Zaten Okutulmuş):**

```json
{
  "error": "bu seansı daha önce zaten okuttunuz"
}
```

#### **Kullanıcı Profili (Me)**

Oturum açmış kullanıcının profil bilgilerini getirir. Uygulamanın ön yüzünde kullanıcı bilgilerini göstermek için kullanılır.

- **URL:** `/api/me`
- **Method:** `GET`
- **Response (200 OK):**

```json
{
  "id": 2,
  "username": "coder_bursa",
  "full_name": "Ahmet Yılmaz",
  "mail": "ahmet@devfestbursa.com",
  "role": "user"
}
```

- **Hata Senaryosu (404 Not Found):**

```json
{
  "error": "Kullanıcı bulunamadı"
}
```

#### **Katılımcı Sıralama Tablosu (Leaderboard)**

En yüksek puanlı ilk 10 katılımcıyı getirir. Liderlik tablosu performansı artırmak için Redis Sorted Set üzerinden çekilir. Redis devre dışı kalırsa veri tabanından anlık hesaplanarak fallback mekanizması işletilir.

- **URL:** `/api/leaderboard`
- **Method:** `GET`
- **Response (200 OK):**

```json
[
  {
    "rank": 1,
    "user_id": 4,
    "username": "gopher_master",
    "full_name": "Kaan Demir",
    "points": 185
  },
  {
    "rank": 2,
    "user_id": 2,
    "username": "coder_bursa",
    "full_name": "Ahmet Yılmaz",
    "points": 160
  }
]
```

#### **Wrapped / Recap Sonuçları (Recap)**

Katılımcının etkinlik boyundaki tüm geçmişini, istatistiklerini, favori konuşmacısını, dominant kategorisini ve kazandığı özel ünvanları/rozetleri getirir.

- **URL:** `/api/recap`
- **Method:** `GET`
- **Response (200 OK):**

```json
{
  "total_scans": 12,
  "total_points": 160,
  "rank": 2,
  "percentile": 94.5,
  "total_participants": 36,
  "category_stats": {
    "ai": 1,
    "backend": 7,
    "cloud": 3,
    "mobile": 1
  },
  "dominant_category": "Backend & Sunucu Teknolojileri",
  "favorite_room": "Salih Tozan Salonu",
  "favorite_speaker": "Caner Hızlı",
  "favorite_tag": "go",
  "total_talks": 6,
  "type_stats": {
    "session": 6,
    "stand": 4,
    "workshop": 2
  },
  "tag_stats": {
    "docker": 2,
    "go": 5,
    "grpc": 3,
    "kubernetes": 1
  },
  "first_scan_time": "09:42",
  "last_scan_time": "16:15",
  "badges": [
    {
      "name": "İlk Adım",
      "description": "DevFest macerasına ilk QR kodunu taratarak başladın!",
      "icon": "🏁"
    },
    {
      "name": "Sistem Mimarı",
      "description": "Go, veritabanları ve ölçeklenebilirlik oturumlarının yıldızı oldun.",
      "icon": "⚙️"
    },
    {
      "name": "Erken Kuş",
      "description": "Sabahın erken saatlerinde etkinlikte yerini aldın!",
      "icon": "🐦"
    }
  ],
  "scanned_sessions": [
    {
      "name": "Go ile Microservice Mimarisi",
      "type": "session",
      "category": "backend",
      "points": 15,
      "room_name": "Salih Tozan Salonu",
      "speaker_name": "Caner Hızlı",
      "speaker_title": "Senior Backend Architect",
      "tags": ["go", "microservices", "grpc"],
      "scanned_at": "11:20"
    }
  ]
}
```

> [!NOTE]
> **Kazanılabilir Rozetler:**
>
> | Rozet | Koşul | İkon |
> |-------|-------|------|
> | İlk Adım | İlk QR tarama (≥1 tarama) | 🏁 |
> | Keşifçi | 5 veya daha fazla tarama | 🔍 |
> | Koleksiyoner | 10 veya daha fazla tarama | 🎯 |
> | Maratoncu | 15 veya daha fazla tarama | 🏃 |
> | Efsane Katılımcı | 20 veya daha fazla tarama | 👑 |
> | Stant Avcısı | 3 veya daha fazla stant ziyareti | 🎪 |
> | Atölye Ustası | 2 veya daha fazla atölye katılımı | 🔧 |
> | Erken Kuş | İlk tarama saat 10:00'dan önce | 🐦 |
> | Gece Baykuşu | Son tarama saat 17:00'den sonra | 🦉 |
> | Sistem Mimarı | Dominant kategori: Backend | ⚙️ |
> | Yapay Zeka Kaşifi | Dominant kategori: AI | 🤖 |
> | Mobil Sihirbazı | Dominant kategori: Mobile | 📱 |
> | Bulut Yolcusu | Dominant kategori: Cloud | ☁️ |

> [!WARNING]
> Eğer `conf.yaml` içindeki `recap_mode` parametresi `false` ise, bu endpoint `403 Forbidden` hatası vererek verileri gizler:
>
> ```json
> { "error": "Recap sonuçları henüz açıklanmadı. Bizi izlemeye devam edin!" }
> ```

---

### 3. Sistem Durumu ve Diğer Rotalar

#### **Durum Bilgisi (Status)**

Sistemin genel sağlığını, servis bağlantılarını ve recap_mode durumunu kontrol etmek için kullanılır.

- **URL:** `/api/status`
- **Method:** `GET`
- **Response (200 OK - Sağlıklı):**

```json
{
  "status": "healthy",
  "recap_mode": true,
  "database": true,
  "redis": true
}
```

#### **Karekod Oluşturucu (Generate QR)**

Verilen herhangi bir metin/link değerini anında sunucu tarafında PNG karekoda dönüştürür. Query parametreleri ile boyut ayarlanabilir.

- **URL:** `/api/qrcode?text=GDGBursaRecap2026&size=300`
- **Method:** `GET`
- **Response:** `image/png` binary veri akışı.

---

### 4. Yönetici API'leri (Sadece Admin Yetkisiyle)

> [!IMPORTANT]
> Bu endpoint'ler için oturum açan kullanıcının rolü `admin` olmalıdır.

#### **Yeni Seans/Oturum Oluşturma**

Uygulama içine elle yeni bir QR kodlu stant veya oturum eklemek için kullanılır.

- **URL:** `/api/admin/sessions`
- **Method:** `POST`
- **Request Body:**

```json
{
  "name": "Google Cloud Standı",
  "description": "GCP standını ziyaret et ve bulut bulmacasını çöz!",
  "type": "stand",
  "category": "cloud",
  "points": 20,
  "qr_code_key": "gcp_stand_special"
}
```

- **Response (201 Created):**

```json
{
  "message": "Seans başarıyla oluşturuldu",
  "session": {
    "id": 16,
    "name": "Google Cloud Standı",
    "description": "GCP standını ziyaret et ve bulut bulmacasını çöz!",
    "type": "stand",
    "category": "cloud",
    "points": 20,
    "qr_code_key": "gcp_stand_special",
    "tags": null,
    "created_at": "2026-05-23T17:10:00Z",
    "updated_at": "2026-05-23T17:10:00Z"
  },
  "qr_code_url": "/api/admin/sessions/16/qrcode"
}
```

#### **Seans Karekodunu Çekme**

Oluşturulan bir seansın QR Kodunu doğrudan admin arayüzünde yazdırmak veya göstermek için PNG dosyası olarak çeker.

- **URL:** `/api/admin/sessions/:id/qrcode?size=512`
- **Method:** `GET`
- **Response:** `image/png` formatında seansa özel görsel.

#### **Tüm Recap Verilerini Toplu Dışa Aktarma**

Etkinlik sonunda kazananları belirlemek veya genel analitik raporlar oluşturmak için tüm kayıtlı katılımcıların Wrapped/Recap verilerini tek bir JSON dosyasında toplu sunar.

- **URL:** `/api/admin/export-recaps`
- **Method:** `GET`
- **Response (200 OK):**

```json
[
  {
    "user_id": 2,
    "username": "coder_bursa",
    "full_name": "Ahmet Yılmaz",
    "mail": "ahmet@devfestbursa.com",
    "recap": {
      "total_scans": 12,
      "total_points": 160,
      "rank": 2,
      "percentile": 94.5,
      "dominant_category": "Backend & Sunucu Teknolojileri",
      "favorite_room": "Salih Tozan Salonu",
      "favorite_speaker": "Caner Hızlı",
      "favorite_tag": "go"
    }
  }
]
```

#### **DevTV ile Anlık El İle Senkronizasyon (Manual Sync)**

Arka plan işini beklemek istemediğinizde DevTV veritabanından oturum ve konuşmacı bilgilerini anlık olarak senkronize etmek için tetiklenir.

- **URL:** `/api/admin/sync-from-devtv`
- **Method:** `POST`
- **Response (200 OK):**

```json
{
  "message": "DevTV veritabanı ile başarıyla senkronize edildi",
  "inserted": 2,
  "updated": 1,
  "fetched": 15,
  "total": 18
}
```

#### **Genel Analiz Paneli (Analytics Overview)**

Etkinlik katılım oranlarını ve genel QR tarama durumunu gösterir.

- **URL:** `/api/admin/analytics/overview`
- **Method:** `GET`
- **Response (200 OK):**

```json
{
  "total_registered_users": 154,
  "total_active_participants": 98,
  "total_scans": 642,
  "average_scans_per_user": 6.551
}
```

#### **Salon Yoğunluk Analizi (Analytics Rooms)**

Hangi salonun veya alanın daha fazla ziyaretçi/tarama çektiğini listeler.

- **URL:** `/api/admin/analytics/rooms`
- **Method:** `GET`
- **Response (200 OK):**

```json
[
  {
    "room_name": "Salih Tozan Salonu",
    "scans": 342
  },
  {
    "room_name": "Giriş Kat Fuaye Alanı",
    "scans": 180
  },
  {
    "room_name": "Mini Seminer Odası A",
    "scans": 120
  }
]
```

---

## 🔄 DevTV Entegrasyonu ve Akıllı Eşleştirme Detayları

### DevTV Nedir?

**DevTV**, `devtv.devfestbursa.com` için özel olarak geliştirilmiş yüksek performanslı etkinlik akışı, konuşmacı yönetimi ve oturum/zaman planlaması (schedule) platformudur. Projenin kaynak kodlarına [poizdev/devtv](https://github.com/poizdev/devtv) adresinden ulaşabilirsiniz.

### Neden Entegre Ettik?

DevFest Recap sisteminin, oturumlar, konuşmacılar ve salon isimleri gibi etkinlik verilerini manuel olarak tekrar girmek yerine, halihazırda bu verilerin canlı olarak yönetildiği DevTV sisteminden beslenmesi amaçlanmıştır. Bu sayede DevTV tarafında yapılan herhangi bir saat veya salon değişikliği, otomatik olarak Recap sistemine yansır ve veri tutarlılığı sağlanır.

Recap API, `DEVTV_DSN` aracılığıyla DevTV sisteminin PostgreSQL şemasına bağlanır ve aşağıdaki kurallara göre akıllı veri dönüşümü gerçekleştirir:

1. **Benzersiz QR Anahtarı:** Her atölye veya sunum zaman dilimi (`workshop_time_slots`), `devtv_slot_<slot_id>` şablonunda benzersiz bir `qr_code_key` kazanır.
2. **Akıllı Kategori Tespiti:** Sunum başlığı ve atölye adına göre akıllı bir metin taraması yapılarak oturumlar 4 ana etiket altında otomatik sınıflandırılır:
   - 🧠 **Yapay Zeka (ai):** İçinde `ai`, `gemini`, `yapay zeka`, `data`, `machine learning`, `model` geçen kelimeler.
   - 📱 **Mobil Uygulama (mobile):** İçinde `flutter`, `android`, `kotlin`, `mobile` geçen kelimeler.
   - ☁️ **Bulut ve DevOps (cloud):** İçinde `cloud`, `bulut`, `gcp`, `firebase`, `kubernetes`, `devops` geçen kelimeler.
   - ⚙️ **Sistem ve Backend (backend):** Go (Golang), Dart ve diğer tüm teknik/genel oturumlar için varsayılan kategori.
3. **Puanlama Standartı:** DevTV üzerinden senkronize edilen tüm sunum ve oturumlar varsayılan olarak **15 puan** olarak tanımlanır.
4. **Veri Tutarlılığı:** Oturum bilgileri veya konuşmacı unvanları DevTV tarafında güncellenirse, arka plan servisi bu değişimleri tespit edip verileri kayba uğratmadan günceller.

---

## 🛡️ Hız Sınırlandırma (Rate Limiting) Algoritması

Sistem, kötü niyetli tarama isteklerini engellemek amacıyla Redis tabanlı **Sliding Window Rate Limiting** algoritması kullanır.

- Her IP adresi için Redis üzerinde bir Sorted Set (`rl:<client_ip>`) tutulur.
- Gelen her istekte, belirlenen süreden (örneğin 1 dakika) daha eski olan zaman damgaları (`ZRemRangeByScore`) silinir.
- Yeni gelen isteğin zaman damgası set içine eklenir.
- Set içindeki eleman sayısı (`ZCard`) limit değerini (örneğin dakikada 100 istek) aşarsa istek `429 Too Many Requests` hata koduyla reddedilir.
- Redis'in durması veya yanıt vermemesi durumunda sistem **Fail-Open** (Hata Durumunda Geçir) ilkesine bağlı olarak çalışmaya kesintisiz devam eder.

---

## 🐳 Production Deployment & Docker Ayrıntıları

Dockerfile, **Multi-Stage Build** yöntemiyle tasarlanmıştır.

```dockerfile
# 1. Aşama: Uygulamanın alpine üzerinde derlenmesi
FROM golang:1.26-alpine AS build
...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o /bin/dfrecap .

# 2. Aşama: Sadece çalıştırılabilir dosyanın distroless'a taşınması
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/dfrecap /devfest-recap
...
```

**Distroless İmaj Avantajları:**

- Güvenlik açığı oluşturabilecek hiçbir shell komutu (`sh`, `bash`, `apk` vb.) barındırmaz.
- Saldırı yüzeyini en aza indirger.
- Sadece `nonroot` kullanıcısı yetkileriyle çalışarak konteynerin ana işletim sistemine sızma riskini yok eder.
- İmaj boyutu minimum seviyededir.

---

## 📞 İletişim ve Destek

Sistemle ilgili sorularınız veya bulduğunuz hatalar için lütfen GitHub Issues'de bir issue açınız.

GitHub: [https://github.com/poizdev/devfest-recap](https://github.com/poizdev/devfest-recap)

Mail: <musa@gdgbursa.com>

**Son Güncelleme:** Temmuz 11, 2026

---

## 📜 Lisans

Bu proje **MIT** lisansı altında lisanslanmıştır. Detaylar için [LICENSE](file:///C:/Users/poizd/Desktop/gdgbursatestleri/devfest-recap/LICENSE) dosyasını inceleyebilirsiniz.

---
