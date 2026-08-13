#!/bin/bash

[ -d pgedge-postgres-mcp ] || git clone https://github.com/pgEdge/pgedge-postgres-mcp.git
cd pgedge-postgres-mcp
make build-server
sudo cp bin/pgedge-postgres-mcp /usr/local/bin/pgedge-postgres-mcp


