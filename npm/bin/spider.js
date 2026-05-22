#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

const binaryName = process.platform === 'win32' ? 'spider.exe' : 'spider';
const binaryPath = path.join(__dirname, '..', 'bin', binaryName);

const args = process.argv.slice(2);

const child = spawn(binaryPath, args, {
  stdio: 'inherit',
  env: { ...process.env },
});

child.on('error', (err) => {
  if (err.code === 'ENOENT') {
    console.error('');
    console.error('  Spider binary not found.');
    console.error('  Run `npm install` in the spider-agent-harness package to download it.');
    console.error('  Or install manually from https://github.com/Yleex/spider-agent-harness/releases');
    console.error('');
    process.exit(1);
  }
  console.error('Failed to start spider:', err.message);
  process.exit(1);
});

child.on('exit', (code) => {
  process.exit(code);
});
