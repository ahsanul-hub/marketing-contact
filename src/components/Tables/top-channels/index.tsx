import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatIDR } from "@/lib/currency";
import { compactFormat } from "@/lib/format-number";
import { cn } from "@/lib/utils";
import { getTopProfit } from "../fetch";

export async function TopChannels({ className }: { className?: string }) {
  const data = await getTopProfit();

  return (
    <div
      className={cn(
        "grid rounded-[10px] bg-white px-7.5 pb-4 pt-7.5 shadow-1 dark:bg-gray-dark dark:shadow-card",
        className,
      )}
    >
      <h2 className="mb-4 text-body-2xlg font-bold text-dark dark:text-white">
        Top Profit
      </h2>

      <Table>
        <TableHeader>
          <TableRow className="border-none uppercase [&>th]:text-center">
            <TableHead className="min-w-[180px] !text-left">Phone Number</TableHead>
            <TableHead className="!text-right">Total Deposit</TableHead>
            <TableHead className="!text-right">Total Profit</TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          {data.length === 0 ? (
            <TableRow>
              <TableCell
                className="text-center text-neutral-500 dark:text-neutral-300"
                colSpan={3}
              >
                Belum ada data transaksi.
              </TableCell>
            </TableRow>
          ) : (
            data.map((item, i) => (
              <TableRow
                className="text-center text-base font-medium text-dark dark:text-white"
                key={item.phoneNumber + i}
              >
                <TableCell className="!text-left font-medium">
                  {item.phoneNumber || "-"}
                </TableCell>

                <TableCell className="!text-right">
                  {formatIDR(item.totalDeposit)}
                </TableCell>

                <TableCell className="!text-right text-green-light-1">
                  {formatIDR(item.totalProfit)}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
