import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { prisma } from "@/lib/prisma";
import { auth } from "@/auth";
import { redirect } from "next/navigation";
import { CreateUserForm } from "./_components/create-user-form";
import { ResetPasswordButton } from "./_components/reset-password-button";
import dayjs from "dayjs";

export const metadata = {
  title: "Manage Users",
};

// Disable performance measurement untuk menghindari error
export const dynamic = "force-dynamic";
export const revalidate = 0;

export default async function UsersPage() {
  // Auth check - pastikan admin sudah login
  // Middleware sudah handle redirect ke /auth/sign-in jika belum login
  const session = await auth();

  if (!session || !session.user) {
    redirect("/auth/sign-in");
  }

  // Authorization check - pastikan role adalah "admin"
  // Middleware sudah handle redirect ke / jika bukan admin, tapi double check untuk safety
  const userRole = (session.user as any)?.role;
  if (userRole !== "admin") {
    redirect("/");
  }

  // Fetch admins dari database
  const admins = await prisma.admin.findMany({
    select: {
      id: true,
      username: true,
      role: true,
      isActive: true,
      createdAt: true,
    },
    orderBy: { createdAt: "desc" },
  });

  return (
    <div className="space-y-6">
      <Breadcrumb pageName="Manage Users" />

      <div className="rounded-[10px] border border-stroke bg-white p-4 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card sm:p-7.5">
        <div className="mb-4">
          <h3 className="text-lg font-semibold text-dark dark:text-white">
            Tambah Admin/User
          </h3>
          <p className="text-sm text-neutral-500 dark:text-neutral-300">
            Tambahkan user baru untuk akses sistem
          </p>
        </div>

        <CreateUserForm />

        <div className="mt-6">
          <h3 className="mb-4 text-lg font-semibold text-dark dark:text-white">
            Daftar Admin
          </h3>

          <Table>
            <TableHeader>
              <TableRow className="border-none bg-[#F7F9FC] dark:bg-dark-2 [&>th]:py-4 [&>th]:text-base [&>th]:text-dark [&>th]:dark:text-white">
                <TableHead className="min-w-[200px]">Username</TableHead>
                <TableHead className="min-w-[100px]">Role</TableHead>
                <TableHead className="min-w-[100px]">Status</TableHead>
                <TableHead className="min-w-[180px]">Created At</TableHead>
                <TableHead className="min-w-[150px]">Actions</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              {admins.length === 0 ? (
                <TableRow>
                  <TableCell
                    className="text-center text-neutral-500 dark:text-neutral-300"
                    colSpan={5}
                  >
                    Belum ada admin.
                  </TableCell>
                </TableRow>
              ) : (
                admins.map((admin: typeof admins[0]) => (
                  <TableRow
                    key={admin.id}
                    className="border-[#eee] dark:border-dark-3"
                  >
                    <TableCell className="font-medium text-dark dark:text-white">
                      {admin.username}
                    </TableCell>
                    <TableCell>
                      <span
                        className={`inline-block rounded-full px-3 py-1 text-xs font-medium ${
                          admin.role === "admin"
                            ? "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200"
                            : "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                        }`}
                      >
                        {admin.role}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span
                        className={`inline-block rounded-full px-3 py-1 text-xs font-medium ${
                          admin.isActive
                            ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                            : "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
                        }`}
                      >
                        {admin.isActive ? "Active" : "Inactive"}
                      </span>
                    </TableCell>
                    <TableCell className="text-neutral-600 dark:text-neutral-300">
                      {dayjs(admin.createdAt).format("YYYY-MM-DD HH:mm")}
                    </TableCell>
                    <TableCell>
                      <ResetPasswordButton
                        userId={admin.id}
                        userEmail={admin.username}
                      />
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
