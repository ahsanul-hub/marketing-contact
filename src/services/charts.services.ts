import { prisma } from "@/lib/prisma";
import { Prisma } from "@prisma/client";

export async function getTopClientsData() {
  // Get top 10 clients by total deposit and profit from transactions
  const topClients = await prisma.$queryRaw<
    {
      client_id: bigint | null;
      client_name: string | null;
      total_deposit: bigint;
      total_profit: bigint;
    }[]
  >`
    SELECT 
      c.id as client_id,
      c.name as client_name,
      COALESCE(SUM(t.total_deposit), 0)::bigint as total_deposit,
      COALESCE(SUM(t.total_profit), 0)::bigint as total_profit
    FROM client c
    LEFT JOIN data d ON d.id_client = c.id
    LEFT JOIN transaction t ON t.phone_number = d.whatsapp
    GROUP BY c.id, c.name
    ORDER BY total_profit DESC, total_deposit DESC
    LIMIT 10
  `;

  // Get organic (no client) data
  const organicData = await prisma.$queryRaw<{
    total_deposit: bigint;
    total_profit: bigint;
  }[]>`
    SELECT 
      COALESCE(SUM(t.total_deposit), 0)::bigint as total_deposit,
      COALESCE(SUM(t.total_profit), 0)::bigint as total_profit
    FROM transaction t
    WHERE t.phone_number IS NOT NULL
      AND NOT EXISTS (
        SELECT 1 FROM data d WHERE d.whatsapp = t.phone_number
      )
  `;

  const organic = organicData[0] || { total_deposit: 0n, total_profit: 0n };

  // Combine and sort all data
  const allClients = [
    ...(Number(organic.total_profit) > 0 || Number(organic.total_deposit) > 0
      ? [
          {
            name: "Organic",
            totalDeposit: Number(organic.total_deposit),
            totalProfit: Number(organic.total_profit),
          },
        ]
      : []),
    ...topClients.map((item) => ({
      name: item.client_name || `Client #${item.client_id?.toString()}`,
      totalDeposit: Number(item.total_deposit),
      totalProfit: Number(item.total_profit),
    })),
  ];

  // Sort by profit descending and take top 10
  return allClients
    .sort((a, b) => b.totalProfit - a.totalProfit)
    .slice(0, 10);
}

export async function getDevicesUsedData(
  timeFrame?: "monthly" | "yearly" | (string & {}),
) {
  // Get client data with registration count
  const clients = await prisma.client.findMany({
    orderBy: [{ name: "asc" }, { id: "asc" }],
    select: {
      id: true,
      name: true,
    },
  });

  // Count registrations per client (those with data.whatsapp matching registration.phone_number)
  const clientData = await Promise.all(
    clients.map(async (client) => {
      const count = await prisma.$queryRaw<{ count: bigint }[]>`
        SELECT COUNT(DISTINCT r.phone_number)::bigint as count
        FROM registration r
        INNER JOIN data d ON d.whatsapp = r.phone_number
        WHERE d.id_client = ${client.id}
      `;
      return {
        name: client.name || `Client #${client.id.toString()}`,
        amount: Number(count[0]?.count ?? 0n),
      };
    }),
  );

  // Count organic registrations (those without matching data)
  const organicCount = await prisma.$queryRaw<{ count: bigint }[]>`
    SELECT COUNT(DISTINCT r.phone_number)::bigint as count
    FROM registration r
    WHERE NOT EXISTS (
      SELECT 1 FROM data d WHERE d.whatsapp = r.phone_number
    )
  `;

  const organicAmount = Number(organicCount[0]?.count ?? 0n);

  // Combine all data
  const data = [
    ...(organicAmount > 0
      ? [
          {
            name: "Organic",
            amount: organicAmount,
          },
        ]
      : []),
    ...clientData.filter((item) => item.amount > 0),
  ];

  return data;
}

