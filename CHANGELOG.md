## [0.2.1] - 2026-03-19

### Changes
- Merge pull request #22 from Anuolu-2020/dev (a4e2f41)
- chore: bump version to 0.1.13 (177c5fe)
- fix(windows): hide console windows when executing external commands (a447417)

## [0.2.0] - 2026-03-19

### Changes
- Merge pull request #21 from Anuolu-2020/dev (f4ab679)
- fix: resolve YAML syntax error in release workflow changelog generation (67bcc81)
- fix: rewrite changelog generation using github-script to avoid YAML escaping issues (9c46eb2)
- Merge pull request #20 from Anuolu-2020/dev (8f8bbc0)
- fix: resolve YAML syntax error by using quoted heredoc and placeholder replacement (58fce59)
- Merge pull request #19 from Anuolu-2020/main (4cb4d94)
- Merge branch 'dev' (56f941e)
- fix: resolve YAML syntax error in release.yml heredoc (d8ba7a1)
- Merge pull request #18 from Anuolu-2020/main (2d6c817)

## [0.1.13] - 2026-03-19

### Changes
- Merge branch 'dev' (c247a5a)
- fix: repair changelog generation and remove linux-arm64 from release body (bfc29ba)
- Merge pull request #17 from Anuolu-2020/main (3401e47)

## [0.1.12] - 2026-03-19

### Changes
- chore: remove linux-arm64 from release matrix (a528a9b)

## [0.1.11] - 2026-03-19

### Changes
- Merge pull request #16 from Anuolu-2020/dev (26d0316)
- fix: resolve merge conflict and ensure frontend build in release workflow (d9736b3)
- Merge pull request #15 from Anuolu-2020/main (4e6f524)
- Merge pull request #14 from Anuolu-2020/dev (87d11ce)
- docs: make installation instructions dynamic for latest releases (d6a7169)
- Merge pull request #13 from Anuolu-2020/main (a36e5c9)
- Merge pull request #12 from Anuolu-2020/main (3f4727b)
- Merge pull request #11 from Anuolu-2020/main (c967f57)
- Merge pull request #10 from Anuolu-2020/main (f10eb62)
- Merge pull request #9 from Anuolu-2020/main (66f9fb9)

## [0.1.10] - 2026-03-19

### Changes
- fix: handle macOS .app bundle and standalone binary in release packaging (50c0eea)

## [0.1.9] - 2026-03-19

### Changes
- fix: ensure Build application step uses bash on all platforms (6cd059c)

## [0.1.8] - 2026-03-19

### Changes
- fix: set shell to bash for icon preparation in release workflow (faf0006)

## [0.1.7] - 2026-03-19

### Changes
- fix: resolve windows build shell and redirection error (1232062)

## [0.1.6] - 2026-03-19

### Changes
- fix: linux build dependency and runner mapping (0643fb7)

## [0.1.5] - 2026-03-19

### Changes
- Merge pull request #8 from Anuolu-2020/dev (7d6caef)
- fix: skip UPX compression on macOS release builds (8f14fa6)
- Merge pull request #7 from Anuolu-2020/main (db4a669)
- Merge pull request #6 from Anuolu-2020/main (a90353a)

## [0.1.4] - 2026-03-19

### Changes
- fix: use arm64 runner for linux-arm64 release builds (d4c8c1a)

## [0.1.3] - 2026-03-19

### Changes
- fix: update WebKit dependency for Ubuntu 24.04 compatibility (b374f89)

## [0.1.2] - 2026-03-19

### Changes
- Merge pull request #5 from Anuolu-2020/dev (933db5d)
- fix: update Go version to 1.22 in release workflow (7266456)
- fix: correct YAML syntax errors in release workflow (8070a6a)
- Merge pull request #4 from Anuolu-2020/main (d812980)
- fix: add Linux GTK dependencies to release workflow (0cd8a49)
- Merge pull request #3 from Anuolu-2020/main (3235ccc)

## [0.1.1] - 2026-03-19

### Changes
- fix: add permissions to GitHub Actions release workflow (1bd3930)

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-19

### Changes
- Merge pull request #2 from Anuolu-2020/dev (f2ae086)
- chore: add RELEASE.md to .gitignore (833e492)
- docs: Update README with logo instead of ASCII art (e0ad50a)
- feat: copy icons to build directory (7173d87)
- feat: add application icons (a323b37)
- docs(README): update README with ASCII art and emoji headings (b898950)
- feat: pass privacy config to GenerateResponse (b25932a)
- feat: Add privacy and data sharing documentation (b12c652)
- feat: add privacy configuration (c5dc166)
- feat(models): add PrivacyConfig struct for AI data sharing settings (6432495)
- feat(ai): Add privacy configuration to AI providers (dc8233e)
- chore: Remove unnecessary Babel dependencies (c0a1975)
- feat(models): add PrivacyConfig model (679b493)
- feat(wailsjs): add privacy config functions (f7b0575)
- feat: add Privacy Settings component (28fb59c)
- Merge pull request #1 from Anuolu-2020/dev (2c40ed8)
- chore: remove codeql sarif upload (4578b1c)
- fix: replace deprecated gosec action with direct installation (4ca83a1)
- fix: resolve all golangci-lint errors (7862a49)
- fix: build frontend before golangci-lint in CI workflow (910efa2)

