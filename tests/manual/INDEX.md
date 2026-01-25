# Manual Testing Quick Reference

## 🚀 Quick Start (3 Steps)

```bash
# 1. Verify environment is ready
./tests/manual/verify_environment.sh

# 2. Review what to test
cat tests/manual/README.md

# 3. Launch TUI and start testing
./tui
```

## 📋 Test Documents

| Task | Document | Description |
|------|----------|-------------|
| T108 | [T108_TUI_NAVIGATION_FLOWS.md](./T108_TUI_NAVIGATION_FLOWS.md) | Navigation, filtering, searching, graph visualization |
| T109 | [T109_ERROR_HANDLING.md](./T109_ERROR_HANDLING.md) | Empty DB, network loss, invalid data, edge cases |

## 🛠️ Test Scripts

| Script | Purpose |
|--------|---------|
| [verify_environment.sh](./verify_environment.sh) | Check prerequisites before testing |
| [run_manual_tests.sh](./run_manual_tests.sh) | Guided test execution with instructions |

## 📖 Documentation

- [README.md](./README.md) - Complete manual testing guide
- [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md) - What was implemented and why

## ✅ Checklist Before Testing

- [ ] Database is running (`docker ps | grep postgres`)
- [ ] Entities exist (`./tests/manual/verify_environment.sh`)
- [ ] TUI is built (`ls -lh tui`)
- [ ] Test documents are reviewed
- [ ] Ready to document results

## 🎯 What to Test

### T108: Navigation (6 test cases)
1. Entity list navigation
2. Type filtering
3. Name search
4. Entity details
5. Graph visualization
6. Graph navigation & expansion

### T109: Error Handling (6 test cases)
1. Empty database
2. Network loss during session
3. Network loss at startup
4. Invalid selections
5. Malformed data
6. Large datasets (10k+ entities)

## 📊 Recording Results

Mark results directly in the test documents:
- Fill in "Actual Results"
- Check "Pass/Fail" boxes
- Note any issues or observations
- Complete summary sections

## 🔧 Common Commands

```bash
# Check entity count
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT COUNT(*) FROM discovered_entities;"

# Check entity types
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT DISTINCT type_category FROM discovered_entities;"

# Rebuild TUI
go build -o tui cmd/tui/main.go

# Stop database (for error testing)
docker-compose stop postgres

# Start database
docker-compose up -d postgres
```

## 📈 Success Criteria

- **All navigation flows work smoothly** (T108)
- **All errors handled gracefully** (T109)
- **Performance**: <3s render for 500 nodes (SC-007)
- **No crashes or hangs**
- **Clear error messages**

## ❓ Need Help?

1. Check [README.md](./README.md) for troubleshooting
2. Review main spec: `specs/001-cognitive-backbone-poc/spec.md`
3. Check TUI code: `internal/tui/`

---

**Status**: Ready for manual test execution  
**Last Updated**: January 24, 2026
