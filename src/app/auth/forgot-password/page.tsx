import { Metadata } from "next";
import Link from "next/link";
import ForgotPasswordForm from "./_components/forgot-password-form";

export const metadata: Metadata = {
  title: "Forgot Password",
  description: "Reset your password",
};

export const dynamic = "force-dynamic";

export default function ForgotPasswordPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 dark:bg-gray-900 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
            Reset Password
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
            Masukkan email dan password baru Anda
          </p>
        </div>

        <ForgotPasswordForm />

        <div className="text-center">
          <Link
            href="/auth/sign-in"
            className="text-sm font-medium text-primary hover:text-opacity-80 dark:text-primary"
          >
            Kembali ke halaman login
          </Link>
        </div>
      </div>
    </div>
  );
}
