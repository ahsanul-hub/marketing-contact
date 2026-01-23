/**
 * NextAuth API Route Handler
 * 
 * Route ini menangani semua NextAuth API requests:
 * - GET/POST /api/auth/signin
 * - GET/POST /api/auth/signout
 * - GET /api/auth/session
 * - GET /api/auth/csrf
 * 
 * Handlers di-export dari auth.ts configuration
 * 
 * IMPORTANT: Pastikan DATABASE_URL dan AUTH_SECRET sudah di-set di .env.local
 */

import { handlers } from "@/auth";

// Export handlers langsung (NextAuth v5 format)
export const { GET, POST } = handlers;