export async function getPaymentsOverviewData(
  timeFrame?: "monthly" | "yearly" | (string & {}),
) {
  // Fake delay
  await new Promise((resolve) => setTimeout(resolve, 1000));

  if (timeFrame === "yearly") {
    return {
      received: [
        { x: 2020, y: 450 },
        { x: 2021, y: 620 },
        { x: 2022, y: 780 },
        { x: 2023, y: 920 },
        { x: 2024, y: 1080 },
      ],
      due: [
        { x: 2020, y: 1480 },
        { x: 2021, y: 1720 },
        { x: 2022, y: 1950 },
        { x: 2023, y: 2300 },
        { x: 2024, y: 1200 },
      ],
    };
  }

  return {
    received: [
      { x: "Jan", y: 0 },
      { x: "Feb", y: 20 },
      { x: "Mar", y: 35 },
      { x: "Apr", y: 45 },
      { x: "May", y: 35 },
      { x: "Jun", y: 55 },
      { x: "Jul", y: 65 },
      { x: "Aug", y: 50 },
      { x: "Sep", y: 65 },
      { x: "Oct", y: 75 },
      { x: "Nov", y: 60 },
      { x: "Dec", y: 75 },
    ],
    due: [
      { x: "Jan", y: 15 },
      { x: "Feb", y: 9 },
      { x: "Mar", y: 17 },
      { x: "Apr", y: 32 },
      { x: "May", y: 25 },
      { x: "Jun", y: 68 },
      { x: "Jul", y: 80 },
      { x: "Aug", y: 68 },
      { x: "Sep", y: 84 },
      { x: "Oct", y: 94 },
      { x: "Nov", y: 74 },
      { x: "Dec", y: 62 },
    ],
  };
}

export async function getWeeksProfitData(timeFrame?: string) {
  // Fake delay
  await new Promise((resolve) => setTimeout(resolve, 1000));

  if (timeFrame === "last week") {
    return {
      sales: [
        { x: "Sat", y: 33 },
        { x: "Sun", y: 44 },
        { x: "Mon", y: 31 },
        { x: "Tue", y: 57 },
        { x: "Wed", y: 12 },
        { x: "Thu", y: 33 },
        { x: "Fri", y: 55 },
      ],
      revenue: [
        { x: "Sat", y: 10 },
        { x: "Sun", y: 20 },
        { x: "Mon", y: 17 },
        { x: "Tue", y: 7 },
        { x: "Wed", y: 10 },
        { x: "Thu", y: 23 },
        { x: "Fri", y: 13 },
      ],
    };
  }

  return {
    sales: [
      { x: "Sat", y: 44 },
      { x: "Sun", y: 55 },
      { x: "Mon", y: 41 },
      { x: "Tue", y: 67 },
      { x: "Wed", y: 22 },
      { x: "Thu", y: 43 },
      { x: "Fri", y: 65 },
    ],
    revenue: [
      { x: "Sat", y: 13 },
      { x: "Sun", y: 23 },
      { x: "Mon", y: 20 },
      { x: "Tue", y: 8 },
      { x: "Wed", y: 13 },
      { x: "Thu", y: 27 },
      { x: "Fri", y: 15 },
    ],
  };
}

export async function getCampaignVisitorsData() {
  // Fake delay
  await new Promise((resolve) => setTimeout(resolve, 1000));

  return {
    total_visitors: 784_000,
    performance: -1.5,
    chart: [
      { x: "S", y: 168 },
      { x: "S", y: 385 },
      { x: "M", y: 201 },
      { x: "T", y: 298 },
      { x: "W", y: 187 },
      { x: "T", y: 195 },
      { x: "F", y: 291 },
    ],
  };
}

export async function getVisitorsAnalyticsData() {
  // Fake delay
  await new Promise((resolve) => setTimeout(resolve, 1000));

  return [
    168, 385, 201, 298, 187, 195, 291, 110, 215, 390, 280, 112, 123, 212, 270,
    190, 310, 115, 90, 380, 112, 223, 292, 170, 290, 110, 115, 290, 380, 312,
  ].map((value, index) => ({ x: index + 1 + "", y: value }));
}

export async function getCostsPerInteractionData() {
  return {
    avg_cost: 560.93,
    growth: 2.5,
    chart: [
      {
        name: "Google Ads",
        data: [
          { x: "Sep", y: 15 },
          { x: "Oct", y: 12 },
          { x: "Nov", y: 61 },
          { x: "Dec", y: 118 },
          { x: "Jan", y: 78 },
          { x: "Feb", y: 125 },
          { x: "Mar", y: 165 },
          { x: "Apr", y: 61 },
          { x: "May", y: 183 },
          { x: "Jun", y: 238 },
          { x: "Jul", y: 237 },
          { x: "Aug", y: 235 },
        ],
      },
      {
        name: "Facebook Ads",
        data: [
          { x: "Sep", y: 75 },
          { x: "Oct", y: 77 },
          { x: "Nov", y: 151 },
          { x: "Dec", y: 72 },
          { x: "Jan", y: 7 },
          { x: "Feb", y: 58 },
          { x: "Mar", y: 60 },
          { x: "Apr", y: 185 },
          { x: "May", y: 239 },
          { x: "Jun", y: 135 },
          { x: "Jul", y: 119 },
          { x: "Aug", y: 124 },
        ],
      },
    ],
  };
}