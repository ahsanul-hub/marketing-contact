import * as XLSX from "xlsx";

export function downloadExcelTemplate(
  headers: string[],
  sampleData: any[][],
  filename: string,
) {
  // Create workbook
  const workbook = XLSX.utils.book_new();

  // Create worksheet with headers and sample data
  const worksheetData = [headers, ...sampleData];
  const worksheet = XLSX.utils.aoa_to_sheet(worksheetData);

  // Set column widths
  const colWidths = headers.map(() => ({ wch: 20 }));
  worksheet["!cols"] = colWidths;

  // Add worksheet to workbook
  XLSX.utils.book_append_sheet(workbook, worksheet, "Template");

  // Generate Excel file and download
  XLSX.writeFile(workbook, filename);
}

export function generateExcelBuffer(
  headers: string[],
  data: any[][],
  sheetName: string = "Data",
) {
  // Create workbook
  const workbook = XLSX.utils.book_new();

  // Create worksheet with headers and data
  const worksheetData = [headers, ...data];
  const worksheet = XLSX.utils.aoa_to_sheet(worksheetData);

  // Set column widths
  const colWidths = headers.map(() => ({ wch: 20 }));
  worksheet["!cols"] = colWidths;

  // Add worksheet to workbook
  XLSX.utils.book_append_sheet(workbook, worksheet, sheetName);

  // Generate buffer
  return XLSX.write(workbook, { type: "buffer", bookType: "xlsx" });
}

export function downloadRegistrationTemplate() {
  const headers = ["phone_number", "client", "registered_at"];
  const sampleData = [
    ["081234567890", "Client A", "05-02-2026"],
    ["081234567891", "Client B", "05-02-2026"],
    ["081234567892", "", "05-02-2026"],
  ];

  downloadExcelTemplate(headers, sampleData, "template-registration.xlsx");
}

export function downloadTransactionTemplate() {
  const headers = [
    "phone_number",
    "transaction_date",
    "total_deposit",
    "total_profit",
    "client",
  ];
  const sampleData = [
    ["081234567890", "2024-01-15 10:30:00", "100000", "5000", "Client A"],
    ["081234567891", "2024-01-16 14:20:00", "200000", "10000", "Client B"],
    ["081234567892", "2024-01-17 09:15:00", "150000", "7500", ""],
  ];

  downloadExcelTemplate(headers, sampleData, "template-transaction.xlsx");
}

export function downloadDataTemplate() {
  const headers = ["whatsapp", "name", "nik", "owner_name"];
  const sampleData = [
    ["081234567890", "John Doe", "3201010101010001", "Pak A"],
    ["081234567891", "Jane Smith", "3201010101010002", "Pak B"],
    ["081234567892", "Bob Johnson", "3201010101010003", "Pak C"],
  ];

  downloadExcelTemplate(headers, sampleData, "template-data.xlsx");
}