#!/usr/bin/env node

const https = require('https');
const { createWriteStream, chmodSync, existsSync, mkdirSync } = require('fs');
const path = require('path');
const { platform, arch } = require('os');

const VERSION = '0.1.0';
const REPO = 'user/spider';

const binaryName = platform() === 'win32' ? 'spider.exe' : 'spider';
const destDir = path.join(__dirname, 'bin');
const destPath = path.join(destDir, binaryName);

if (existsSync(destPath)) {
  console.log('  Spider binary already installed.');
  process.exit(0);
}

const archMap = {
  x64: 'amd64',
  arm64: 'arm64',
};

const platformMap = {
  win32: 'windows',
  darwin: 'darwin',
  linux: 'linux',
};

const goArch = archMap[arch()];
const goOS = platformMap[platform()];

if (!goArch || !goOS) {
  console.error(`  Unsupported platform: ${platform()}/${arch()}`);
  process.exit(1);
}

const url = `https://github.com/${REPO}/releases/download/v${VERSION}/spider_${VERSION}_${goOS}_${goArch}.tar.gz`;

console.log(`  Downloading Spider v${VERSION} for ${goOS}/${goArch}...`);

if (!existsSync(destDir)) {
  mkdirSync(destDir, { recursive: true });
}

const tempFile = destPath + '.download';
const file = createWriteStream(tempFile);

https.get(url, (response) => {
  if (response.statusCode !== 200) {
    console.error(`  Download failed (HTTP ${response.statusCode}).`);
    console.error(`  Please install manually: ${url}`);
    process.exit(1);
  }

  response.pipe(file);

  file.on('finish', () => {
    file.close(() => {
      // Extract from tar.gz or rename directly
      // For simplicity, assume the release has the raw binary
      const fs = require('fs');
      fs.renameSync(tempFile, destPath);
      if (platform() !== 'win32') {
        chmodSync(destPath, 0o755);
      }
      console.log('  Spider installed successfully.');
      console.log(`  Run: npx spider help`);
    });
  });
}).on('error', (err) => {
  console.error(`  Download failed: ${err.message}`);
  process.exit(1);
});
