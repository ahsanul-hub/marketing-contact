/**
 * NextAuth Configuration
 * 
 * This file configures authentication using NextAuth.js with Credentials provider.
 * Users are authenticated against the User table in the database.
 * 
 * Flow:
 * 1. User submits username/password via sign-in form
 * 2. authorize() function checks credentials against database
 * 3. Password is verified using bcrypt (password stored as bytea)
 * 4. If valid, user data is returned and stored in JWT token
 * 5. Session is created with user id, username, and role
 * 
 * Security:
 * - Passwords are hashed with bcrypt (10 rounds) and stored as bytea
 * - JWT tokens are signed with AUTH_SECRET
 * - Only active users can login
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
       * Authorize function - validates user credentials
       * @param credentials - Username and password from login form
       * @returns User object if valid, null if invalid
       */
      authorize: async (credentials) => {
        try {
          const username = credentials?.username?.toString()?.trim() || "";
          const password = credentials?.password?.toString() || "";

          if (!username || !password) {
            console.log("[Auth] Missing username or password");
            return null;
          }

          // Cari user di database berdasarkan username
          const user = await prisma.user.findUnique({
            where: { username },
          });

          // Cek apakah user ada
          if (!user) {
            console.log(`[Auth] User not found: ${username}`);
            return null;
          }

          // Cek apakah user aktif
          if (!user.active) {
            console.log(`[Auth] User is inactive: ${username}`);
            return null;
          }

          // Convert password bytea (Buffer) ke string untuk bcrypt.compare
          // Password di database disimpan sebagai bytea (binary), perlu di-convert ke string
          const passwordHash = Buffer.from(user.password).toString("utf-8");

          // Verifikasi password menggunakan bcrypt
          const isValid = await bcrypt.compare(password, passwordHash);

          if (!isValid) {
            console.log(`[Auth] Invalid password for: ${username}`);
            return null;
          }

          console.log(`[Auth] Login successful: ${username}`);

          // Return user data yang akan disimpan di JWT token
          return {
            id: user.id.toString(),
            username: user.username,
            name: user.username, // Use username as name
            role: user.role, // "admin", "user", atau "client"
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
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  callbacks: {
    /**
     * JWT callback - dipanggil saat token dibuat/updated
     * Menyimpan user id dan role ke token
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

