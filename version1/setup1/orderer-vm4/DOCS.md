# orderer-vm4 — Raft Ordering Service

This folder configures the **ordering service** for the GO Fabric network.

## Components
- **4 Raft orderer nodes** running etcdraft consensus:
  | Node | Host | Port (gRPC) | Port (Admin) |
  |------|------|------------|--------------|
  | orderer1 | `orderer.GOnetwork.com` | 7050 | 7053 |
  | orderer2 | `orderer2.GOnetwork.com` | 8050 | 8053 |
  | orderer3 | `orderer3.GOnetwork.com` | 9050 | 9053 |
  | orderer4 | `orderer4.GOnetwork.com` | 10050 | 10053 |

- **Fabric CA** (port 9054) for orderer identity enrollment

## Files
- **`.env`** — Defines `IMAGE_TAG`, `CA_IMAGE_TAG`, `COMPOSE_PROJECT_NAME`
- **`docker-compose.yaml`** — Docker Compose defining orderer CA + 4 orderer containers with TLS, genesis block mounts, and persistent volumes
- **`create-cryptomaterial-orderer/`** — Scripts to enroll orderer admin and register orderer node identities via the Fabric CA
