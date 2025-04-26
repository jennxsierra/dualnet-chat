# Proposed Project Structure

```markdown
  dualnet-chat/
  │
  ├─ cmd/
  │  ├─ tcp/
  │  │   ├─ server/
  │  │   │   └─ main.go
  │  │   └─ client/
  │  │       └─ main.go
  │  └─ udp/
  │      ├─ server/
  │      │   └─ main.go
  │      └─ client/ 
  │          └─ main.go
  │
  ├─ internal/
  │  ├─ chat/                      # Core logic: user registry, broadcaster, message type
  │  │   ├─ hub.go
  │  │   └─ message.go
  │  ├─ tcp/                       # Helpers that wrap `net.TCPConn` read/write details
  │  │   └─ transport.go
  │  └─ udp/                       # Similar wrappers for `net.UDPConn`
  │      └─ transport.go
  │
  ├─ tests/                        # (Optional) Unit tests
  │  └─ chat_test.go
  │
  ├─ Makefile
  ├─ go.mod
  ├─ README.md
  ├─ LICENSE
  └─ .gitignore
```
