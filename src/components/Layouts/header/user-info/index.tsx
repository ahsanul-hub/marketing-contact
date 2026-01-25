"use client";

import { ChevronUpIcon } from "@/assets/icons";
import {
  Dropdown,
  DropdownContent,
  DropdownTrigger,
} from "@/components/ui/dropdown";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { LogOutIcon, LogInIcon } from "./icons";
import { signOut } from "next-auth/react";
import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import Link from "next/link";

export function UserInfo() {
  const [isOpen, setIsOpen] = useState(false);
  const { data: session } = useSession();
  const router = useRouter();

  const handleLogout = async () => {
    setIsOpen(false);
    try {
      // Sign out dengan redirect: true untuk memastikan session dihapus dan redirect terjadi
      await signOut({ 
        redirect: true,
        callbackUrl: "/auth/sign-in" 
      });
    } catch (error) {
      console.error("Logout error:", error);
      // Jika ada error, tetap redirect ke sign-in
      router.push("/auth/sign-in");
      router.refresh();
    }
  };

  // Jika belum login, tampilkan tombol login
  if (!session) {
    return (
      <Link
        href="/auth/sign-in"
        className="flex items-center gap-2 rounded-lg border border-primary bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-dark focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 dark:focus:ring-offset-gray-dark"
      >
        <LogInIcon />
        <span>Login</span>
      </Link>
    );
  }

  const userName = (session?.user as any)?.name || session?.user?.email || "User";
  const userEmail = session?.user?.email || "";
  const userRole = (session?.user as any)?.role || "";

  // Get initials for avatar badge
  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const initials = getInitials(userName);

  return (
    <Dropdown isOpen={isOpen} setIsOpen={setIsOpen}>
      <DropdownTrigger className="rounded align-middle outline-none ring-primary ring-offset-2 focus-visible:ring-1 dark:ring-offset-gray-dark">
        <span className="sr-only">My Account</span>

        <figure className="flex items-center gap-3">
          {/* Avatar Siluet */}
          <div className="relative">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-gray-200 dark:bg-gray-700 border-2 border-gray-300 dark:border-gray-600">
              <svg
                className="h-8 w-8 text-gray-400 dark:text-gray-500"
                fill="currentColor"
                viewBox="0 0 20 20"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fillRule="evenodd"
                  d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            {/* Badge dengan initials */}
            <div className="absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-primary text-xs font-semibold text-white shadow-md">
              {initials}
            </div>
          </div>
          <figcaption className="flex items-center gap-1 font-medium text-dark dark:text-dark-6 max-[1024px]:sr-only">
            <span>{userName}</span>

            <ChevronUpIcon
              aria-hidden
              className={cn(
                "rotate-180 transition-transform",
                isOpen && "rotate-0",
              )}
              strokeWidth={1.5}
            />
          </figcaption>
        </figure>
      </DropdownTrigger>

      <DropdownContent
        className="border border-stroke bg-white shadow-md dark:border-dark-3 dark:bg-gray-dark min-[230px]:min-w-[17.5rem]"
        align="end"
      >
        <h2 className="sr-only">User information</h2>

        <figure className="flex items-center gap-2.5 px-5 py-3.5">
          {/* Avatar Siluet */}
          <div className="relative">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-gray-200 dark:bg-gray-700 border-2 border-gray-300 dark:border-gray-600">
              <svg
                className="h-8 w-8 text-gray-400 dark:text-gray-500"
                fill="currentColor"
                viewBox="0 0 20 20"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fillRule="evenodd"
                  d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            {/* Badge dengan initials */}
            <div className="absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-primary text-xs font-semibold text-white shadow-md">
              {initials}
            </div>
          </div>

          <figcaption className="space-y-1 text-base font-medium">
            <div className="mb-2 leading-none text-dark dark:text-white">
              {userName}
            </div>

            {userRole && (
              <div className="leading-none text-gray-6">
                Role: {userRole}
              </div>
            )}
          </figcaption>
        </figure>

        <hr className="border-[#E8E8E8] dark:border-dark-3" />

        <div className="p-2 text-base text-[#4B5563] dark:text-dark-6">
          <button
            className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-[9px] hover:bg-gray-2 hover:text-dark dark:hover:bg-dark-3 dark:hover:text-white"
            onClick={handleLogout}
          >
            <LogOutIcon />

            <span className="text-base font-medium">Log out</span>
          </button>
        </div>
      </DropdownContent>
    </Dropdown>
  );
}
