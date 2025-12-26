# mcp-go Documentation

Welcome to the mcp-go documentation! This directory contains comprehensive guides and references for using mcp-go.

## 📚 Documentation Structure

### [Architecture](architecture.md)
**Deep dive into mcp-go's design and internals**

Learn about:
- Layered architecture design
- Transport-agnostic approach
- Component responsibilities
- Message flow patterns
- Extension points
- Performance considerations
- Security best practices

Perfect for:
- Understanding how mcp-go works internally
- Contributing to the project
- Building custom transports
- Architectural decision-making

### [Transport Guide](transport-guide.md)
**Choose and use the right transport for your needs**

Covers:
- Quick decision matrix
- Feature comparison
- Detailed transport guides (Stdio, SSE, HTTP)
- Migration strategies
- Troubleshooting
- Performance tuning

Perfect for:
- Deciding which transport to use
- Learning transport-specific features
- Migrating between transports
- Solving transport issues

### [API Reference](api-reference.md)
**Complete API documentation**

Includes:
- Server API (registration, startup)
- Client API (calls, queries)
- Type definitions
- Schema generation
- Error handling
- Best practices
- Complete examples

Perfect for:
- Quick API lookups
- Understanding function signatures
- Learning proper usage patterns
- Finding example code

## 🚀 Quick Start

### For New Users

1. **Start with the Transport Guide** to choose your transport
2. **Read the API Reference** for implementation details
3. **Check the Architecture doc** if you need deeper understanding

### For Contributors

1. **Start with the Architecture doc** to understand the design
2. **Read the Transport Guide** to understand transport patterns
3. **Reference the API doc** for implementation details

## 📖 Additional Resources

- **[Main README](../README.md)**: Project overview and quick start
- **[Examples](../examples/README.md)**: Working code examples for each transport
- **[Tests](../test/)**: Unit tests showing usage patterns

## 💡 Learning Paths

### Building a CLI Tool

```
Transport Guide → Stdio section → API Reference → Examples
```

### Building a Web Service

```
Transport Guide → SSE or HTTP section → API Reference → Examples
```

### Contributing a New Transport

```
Architecture → Transport Layer → Transport Interface → Existing Implementations
```

### Integrating with Claude Desktop

```
Transport Guide → Stdio section → Claude Desktop Integration
```

## 🤝 Contributing to Documentation

Documentation improvements are welcome! When contributing:

1. Keep examples working and up-to-date
2. Update all affected sections when APIs change
3. Add diagrams where helpful
4. Include code examples for clarity
5. Test code snippets before committing

## 📝 Documentation Standards

- **Code Examples**: Always use complete, runnable examples
- **Diagrams**: Use ASCII art for diagrams (portable, version-controllable)
- **Links**: Use relative links within documentation
- **Formatting**: Follow standard Markdown conventions
- **Language**: Clear, concise, technical but accessible

## 🔍 Need Help?

- Check the relevant doc section first
- Look at working examples in `../examples/`
- Review unit tests in `../test/`
- Open an issue on GitHub if stuck

## 📊 Documentation Coverage

| Topic | Documentation | Examples | Tests |
|-------|--------------|----------|-------|
| Server API | ✅ Complete | ✅ Yes | ✅ Yes |
| Client API | ✅ Complete | ✅ Yes | ✅ Yes |
| Stdio Transport | ✅ Complete | ✅ Yes | ⚠️ Manual |
| SSE Transport | ✅ Complete | ✅ Yes | ⚠️ Manual |
| HTTP Transport | ✅ Complete | ✅ Yes | ⚠️ Manual |
| Schema Generation | ✅ Complete | ✅ Yes | ✅ Yes |
| Mock Transport | ✅ Complete | ✅ Yes | ✅ Yes |

## 🎯 Next Steps

Ready to build with mcp-go? Here's what to do:

1. **Choose your transport** using the [Transport Guide](transport-guide.md)
2. **Review the API** in the [API Reference](api-reference.md)
3. **Check examples** in [`../examples/`](../examples/)
4. **Build something awesome!** 🚀

---

*Last updated: December 2025*
