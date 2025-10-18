# RubberPipe

Dead simple backup orchestrator.

**Version:** v0.1 (pre-alpha)

## What is RubberPipe?

RubberPipe is a lightweight tool to backup and restore data from self-hosted applications.  
It currently supports:

- **Sources:** PostgreSQL  
- **Destinations:** Local filesystem  

RubberPipe is designed with modularity in mind. Sources produce backup files, destinations store them, and the Hub orchestrates everything. Configs and backup metadata are stored in SQLite for simplicity.


## Quick Start

### Backup

```bash
rpipe backup <source_config> <destination_config>
# Example:
rpipe backup postgres_main local_main
```

### Restore

```bash
rpipe restore <backup_id>
# Example:
rpipe restore 3
```
### View Backup History

```bash
rpipe list
```

### Manage Adapter Configs

```bash
rpipe config list
rpipe config add <name> <type> <json_config>
rpipe config remove <name>
```




## Status

- [x] Core backup & restore working 
- [x] Configurable adapters (Postgres & Local) 
- [x] SQLite logging 
- [x] CLI Working
- [ ] **Not yet implemented:** Docker volume backups, incremental backups, scheduling, Web UI  

Use with caution—pre-alpha, APIs unstable.


## Next Steps (v0.2+)

- Add Docker volume adapter  
- Individual adapter documentation
- Incremental backups & validation  
- Retention policies & scheduling  
- Web UI for monitoring & restore  
- Plugin/adapter registry for 3rd-party sources & destinations  


## Installation

Clone the repo and run:

```bash
go run ./cmd/rubberpipe/main.go
```

Or build a binary:

```bash
go build -o rpipe ./cmd/rubberpipe
```
