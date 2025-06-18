
# ECHO-SVELTE

A project creation tool for svelte based webapps.

### Installation
```bash
go install github.com/tonyv2576/echo-svelte 
```
---
### Creating a new project (required flags):
```bash
echo-svelte -name project_name
```
**Note**: the default `-port` is `8080`.

### Creating a new project with optional flags:
```bash
echo-svelte -name project_name -dir projects/golang -port 3000 -ts
```
**Tip**: Setting the `-dir` flag to `./` will use the current directory instead of making a new folder and if the project name isn't set, it'll use the folder name instead.