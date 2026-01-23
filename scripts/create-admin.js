/**
 * Script untuk membuat admin user
 * 
 * Usage:
 * node scripts/create-admin.js
 * 
 * Atau dengan email custom:
 * node scripts/create-admin.js admin@example.com
 */

const bcrypt = require('bcryptjs');

async function createAdmin() {
  const email = process.argv[2] || 'admin@example.com';
  const password = 'password123';
  const name = 'Administrator';

  // Hash password
  const hashedPassword = await bcrypt.hash(password, 10);

  console.log('\n=== Create Admin User ===');
  console.log('Email:', email);
  console.log('Password:', password);
  console.log('Name:', name);
  console.log('\nHashed Password:');
  console.log(hashedPassword);
  console.log('\n=== SQL Query ===');
  console.log(`INSERT INTO "User" (email, password, name, role, "isActive", "createdAt", "updatedAt")`);
  console.log(`VALUES ('${email}', '${hashedPassword}', '${name}', 'admin', true, NOW(), NOW());`);
  console.log('\n=== Atau gunakan API ===');
  console.log(`curl -X POST http://localhost:3000/api/setup \\`);
  console.log(`  -H "Content-Type: application/json" \\`);
  console.log(`  -d '{"email":"${email}","password":"${password}","name":"${name}"}'`);
  console.log('\n');
}

createAdmin().catch(console.error);
