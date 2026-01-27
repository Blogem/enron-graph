# Graph Explorer - Build and Deployment Guide

This document provides comprehensive instructions for building and deploying the Enron Graph Explorer desktop application.

## Project Structure

The Graph Explorer follows the standard Wails project structure:

```
enron-graph-2/
├── cmd/
│   └── explorer/
│       ├── main.go              # Application entry point
│       ├── app.go               # Go backend application logic
│       ├── wails.json           # Wails configuration
│       └── frontend/            # React + TypeScript frontend
│           ├── src/
│           ├── dist/            # Build output (generated)
│           ├── package.json
│           └── vite.config.ts
└── internal/
    └── explorer/                # Shared backend services
        ├── schema_service.go
        └── graph_service.go
```

## Prerequisites

### Required Tools

- **Go** 1.21 or later
- **Node.js** 16+ and npm
- **Wails CLI** v2.11.0 or later: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Platform-Specific Requirements

**macOS:**
- Xcode Command Line Tools: `xcode-select --install`
- Minimum OS: macOS 10.13 High Sierra

**Linux:**
- gtk3, webkit2gtk development libraries
- Ubuntu/Debian: `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`
- Fedora: `sudo dnf install gtk3-devel webkit2gtk3-devel`

**Windows:**
- WebView2 runtime (automatically downloaded by Wails)
- MinGW-w64 or MSVC for C++ compilation

### Runtime Dependencies

The application requires access to:
- PostgreSQL database (configure via `DATABASE_URL` environment variable or config file)
- Ollama (optional, for LLM-powered entity extraction during data loading)

## Development Workflow

### Running in Development Mode

Development mode provides hot reload and debugging tools:

```bash
# Navigate to the explorer directory
cd cmd/explorer

# Start development server (hot reload enabled)
wails dev
```

**What happens:**
- Vite dev server starts on a random port (auto-detected)
- Go backend starts with live reload
- Browser window opens with Developer Tools enabled
- Changes to frontend or backend trigger automatic reload

**Environment Configuration:**

Create `.env` file in project root:
```bash
DATABASE_URL=postgres://user:password@localhost:5432/enron_graph?sslmode=disable
OLLAMA_HOST=http://localhost:11434  # Optional
```

Or set environment variables:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/enron_graph?sslmode=disable"
wails dev
```

## Production Builds

### Building for Current Platform

```bash
# Navigate to explorer directory
cd cmd/explorer

# Clean build (recommended)
wails build -clean

# Output location:
# macOS:   build/bin/explorer.app
# Linux:   build/bin/explorer
# Windows: build/bin/explorer.exe
```

**Build Process:**

1. **Generate Go bindings** - Creates TypeScript/JavaScript bindings from Go structs
2. **Install frontend dependencies** - Runs `npm install` in frontend directory
3. **Build frontend** - Runs `npm run build` (Vite production build)
4. **Embed frontend assets** - Go embed directive packages `frontend/dist` into binary
5. **Compile Go application** - Creates optimized binary with embedded assets
6. **Package application** - Creates platform-specific bundle (.app for macOS)

**Build Flags:**

```bash
# Clean build (removes previous build artifacts)
wails build -clean

# Skip frontend build (use existing dist/)
wails build -skipfrontend

# Enable debug mode (includes DevTools in production)
wails build -debug

# Compress binary with UPX
wails build -upx

# Build for specific platform/arch
wails build -platform darwin/arm64
```

### Cross-Platform Builds

**Build for multiple platforms:**

```bash
# macOS → Windows
wails build -platform windows/amd64

# macOS → Linux
wails build -platform linux/amd64

# Build for all platforms
wails build -platform darwin/universal,linux/amd64,windows/amd64
```

**Notes:**
- Cross-compilation has limitations (especially macOS builds from other platforms)
- For best results, build on the target platform
- Windows builds from macOS/Linux require MinGW-w64

## Distribution

### macOS

**App Bundle Structure:**
```
explorer.app/
├── Contents/
│   ├── Info.plist           # App metadata
│   ├── MacOS/
│   │   └── explorer         # Binary with embedded assets
│   └── Resources/
│       └── iconfile.icns    # App icon
```

**Code Signing (for distribution):**

```bash
# Sign the app
codesign --deep --force --verify --verbose \
  --sign "Developer ID Application: Your Name" \
  build/bin/explorer.app

