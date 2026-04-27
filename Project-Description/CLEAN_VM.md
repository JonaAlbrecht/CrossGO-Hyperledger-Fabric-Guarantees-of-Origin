# Clean Virtual Machine

Reusable guide to tear down the Hyperledger Fabric network and clean generated artifacts from the Hetzner VM (`204.168.234.18`).

## Prerequisites

SSH into the VM:

```bash
ssh -i ~/.ssh/hetzner root@204.168.234.18
```

Set PATH (needed if shell is non-interactive):

```bash
export PATH=$PATH:/bin:/usr/bin:/usr/local/bin
```

## 1. Stop the Application (Backend + Frontend)

```bash
fuser -k 3001/tcp 2>/dev/null || true   # Express backend
fuser -k 5173/tcp 2>/dev/null || true   # Vite frontend
```

## 2. Tear Down the HLF Network

The existing `network-down.sh` stops all Docker Compose stacks (orderers, peers, CouchDB, CAs), removes chaincode containers/images, and deletes `organizations/` and `channel-artifacts/`.

```bash
cd /root/hlf-go/repo/network
bash scripts/network-down.sh
```

**What `network-down.sh` does:**
- `docker compose down --volumes --remove-orphans` on all 6 compose files (buyer, hproducer, eproducer, issuer, orderer, ca)
- Removes `dev-peer*` chaincode containers and images
- Deletes `organizations/` (crypto material) and `channel-artifacts/` (genesis block, channel tx)

## 3. Deep Clean — Orphan Volumes, Stopped Containers, Unused Images

After `network-down.sh`, leftover Docker resources may remain from previous runs.

```bash
# Remove stopped containers (holds references to images)
docker container prune -f

# Remove orphan/anonymous volumes
docker volume prune -f

# Remove ALL unused images (chaincode build layers, dangling <none> images)
docker image prune -af
```

> **Note:** `docker image prune -af` removes **all** unused images, including base Fabric images (peer, orderer, ca, couchdb, baseos, ccenv). These will be re-pulled automatically by `docker compose up` when you boot the network again.

To keep base images and only remove dangling ones:

```bash
docker image prune -f   # only removes <none> dangling images
```

## 4. Verify Clean State

```bash
echo "=== Containers ===" && docker ps -a
echo "=== Volumes ===" && docker volume ls
echo "=== Images ===" && docker images
echo "=== Crypto ===" && ls /root/hlf-go/repo/network/organizations 2>&1 || echo "organizations/ removed ✓"
echo "=== Channel artifacts ===" && ls /root/hlf-go/repo/network/channel-artifacts 2>&1 || echo "channel-artifacts/ removed ✓"
echo "=== Disk ===" && df -h /
```

Expected result: zero containers, zero volumes, zero images (or only base images if you used `-f` instead of `-af`), no `organizations/` or `channel-artifacts/` directories.

## What Is Preserved

| Kept | Path | Size |
|------|------|------|
| Source code | `/root/hlf-go/repo/` | — |
| Fabric binaries | `/root/hlf-go/repo/fabric-bin/` | ~242 MB |
| Backend dependencies | `/root/hlf-go/repo/backend/node_modules/` | ~75 MB |
| Frontend dependencies | `/root/hlf-go/repo/frontend/node_modules/` | ~129 MB |
| Docker engine | system | — |
| Network config files | `network/docker/*.yaml`, `network/*.yaml` | — |
| Scripts | `network/scripts/` | — |

## What Is Removed

| Removed | Description |
|---------|-------------|
| `network/organizations/` | Crypto material (MSP certs, TLS keys) |
| `network/channel-artifacts/` | Genesis block, channel transaction |
| Docker containers | All Fabric peers, orderers, CouchDB, CAs, chaincode |
| Docker volumes | Ledger data, CouchDB state, orderer WAL |
| Docker images (deep clean) | Chaincode build images, base Fabric images |

## Re-bootstrap After Clean

```bash
cd /root/hlf-go/repo
bash boot-network.sh
```

This regenerates crypto, starts containers, creates the channel, joins peers, deploys chaincode, and initializes the ledger.

Then start the app:

```bash
cd /root/hlf-go/repo/backend && node src/server.js &
cd /root/hlf-go/repo/frontend && npx vite --host 0.0.0.0 &
```

## One-Liner: Full Clean

```bash
cd /root/hlf-go/repo/network && bash scripts/network-down.sh && \
docker container prune -f && docker volume prune -f && docker image prune -af
```
