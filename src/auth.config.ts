import type { NextAuthConfig } from "next-auth";

export const authConfig = {
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
  providers: [], // Providers configured in auth.ts
} satisfies NextAuthConfig;