# Create notarized DMG (required for macOS 10.15+)
# 1. Create DMG
hdiutil create -volname "Enron Graph Explorer" \
  -srcfolder build/bin/explorer.app \
  -ov -format UDZO explorer.dmg

# 2. Submit for notarization
xcrun notarytool submit explorer.dmg \
  --apple-id your@email.com \
  --password "app-specific-password" \
  --team-id TEAMID

# 3. Staple notarization ticket
xcrun stapler staple explorer.dmg
```

**Installation:**
- Drag `explorer.app` to `/Applications`
- Or distribute as `.dmg` for user-friendly installation

### Linux

**Binary Distribution:**

```bash
# Build static binary (if possible)
CGO_ENABLED=1 wails build -clean

# Create tarball
tar -czf explorer-linux-amd64.tar.gz -C build/bin explorer

# Or create .deb package (requires additional tools)
```

**System Integration:**

Create desktop entry at `/usr/share/applications/explorer.desktop`:
```ini
[Desktop Entry]
Name=Enron Graph Explorer
Comment=Interactive knowledge graph visualization
Exec=/usr/local/bin/explorer
Icon=/usr/local/share/icons/explorer.png
Terminal=false
Type=Application
Categories=Development;Database;
```

### Windows

**Executable Distribution:**

```bash
# Build produces explorer.exe
wails build -clean -platform windows/amd64

# Bundle with dependencies (optional)
# Include config.yaml, README, etc.
```

**Installer Creation:**

Use tools like:
- **Inno Setup** - Free, script-based installer
- **WiX Toolset** - MSI installer creation
- **NSIS** - Nullsoft installer system

## Configuration Management

### Application Configuration

The app reads configuration from (in order of precedence):

1. **Environment variables** (highest priority)
   - `DATABASE_URL` - PostgreSQL connection string
   - `OLLAMA_HOST` - Ollama API endpoint (optional)

2. **Config file** - `config.yaml` in working directory or app directory:
   ```yaml
   database_url: postgres://user:pass@localhost:5432/enron_graph?sslmode=disable
   ollama_host: http://localhost:11434
   ```

3. **Default values** (lowest priority)

### Deployment Configuration

**Production Database:**

```bash
# Set production database URL
export DATABASE_URL="postgres://user:pass@prod-host:5432/enron_graph?sslmode=disable"

# Run app
./build/bin/explorer.app/Contents/MacOS/explorer  # macOS
./build/bin/explorer                               # Linux
```

**Read-Only Deployment:**

For read-only graph exploration (no schema promotion):
- App requires only SELECT permissions on database
- No migrations or schema modifications needed
- Safe for shared/demo environments

## Troubleshooting

### Build Issues

**Problem:** `no 'index.html' could be found in your Assets fs.FS`

**Solution:** Frontend wasn't built before Go compilation
```bash
cd cmd/explorer/frontend
npm install
npm run build
cd ..
wails build -clean
```

---

**Problem:** `frontend/dist: no such file or directory`

**Solution:** Run frontend build manually
```bash
cd cmd/explorer/frontend
npm run build
```

---

**Problem:** `pattern all:frontend/dist: cannot embed directory`

**Solution:** Ensure `frontend/` directory exists at `cmd/explorer/frontend/` (not at repo root)

### Runtime Issues

**Problem:** App won't launch on macOS (blank screen or immediate quit)

**Solutions:**
1. Remove quarantine attribute: `xattr -cr build/bin/explorer.app`
2. Rebuild with clean: `wails build -clean`
3. Check database connectivity: ensure `DATABASE_URL` is correct

---

**Problem:** "Connection refused" database error

**Solutions:**
1. Verify PostgreSQL is running: `pg_isready`
2. Check DATABASE_URL format
3. Test connection: `psql "$DATABASE_URL"`

---

**Problem:** GraphQL/Graph queries return empty

**Solution:** Ensure data is loaded and schema is migrated:
```bash
# Load emails with entity extraction
go run cmd/loader/main.go --extract

