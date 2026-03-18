# Contributing to SysMind

Thank you for your interest in contributing to SysMind! We welcome contributions from everyone.

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- Wails CLI v2
- Git

### Development Setup
1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/Anuolu-2020/sysmind.git
   cd sysmind
   ```
3. Install dependencies:
   ```bash
   cd frontend && npm install && cd ..
   go mod tidy
   ```
4. Run in development mode:
   ```bash
   wails dev
   ```

## 📋 How to Contribute

### Reporting Bugs
1. Check existing issues to avoid duplicates
2. Create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - System information (OS, Go version, etc.)
   - Screenshots if applicable

### Suggesting Features
1. Check existing feature requests
2. Open a new discussion or issue
3. Describe the feature and its benefits
4. Provide mockups or examples if helpful

### Code Contributions

#### Branch Naming
- `feat/feature-name` - New features
- `fix/bug-description` - Bug fixes
- `docs/update-description` - Documentation
- `refactor/component-name` - Code refactoring

#### Commit Messages
We use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Examples:
- `feat(ai): add support for Anthropic Claude`
- `fix(timeline): resolve canvas rendering issue`
- `docs: update installation instructions`

#### Pull Request Process
1. Create a feature branch from `main`
2. Make your changes with clear, focused commits
3. Add or update tests as needed
4. Update documentation if required
5. Ensure all tests pass: `go test ./...`
6. Create a pull request with:
   - Clear title and description
   - Link to related issues
   - Screenshots for UI changes
   - Testing notes

## 🧪 Testing

### Running Tests
```bash
# Run Go tests
go test ./...

# Run with coverage
go test -cover ./...

# Run frontend tests
cd frontend && npm test
```

### Test Guidelines
- Write tests for new features
- Maintain or improve code coverage
- Test cross-platform compatibility when possible
- Include unit tests for utility functions
- Add integration tests for complex features

## 📝 Code Style

### Go Code
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions focused and small
- Handle errors appropriately

### Frontend Code
- Use consistent component structure
- Follow React best practices
- Use descriptive component and prop names
- Add JSDoc comments for complex functions
- Keep components focused on single responsibility

### File Organization
```
internal/
├── ai/           # AI provider implementations
├── collectors/   # Platform-specific data collection
├── models/       # Data structures and types
└── services/     # Business logic

frontend/src/
├── components/   # React components
├── contexts/     # React contexts
└── utils/        # Utility functions
```

## 🌍 Platform-Specific Development

### Linux Development
- Test with different distributions
- Use `/proc` filesystem appropriately
- Handle permission requirements

### macOS Development  
- Test on different macOS versions
- Use proper system APIs
- Handle sandboxing requirements

### Windows Development
- Test on different Windows versions
- Use appropriate WMI queries
- Handle UAC and permissions

## 🔍 Code Review Guidelines

### For Authors
- Keep PRs focused and reasonably sized
- Provide context and testing instructions
- Respond promptly to feedback
- Be open to suggestions and improvements

### For Reviewers
- Be constructive and respectful
- Focus on code quality and maintainability
- Test the changes locally when possible
- Approve when ready, request changes when needed

## 📦 Release Process

### Version Numbering
We use [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes (backward compatible)

### Release Checklist
- [ ] Update version numbers
- [ ] Update CHANGELOG.md
- [ ] Test on all supported platforms
- [ ] Create GitHub release
- [ ] Update documentation

## 🎯 Areas for Contribution

### High Priority
- 🐛 Bug fixes and stability improvements
- 🚀 Performance optimizations
- 📖 Documentation improvements
- 🧪 Test coverage expansion

### Feature Ideas
- 🔌 Additional AI provider integrations
- 📊 New monitoring metrics and visualizations
- 🌍 Internationalization support
- 🔒 Enhanced security features
- 📱 Mobile companion app
- 🐳 Container monitoring integration

### Community
- 📝 Blog posts and tutorials
- 🎥 Video demonstrations
- 🗣️ Conference presentations
- 🎨 UI/UX improvements

## ❓ Questions?

- 💬 GitHub Discussions for general questions
- 🐛 GitHub Issues for bugs and feature requests
- 📧 Email maintainers for sensitive issues

## 🙏 Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes
- Documentation credits
- Special mentions for significant contributions

Thank you for making SysMind better! 🎉