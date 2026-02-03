import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatIDR } from "@/lib/currency";
import { cn } from "@/lib/utils";
import { getTopClientsData } from "@/services/charts.services";

type PropsType = {
  className?: string;
  filter?: {
    startDate?: Date;
    endDate?: Date;
    clientId?: number;
    isOrganic?: boolean;
  };
};

export async function TopClients({ className, filter }: PropsType) {
  const data = await getTopClientsData(filter);

  return (
    <div
      className={cn(
        "grid rounded-[10px] bg-white px-7.5 pb-4 pt-7.5 shadow-1 dark:bg-gray-dark dark:shadow-card",
        className,
      )}
    >
      <h2 className="mb-4 text-body-2xlg font-bold text-dark dark:text-white">
        Top 10 Clients
      </h2>

      <Table>
        <TableHeader>
          <TableRow className="border-none uppercase [&>th]:text-center">
            <TableHead className="min-w-[180px] !text-left">Client</TableHead>
            <TableHead className="!text-right">Total Deposit</TableHead>
            <TableHead className="!text-right">Total Profit</TableHead>
            <TableHead className="!text-right">Conversion Rate</TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          {data.length === 0 ? (
            <TableRow>
              <TableCell
                className="text-center text-neutral-500 dark:text-neutral-300"
                colSpan={4}
              >
                Belum ada data client.
              </TableCell>
            </TableRow>
          ) : (
            data.map((item, i) => (
              <TableRow
                className="text-center text-base font-medium text-dark dark:text-white"
                key={item.name + i}
              >
                <TableCell className="!text-left font-medium">
                  {item.name || "-"}
                </TableCell>

                <TableCell className="!text-right">
                  {formatIDR(item.totalDeposit)}
                </TableCell>

                <TableCell className="!text-right text-green-light-1">
                  {formatIDR(item.totalProfit)}
                </TableCell>

                <TableCell className="!text-right">
                  {item.conversionRate?.toFixed(2) ?? "0"}%
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
