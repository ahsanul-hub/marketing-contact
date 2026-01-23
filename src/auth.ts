/**
 * NextAuth Configuration
 * 
 * This file configures authentication using NextAuth.js with Credentials provider.
 * Admins are authenticated against the Admin table in the database.
 * 
 * Flow:
 * 1. Admin submits username/password via sign-in form
 * 2. authorize() function checks credentials against database
 * 3. Password is verified using bcrypt (password stored as bytea)
 * 4. If valid, admin data is returned and stored in JWT token
 * 5. Session is created with admin id, username, and role
 * 
 * Security:
 * - Passwords are hashed with bcrypt (10 rounds) and stored as bytea
 * - JWT tokens are signed with AUTH_SECRET
 * - Only active admins can login
 */

import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import { prisma } from "@/lib/prisma";
import bcrypt from "bcryptjs";

export const { handlers, auth, signIn, signOut } = NextAuth({
  // Secret untuk sign JWT tokens
  // IMPORTANT: Set AUTH_SECRET di .env.local untuk production!
  secret: process.env.AUTH_SECRET || process.env.NEXTAUTH_SECRET || "default-secret-change-in-production",
  
  providers: [
    Credentials({
      name: "Credentials",
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
      },
      /**
       * Authorize function - validates admin credentials
       * @param credentials - Username and password from login form
       * @returns Admin object if valid, null if invalid
       */
      authorize: async (credentials) => {
        try {
          const username = credentials?.username?.toString()?.trim() || "";
          const password = credentials?.password?.toString() || "";

          if (!username || !password) {
            console.log("[Auth] Missing username or password");
            return null;
          }

          // Cari admin di database berdasarkan username
          const admin = await prisma.admin.findUnique({
            where: { username },
          });

          // Cek apakah admin ada
          if (!admin) {
            console.log(`[Auth] Admin not found: ${username}`);
            return null;
          }

          // Cek apakah admin aktif
          if (!admin.isActive) {
            console.log(`[Auth] Admin is inactive: ${username}`);
            return null;
          }

          // Convert password bytea (Buffer) ke string untuk bcrypt.compare
          // Password di database disimpan sebagai bytea (binary), perlu di-convert ke string
          const passwordHash = Buffer.from(admin.password).toString("utf-8");

          // Verifikasi password menggunakan bcrypt
          const isValid = await bcrypt.compare(password, passwordHash);

          if (!isValid) {
            console.log(`[Auth] Invalid password for: ${username}`);
            return null;
          }

          console.log(`[Auth] Login successful: ${username}`);

          // Return admin data yang akan disimpan di JWT token
          return {
            id: admin.id.toString(),
            username: admin.username,
            name: admin.username, // Use username as name
            role: admin.role, // "admin" atau "user"
          };
        } catch (error) {
          console.error("[Auth] Error:", error);
          // Return null on error to prevent exposing error details
          return null;
        }
      },
    }),
  ],
  pages: {
    signIn: "/auth/sign-in", // Custom sign-in page
  },
  session: {
    strategy: "jwt", // Menggunakan JWT strategy (tidak perlu database session)
  },
  callbacks: {
    /**
     * JWT callback - dipanggil saat token dibuat/updated
     * Menyimpan admin id dan role ke token
     */
    async jwt({ token, user }) {
      if (user) {
        token.id = user.id;
        token.role = (user as any).role;
        token.username = (user as any).username;
      }
      return token;
    },
    /**
     * Session callback - dipanggil saat session diakses
     * Menambahkan id, username, dan role ke session object
     */
    async session({ session, token }) {
      if (session.user) {
        (session.user as any).id = token.id;
        (session.user as any).role = token.role;
        (session.user as any).username = token.username;
      }
      return session;
    },
  },
});

