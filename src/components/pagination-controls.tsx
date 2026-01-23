import Link from "next/link";

type PaginationControlsProps = {
  page: number;
  limit: number;
  total: number;
  basePath: string;
  params?: Record<string, string | undefined>;
};

export function PaginationControls({
  page,
  limit,
  total,
  basePath,
  params,
}: PaginationControlsProps) {
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const currentPage = Math.min(page, totalPages);
  const hasPrevious = currentPage > 1;
  const hasNext = currentPage < totalPages;

  const start = total === 0 ? 0 : (currentPage - 1) * limit + 1;
  const end = total === 0 ? 0 : Math.min(currentPage * limit, total);

  const buildHref = (targetPage: number) => {
    const search = new URLSearchParams();
    search.set("page", targetPage.toString());
    // Always keep limit param stable across paging.
    search.set("limit", limit.toString());

    Object.entries(params ?? {}).forEach(([key, value]) => {
      if (value !== undefined && value !== "") search.set(key, value);
    });

    return `${basePath}?${search.toString()}`;
  };

  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm text-neutral-600 dark:text-neutral-300">
      <div>
        {total === 0
          ? "No records found"
          : `Showing ${start}-${end} of ${total}`}
      </div>

      <div className="flex items-center gap-2">
        <PaginationLink disabled={!hasPrevious} href={buildHref(currentPage - 1)}>
          Previous
        </PaginationLink>

        <span className="text-neutral-500 dark:text-neutral-400">
          Page {currentPage} of {totalPages}
        </span>

        <PaginationLink disabled={!hasNext} href={buildHref(currentPage + 1)}>
          Next
        </PaginationLink>
      </div>
    </div>
  );
}

function PaginationLink({
  disabled,
  href,
  children,
}: {
  disabled: boolean;
  href: string;
  children: React.ReactNode;
}) {
  const baseClasses =
    "rounded-md border border-neutral-200 px-3 py-1.5 text-sm font-medium transition hover:border-primary hover:text-primary dark:border-dark-3";

  if (disabled) {
    return (
      <span className={`${baseClasses} cursor-not-allowed opacity-50`}>
        {children}
      </span>
    );
  }

  return (
    <Link className={baseClasses} href={href}>
      {children}
    </Link>
  );
}

