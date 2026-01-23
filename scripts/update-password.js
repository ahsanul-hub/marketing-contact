/**
 * Script untuk update password user di database
 * 
 * Usage:
 * node scripts/update-password.js fahrilake45@gmail.com password123
 */

const { PrismaClient } = require('@prisma/client');
const bcrypt = require('bcryptjs');

const prisma = new PrismaClient();

async function updatePassword() {
  const email = process.argv[2];
  const password = process.argv[3] || 'password123';

  if (!email) {
    console.error('Usage: node scripts/update-password.js <email> [password]');
    process.exit(1);
  }

  try {
    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);

    console.log('\n=== Update Password ===');
    console.log('Email:', email);
    console.log('Password:', password);
    console.log('\nHashed Password:');
    console.log(hashedPassword);

    // Update password di database
    const user = await prisma.user.update({
      where: { email: email.toLowerCase() },
      data: {
        password: hashedPassword,
      },
      select: {
        id: true,
        email: true,
        name: true,
        role: true,
      },
    });

    console.log('\n✅ Password updated successfully!');
    console.log('User:', user);
    console.log('\n=== SQL Query (for reference) ===');
    console.log(`UPDATE "User" SET password = '${hashedPassword}', "updatedAt" = NOW() WHERE email = '${email.toLowerCase()}';`);
    console.log('\n=== Atau gunakan API ===');
    console.log(`curl -X POST http://localhost:3000/api/reset-password \\`);
    console.log(`  -H "Content-Type: application/json" \\`);
    console.log(`  -d '{"email":"${email}","password":"${password}"}'`);
    console.log('\n');
  } catch (error) {
    console.error('\n❌ Error:', error.message);
    if (error.code === 'P2025') {
      console.error('User not found. Make sure the email is correct.');
    }
    process.exit(1);
  } finally {
    await prisma.$disconnect();
  }
}

updatePassword();
