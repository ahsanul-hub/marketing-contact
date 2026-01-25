/**
 * Script untuk membuat user pertama kali di production
 * 
 * Usage:
 * DATABASE_URL="postgresql://user:pass@host:5432/db" node scripts/create-user-production.js
 * 
 * Atau dengan parameter:
 * DATABASE_URL="..." node scripts/create-user-production.js admin password123 admin
 */

const { PrismaClient } = require('@prisma/client');
const bcrypt = require('bcryptjs');

const prisma = new PrismaClient();

async function createUser() {
  const username = process.argv[2] || 'admin';
  const password = process.argv[3] || 'admin123';
  const role = process.argv[4] || 'admin';

  if (!process.env.DATABASE_URL) {
    console.error('\n❌ Error: DATABASE_URL environment variable is not set');
    console.log('\nUsage:');
    console.log('  DATABASE_URL="postgresql://user:pass@host:5432/db" node scripts/create-user-production.js');
    console.log('\nOr with parameters:');
    console.log('  DATABASE_URL="..." node scripts/create-user-production.js <username> <password> <role>');
    process.exit(1);
  }

  try {
    // Check if user already exists
    const existing = await prisma.user.findUnique({
      where: { username },
    });

    if (existing) {
      console.log(`\n⚠️  User "${username}" already exists!`);
      process.exit(1);
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    const passwordBuffer = Buffer.from(hashedPassword, 'utf-8');

    // Create user
    const user = await prisma.user.create({
      data: {
        username,
        password: passwordBuffer,
        role: role === 'admin' ? 'admin' : role === 'client' ? 'client' : 'user',
        active: true,
        updated_at: new Date(),
      },
      select: {
        id: true,
        username: true,
        role: true,
        active: true,
        created_at: true,
      },
    });

    console.log('\n✅ User created successfully!');
    console.log('\n=== User Details ===');
    console.log('ID:', user.id);
    console.log('Username:', user.username);
    console.log('Role:', user.role);
    console.log('Active:', user.active);
    console.log('Created:', user.created_at);
    console.log('\n=== Login Credentials ===');
    console.log('Username:', username);
    console.log('Password:', password);
    console.log('\n⚠️  Please save these credentials securely!');
    console.log('⚠️  Change the password after first login!\n');
  } catch (error) {
    console.error('\n❌ Error creating user:', error.message);
    if (error.code === 'P2002') {
      console.error('   User with this username already exists!');
    }
    process.exit(1);
  } finally {
    await prisma.$disconnect();
  }
}

createUser();

