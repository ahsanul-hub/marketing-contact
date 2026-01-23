# Setup Guide - Marketing Contact Analysis

## ⚠️ Troubleshooting Errors

### Error: DATABASE_URL not found

Jika Anda mendapat error:
```
error: Environment variable not found: DATABASE_URL
```

### Solusi 1: Pastikan file .env.local ada di root project

File `.env.local` harus ada di **root project** (sama level dengan `package.json`), bukan di folder lain.

```
marketing-contact-analysis/
├── .env.local          ← HARUS DI SINI
├── package.json
├── prisma/
└── src/
```

### Solusi 2: Buat file .env (tanpa .local)

Prisma CLI kadang tidak membaca `.env.local`. Buat juga file `.env`:

```bash
# Copy dari .env.local
cp .env.local .env
```

Atau buat manual:
```bash
DATABASE_URL="postgresql://postgres:admin@localhost:5432/phone_data"
AUTH_SECRET="your-secret-key-here"
```

### Solusi 3: Restart Development Server

Setelah membuat/mengubah `.env.local` atau `.env`:
1. Stop development server (Ctrl+C)
2. Start lagi: `npm run dev`

### Solusi 4: Cek Format DATABASE_URL

Format yang benar:
```
postgresql://USER:PASSWORD@HOST:PORT/DATABASE
```

Contoh:
```
postgresql://postgres:admin@localhost:5432/phone_data
```

### Solusi 5: Test Connection

Test apakah database connection berfungsi:
```bash
npx prisma db pull
```

Jika berhasil, akan muncul schema dari database.

## 📋 Checklist Setup

- [ ] File `.env.local` ada di root project
- [ ] DATABASE_URL format benar
- [ ] Database sudah dibuat (`phone_data`)
- [ ] PostgreSQL running
- [ ] Run `npx prisma generate`
- [ ] Run `npx prisma db push` atau `npx prisma migrate dev`
- [ ] Restart development server
- [ ] Setup admin pertama kali via `/api/setup`

### Error: Failed to execute 'json' on 'Response': Unexpected end of JSON input

Error ini biasanya terjadi di halaman sign-in. Kemungkinan penyebab:

1. **Database tidak terhubung**
   - Pastikan PostgreSQL running
   - Test connection: `npx prisma db pull`
   - Cek DATABASE_URL format benar

2. **Tabel User belum ada**
   - Run migration: `npx prisma migrate dev`
   - Atau: `npx prisma db push`

3. **Belum ada admin user**
   - Setup admin via `/api/setup`
   - Atau insert manual ke database

4. **AUTH_SECRET tidak di-set**
   - Tambahkan `AUTH_SECRET` di `.env.local`
   - Generate: `openssl rand -base64 32`

### Error: Performance measurement negative timestamp

Error ini biasanya terjadi karena:
- Async operations yang terlalu cepat
- Sudah diperbaiki dengan `dynamic = "force-dynamic"`

Jika masih terjadi, restart development server.

## 🔍 Verifikasi

Setelah setup, cek:
1. Database connection: `npx prisma db pull`
2. Admin bisa login di `/auth/sign-in`
3. Dashboard bisa diakses setelah login
4. Activity logs muncul di notification bell
