/**
 * Middleware - Route Protection & Authorization
 * 
 * Middleware ini berjalan sebelum setiap request dan melakukan:
 * 1. Authentication check - semua route kecuali public routes memerlukan login
 * 2. Authorization check - route /admin/* hanya bisa diakses role "admin"
 * 3. API protection - API tertentu hanya bisa diakses admin
 * 
 * Authorization Rules:
 * - Role "user": Bisa akses semua halaman kecuali /admin/*
 * - Role "admin": Bisa akses semua halaman termasuk /admin/*
 * - Belum login: Redirect ke /auth/sign-in
 * 
 * Flow:
 * - Public routes (login, forgot password, NextAuth API) -> langsung allow
 * - API routes -> check authentication/authorization sesuai kebutuhan
 * - Page routes -> require authentication, admin routes require admin role
 * - Jika tidak authenticated -> redirect ke /auth/sign-in
 * - Jika tidak authorized -> redirect ke home atau return 403
 */

import { NextResponse } from "next/server";
import { auth } from "@/auth";

// Routes yang bisa diakses tanpa login
// Hanya routes yang benar-benar public (login page, forgot password, dan NextAuth API)
const publicRoutes = [
  "/auth/sign-in",        // Halaman login
  "/auth/sign-up",        // Halaman registrasi
  "/auth/forgot-password", // Halaman forgot password
  "/api/auth",            // NextAuth API endpoints (signin, signout, session, csrf)
  "/api/reset-password",  // API untuk reset password (public untuk forgot password)
  "/api/register",        // API untuk registrasi user baru
];

export default auth((req) => {
  const { pathname } = req.nextUrl;
  const session = req.auth;

  console.log(`[Middleware] Path: ${pathname}, Session: ${!!session}`);

  // 1. PUBLIC ROUTES - Allow tanpa authentication
  if (publicRoutes.some(route => pathname.startsWith(route))) {
    return NextResponse.next();
  }

  // 2. API ROUTES - Handle secara khusus
  if (pathname.startsWith("/api")) {
    // API yang bisa diakses public (untuk forgot password dan register)
    if (pathname.startsWith("/api/reset-password") || pathname.startsWith("/api/register")) {
      return NextResponse.next();
    }

    // API yang memerlukan admin role (admin-only APIs)
    const adminOnlyAPIs = [
      "/api/admin",
      "/api/users",
      "/api/setup",
      "/api/create-user",
    ];

    if (adminOnlyAPIs.some(route => pathname.startsWith(route))) {
      // Cek authentication
      if (!session || !session.user) {
        return NextResponse.json(
          { error: "Unauthorized" },
          { status: 401 },
        );
      }
      // Cek authorization - hanya admin yang bisa akses
      if ((session.user as any)?.role !== "admin") {
        return NextResponse.json(
          { error: "Forbidden. Admin access required." },
          { status: 403 },
        );
      }
    } else {
      // API lainnya memerlukan authentication (role user atau admin bisa akses)
      // Contoh: /api/clients, /api/registration/bulk, /api/transaction/bulk, /api/data/bulk, /api/activity-logs
      if (!session || !session.user) {
        return NextResponse.json(
          { error: "Unauthorized" },
          { status: 401 },
        );
      }
    }
    return NextResponse.next();
  }

  // 3. PAGE ROUTES - Require authentication
  // Cek apakah sudah login
  if (!session || !session.user) {
    // Redirect ke login dengan callback URL
    const signInUrl = new URL("/auth/sign-in", req.url);
    signInUrl.searchParams.set("callbackUrl", pathname);
    // Set no-cache headers untuk memastikan tidak ada cache
    const response = NextResponse.redirect(signInUrl);
    response.headers.set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate");
    response.headers.set("Pragma", "no-cache");
    response.headers.set("Expires", "0");
    return response;
  }

  // 4. ADMIN ROUTES - Hanya role "admin" yang bisa akses
  // Role "user" tidak bisa akses /admin/*
  if (pathname.startsWith("/admin")) {
    const userRole = (session.user as any)?.role;
    if (userRole !== "admin") {
      // Redirect ke home jika bukan admin
      return NextResponse.redirect(new URL("/", req.url));
    }
  }

  // 5. Semua route lainnya bisa diakses oleh role "user" dan "admin"
  // Contoh: /, /registration, /transaction, /data, /client
  return NextResponse.next();
});

export const config = {
  // Match semua routes kecuali static files dan Next.js internal files
  // Matcher ini akan memproses semua request termasuk API routes
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - images (public images folder)
     * - *.svg, *.png, *.jpg, *.jpeg, *.gif, *.ico, *.webp (static assets)
     */
    "/((?!_next/static|_next/image|favicon.ico|images|.*\\.(?:svg|png|jpg|jpeg|gif|ico|webp)$).*)",
  ],
};

