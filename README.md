# Marketing Contact Analysis Dashboard

Dashboard untuk analisis data marketing contact dengan fitur import Excel, tracking transaksi, dan manajemen client.

## 🚀 Setup & Installation

### 1. Install Dependencies

```bash
npm install
# atau
yarn install
```

### 2. Setup Database

1. Buat file `.env.local` di root project:
```bash
DATABASE_URL="postgresql://postgres:admin@localhost:5432/phone_data"
AUTH_SECRET="generate-secret-with-openssl-rand-base64-32"
```

2. **PENTING**: Pastikan `.env.local` ada di root project (bukan di folder lain)
   - Prisma CLI membaca `.env` atau `.env.local` dari root project
   - Jika masih error, coba buat file `.env` (tanpa .local) untuk Prisma CLI

3. Generate Prisma Client:
```bash
npx prisma generate
```

4. Run database migration:
```bash
npx prisma migrate dev
# atau jika database sudah ada:
npx prisma db push
```

### 3. Setup Admin Pertama Kali

Setelah database siap, buat admin pertama kali:

```bash
curl -X POST http://localhost:3000/api/setup \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123","name":"Administrator"}'
```

Atau langsung insert ke database:
```sql
INSERT INTO "User" (email, password, name, role, "isActive") 
VALUES ('admin@example.com', '$2a$10$...', 'Administrator', 'admin', true);
```
(Password harus di-hash dengan bcrypt)

### 4. Run Development Server

```bash
npm run dev
# atau
yarn dev
```