# Promote entities to formal schema
go run cmd/promoter/main.go promote person
```

## Performance Optimization

### Build Optimization

```bash
# Strip debug symbols (smaller binary)
wails build -clean -ldflags "-s -w"

# Compress with UPX (reduces size by ~60%)
wails build -clean -upx

# Enable obfuscation (code protection)
wails build -clean -obfuscated
```

### Frontend Optimization

**Vite Configuration** (`cmd/explorer/frontend/vite.config.ts`):

```typescript
export default defineConfig({
  build: {
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,  // Remove console.log
      },
    },
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          graph: ['d3-force', 'd3-zoom'],
        },
      },
    },
  },
});
```

### Database Optimization

For production deployments:
- Create indexes on frequently queried columns
- Use connection pooling (configure in `DATABASE_URL`)
- Consider read replicas for multi-user scenarios

## Security Considerations

### Code Signing

**macOS:** Required for distribution outside App Store
- Get Apple Developer ID certificate
- Sign with `codesign` as shown above
- Notarize for macOS 10.15+

**Windows:** Recommended for avoiding SmartScreen warnings
- Get code signing certificate (Sectigo, DigiCert, etc.)
- Sign with `signtool.exe`

### Application Security

- **Database credentials:** Never hardcode, use environment variables or secure config
- **API keys:** Store in secure keychain/credential manager
- **Updates:** Implement secure update mechanism (signed updates)

## Monitoring and Logging

### Application Logs

Wails apps log to stdout/stderr by default:

```bash
# Redirect logs to file
./explorer 2>&1 | tee explorer.log

# Or use system logging (macOS)
./explorer 2>&1 | logger -t "EnronExplorer"
```

### Error Reporting

Consider integrating:
- **Sentry** - Error tracking and performance monitoring
- **Application Insights** - Microsoft Azure monitoring
- Custom logging to file or remote service

## Version Management

### Semantic Versioning

Update version in multiple locations:

1. **wails.json:**
   ```json
   {
     "version": "1.0.0"
   }
   ```

2. **package.json (frontend):**
   ```json
   {
     "version": "1.0.0"
   }
   ```

3. **Build script:**
   ```bash
   wails build -ldflags "-X main.version=1.0.0"
   ```

### Release Process

1. **Update version** in wails.json and package.json
2. **Update CHANGELOG.md** with release notes
3. **Create git tag:** `git tag -a v1.0.0 -m "Release 1.0.0"`
4. **Build for all platforms**
5. **Sign and notarize** (macOS/Windows)
6. **Create GitHub release** with binaries
7. **Push tag:** `git push origin v1.0.0`

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Build
        run: |
          cd cmd/explorer
          wails build -clean -platform darwin/universal
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: explorer-macos
          path: cmd/explorer/build/bin/explorer.app

  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - uses: actions/setup-node@v3
      - name: Install dependencies
        run: sudo apt-get update && sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Build
        run: |
          cd cmd/explorer
          wails build -clean
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: explorer-linux
          path: cmd/explorer/build/bin/explorer

  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - uses: actions/setup-node@v3
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Build
        run: |
          cd cmd/explorer
          wails build -clean
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: explorer-windows
          path: cmd/explorer/build/bin/explorer.exe
```

## Support and Maintenance

### Update Strategy

- **Dependencies:** Regularly update Go modules, npm packages, Wails CLI
- **Security patches:** Monitor CVEs and update promptly
- **Database migrations:** Version and test thoroughly before deployment

### Backup and Recovery

- **Database:** Regular PostgreSQL backups (pg_dump)
- **Configuration:** Version control all config files
- **User data:** If app stores local preferences, document location

## Additional Resources

- **Wails Documentation:** https://wails.io/docs/
- **Wails CLI Reference:** https://wails.io/docs/reference/cli
- **Project README:** See `/README.md` for quick start guide
- **API Documentation:** See `/specs/002-graph-explorer/contracts/`
