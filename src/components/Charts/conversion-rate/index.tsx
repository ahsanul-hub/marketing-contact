import { cn } from "@/lib/utils";
import { getConversionRateData } from "@/services/analytics";

type Props = {
  className?: string;
  filter?: {
    startDate?: Date;
    endDate?: Date;
    clientId?: number;
    isOrganic?: boolean;
  };
};

export async function ConversionRateCard({ className, filter }: Props) {
  const data = await getConversionRateData(filter);

  return (
    <div
      className={cn(
        "rounded-[10px] border border-stroke bg-white px-7.5 py-6 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card",
        className,
      )}
    >
      <div className="flex items-start justify-between">
        <div>
          <h4 className="mb-2 text-body-sm font-bold text-dark dark:text-white">
            Conversion Rate
          </h4>
          <p className="text-sm text-neutral-600 dark:text-neutral-400">
            Registration to Transaction
          </p>
        </div>
      </div>

      <div className="mt-6 flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-neutral-600 dark:text-neutral-400">
            Registrations:
          </span>
          <span className="text-lg font-bold text-dark dark:text-white">
            {data.registrations}
          </span>
        </div>

        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-neutral-600 dark:text-neutral-400">
            Transactions:
          </span>
          <span className="text-lg font-bold text-dark dark:text-white">
            {data.transactions}
          </span>
        </div>

        <div className="border-t border-stroke dark:border-dark-3 pt-4">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-neutral-600 dark:text-neutral-400">
              Conversion Rate:
            </span>
            <span className="text-2xl font-bold text-primary">
              {data.conversionRate}%
            </span>
          </div>
        </div>

        {/* Progress bar */}
        <div className="mt-4 h-2 w-full overflow-hidden rounded-full bg-neutral-200 dark:bg-dark-3">
          <div
            className="h-full bg-primary transition-all duration-300"
            style={{ width: `${Math.min(data.conversionRate, 100)}%` }}
          />
        </div>
      </div>
    </div>
  );
}
