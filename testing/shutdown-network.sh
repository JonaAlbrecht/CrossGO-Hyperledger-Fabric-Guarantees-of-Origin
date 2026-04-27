#!/bin/bash
# Shutdown HLF Network and Clean Blockchain Data

echo "====== Shutting Down HLF Network ======"

# Stop all chaincode containers
echo "Stopping chaincode containers..."
docker stop $(docker ps -q --filter "name=dev-peer") 2>/dev/null

# Stop all peer containers
echo "Stopping peer containers..."
docker stop $(docker ps -q --filter "name=peer0.") 2>/dev/null

# Stop all orderer containers
echo "Stopping orderer containers..."
docker stop $(docker ps -q --filter "name=orderer") 2>/dev/null

# Stop all couchdb containers
echo "Stopping couchdb containers..."
docker stop $(docker ps -q --filter "name=couchdb") 2>/dev/null

# Remove all containers
echo "Removing all stopped containers..."
docker rm $(docker ps -aq --filter "name=dev-peer") 2>/dev/null
docker rm $(docker ps -aq --filter "name=peer0.") 2>/dev/null
docker rm $(docker ps -aq --filter "name=orderer") 2>/dev/null
docker rm $(docker ps -aq --filter "name=couchdb") 2>/dev/null

# Remove volumes (blockchain data)
echo "Removing docker volumes (blockchain data)..."
docker volume prune -f

# Clean up production directories (peer ledger data)
echo "Cleaning up peer production data..."
cd /root/hlf-go/repo/network/organizations/peerOrganizations
for org in */; do
    for peer in ${org}peers/*/; do
        if [ -d "${peer}production" ]; then
            echo "Removing ${peer}production"
            rm -rf ${peer}production
        fi
    done
done

# Clean up orderer production directories
echo "Cleaning up orderer production data..."
cd /root/hlf-go/repo/network/organizations/ordererOrganizations
for org in */; do
    for orderer in ${org}orderers/*/; do
        if [ -d "${orderer}production" ]; then
            echo "Removing ${orderer}production"
            rm -rf ${orderer}production
        fi
    done
done

echo ""
echo "====== Network Shutdown Complete ======"
echo "All containers stopped and removed"
echo "All blockchain data deleted"
echo "Network ready for fresh restart"