Buka [http://localhost:3000](http://localhost:3000)

## 📁 Struktur Project

```
src/
├── app/                          # Next.js App Router
│   ├── (home)/                   # Dashboard home page
│   │   ├── _components/          # Components khusus home page
│   │   ├── fetch.ts              # Data fetching untuk home
│   │   └── page.tsx              # Home page
│   ├── admin/                    # Admin-only pages
│   │   └── users/                # Manage users page
│   ├── api/                      # API routes
│   │   ├── activity-logs/        # Activity log API
│   │   ├── clients/              # Client CRUD API
│   │   ├── data/bulk/            # Bulk import data API
│   │   ├── registration/bulk/    # Bulk import registration API
│   │   ├── transaction/bulk/     # Bulk import transaction API
│   │   ├── setup/                # Setup admin API
│   │   └── users/                # User management API
│   ├── auth/                     # Authentication pages
│   ├── client/                    # Client management page
│   ├── data/                     # Data listing page
│   ├── registration/             # Registration listing page
│   ├── transaction/              # Transaction listing page
│   └── layout.tsx                # Root layout
├── components/                   # Reusable components
│   ├── Charts/                   # Chart components
│   ├── Layouts/                  # Layout components (Header, Sidebar)
│   ├── Tables/                   # Table components
│   └── ui/                       # UI components
├── lib/                          # Utility libraries
│   ├── activity-log.ts           # Activity logging utility
│   ├── currency.ts               # Currency formatting
│   ├── pagination.ts             # Pagination utilities
│   └── prisma.ts                 # Prisma client singleton
├── services/                     # Business logic services
│   ├── analytics.ts              # Analytics calculations
│   └── charts.services.ts        # Chart data services
└── auth.ts                       # NextAuth configuration

prisma/
└── schema.prisma                 # Database schema
```

## 🔐 Authentication & Authorization

### Flow Authentication

1. User login via `/auth/sign-in`
2. Credentials divalidasi di `src/auth.ts` (authorize function)
3. Password di-verify dengan bcrypt
4. Jika valid, session dibuat dengan JWT
5. Middleware (`middleware.ts`) protect semua routes kecuali public routes

### Role-based Access

- **Admin**: Bisa akses semua halaman termasuk `/admin/users`
- **User**: Hanya bisa akses halaman umum (dashboard, registration, transaction, data, client)

### Public Routes

- `/auth/sign-in` - Login page
- `/api/auth/*` - NextAuth API
- `/api/setup` - Setup admin pertama kali

## 📊 Fitur Utama

### 1. Dashboard (`/`)
- Overview cards: Total Deposit, Total Profit, Registrations, Contacts
- Charts: Payments Overview, Weeks Profit, Clients Distribution, Top 10 Clients
- Table: Top Profit (by phone number)
- Filter: Date range dan Client filter

### 2. Registration (`/registration`)
- List semua registrasi dengan pagination
- Filter: Date range, Client (All/Organic/Specific Client)
- Bulk import dari Excel
- Download template Excel

### 3. Transaction (`/transaction`)
- List semua transaksi dengan pagination
- Filter: Date range
- Bulk import dari Excel
- Download template Excel
- Default filter: Hari ini

### 4. Data (`/data`)
- List semua data dengan relasi client
- Filter: Date range
- Bulk import dari Excel
- Download template Excel
- Default filter: Hari ini

### 5. Client (`/client`)
- List semua client
- Tambah client baru
- Form dengan success/error message

### 6. Manage Users (`/admin/users`) - Admin Only
- List semua users
- Tambah user/admin baru
- Lihat role dan status user

## 📝 Activity Log System

Sistem activity log mencatat semua insert operations:

- **Registration Bulk Import**: Mencatat jumlah data yang di-import
- **Transaction Bulk Import**: Mencatat jumlah transaksi yang di-import
- **Data Bulk Import**: Mencatat jumlah data yang di-import
- **Create Client**: Mencatat client yang dibuat
- **Create User**: Mencatat user yang dibuat

Activity log ditampilkan di notification bell di header.

## 🔧 API Endpoints

### Public APIs
- `POST /api/setup` - Setup admin pertama kali

### Authenticated APIs
- `GET /api/clients` - List clients
- `POST /api/clients` - Create client
- `POST /api/registration/bulk` - Bulk import registration
- `POST /api/transaction/bulk` - Bulk import transaction
- `POST /api/data/bulk` - Bulk import data
- `GET /api/activity-logs` - Get activity logs

### Admin-only APIs
- `GET /api/users` - List users
- `POST /api/users` - Create user/admin

## 🗄️ Database Schema

### Tables
- `registration` - Data registrasi (phone_number)
- `transaction` - Data transaksi (phone_number, transaction_date, total_deposit, total_profit)
- `data` - Data pengguna (whatsapp, name, nik, client_id)
- `client` - Data client
- `User` - User untuk authentication (email, password, role)
- `activity_log` - Log aktivitas user

## 📦 Import Excel

### Format Template

**Registration:**
- Kolom: `phone_number`

**Transaction:**
- Kolom: `phone_number`, `transaction_date`, `total_deposit`, `total_profit`

**Data:**
- Kolom: `whatsapp`, `name`, `nik`, `client`

Template bisa didownload dari masing-masing halaman.

## 🛠️ Development

### Menambah Fitur Baru

1. **API Route**: Buat di `src/app/api/[nama]/route.ts`
2. **Page**: Buat di `src/app/[nama]/page.tsx`
3. **Component**: Buat di `src/components/[nama]/`
4. **Service**: Buat di `src/services/[nama].ts`
5. **Activity Log**: Tambahkan `createActivityLog()` di API yang melakukan insert

### Menambah Menu Sidebar

Edit `src/components/Layouts/sidebar/data/index.ts`:

```typescript
{
  title: "Menu Baru",
  url: "/menu-baru",
  icon: Icons.IconName,
  items: [],
}
```

## 🐛 Troubleshooting

### Error: DATABASE_URL not found
- Pastikan `.env.local` ada di root project
- Cek format DATABASE_URL: `postgresql://user:password@host:port/database`
- Restart development server setelah menambah .env.local

### Error: MissingSecret
- Tambahkan `AUTH_SECRET` di `.env.local`
- Generate secret: `openssl rand -base64 32`

### Error: Prisma migration
- Pastikan database sudah dibuat
- Cek connection string di DATABASE_URL
- Run `npx prisma db push` untuk sync schema tanpa migration

## 📄 License

Private project
# marketing-contact
