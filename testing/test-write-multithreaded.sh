#!/bin/bash
# Multi-threaded performance benchmarking for write operations using Fabric Peer CLI
# Tests: AddToBacklogElectricity, CreateElectricityGO, LockGOForConversion
# Uses GNU parallel or background jobs for concurrent execution

set -e

export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Test configuration
NUM_TRANSACTIONS=30
NUM_WORKERS=5  # Concurrent workers

echo "=========================================="
echo "Multi-threaded Write Operation Benchmark"
echo "=========================================="
echo "Transactions per function: $NUM_TRANSACTIONS"
echo "Concurrent workers: $NUM_WORKERS"
echo ""

# Create temp directory for results
TEMP_DIR="/tmp/hlf-perf-$$"
mkdir -p "$TEMP_DIR"

# Function to execute transaction and measure latency
execute_transaction() {
    local func_name=$1
    local identity=$2
    local tx_data=$3
    local tx_id=$4
    local result_file=$5
    
    local tx_start=$(date +%s%N)
    
    # Set identity based on function requirements
    case $identity in
        "eproducer1")
            export CORE_PEER_LOCALMSPID="eproducer1MSP"
            export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
            export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/users/Admin@eproducer1.go-platform.com/msp
            export CORE_PEER_ADDRESS=localhost:9051
            ;;
        "eissuer")
            export CORE_PEER_LOCALMSPID="eissuerMSP"
            export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
            export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/users/Admin@eissuer.go-platform.com/msp
            export CORE_PEER_ADDRESS=localhost:7051
            ;;
    esac
    
    # Execute transaction
    if /root/hlf-go/repo/fabric-bin/bin/peer chaincode invoke \
        -o localhost:7050 \
        --ordererTLSHostnameOverride orderer1.go-platform.com \
        --tls --cafile "$ORDERER_CA" \
        -C electricity-de -n golifecycle \
        -c "$func_name" \
        --transient "$tx_data" \
        > /dev/null 2>&1; then
        
        local tx_end=$(date +%s%N)
        local latency_ns=$((tx_end - tx_start))
        local latency_ms=$((latency_ns / 1000000))
        
        echo "SUCCESS $latency_ms" >> "$result_file"
    else
        echo "FAIL 0" >> "$result_file"
    fi
}

# Export function for parallel execution
export -f execute_transaction
export FABRIC_CFG_PATH CORE_PEER_TLS_ENABLED ORDERER_CA

# Test 1: AddToBacklogElectricity (producer identity)
echo -e "${BLUE}Test 1: AddToBacklogElectricity (producer role)${NC}"
echo "========================================"

BACKLOG_RESULT="$TEMP_DIR/backlog_results.txt"
> "$BACKLOG_RESULT"

start_time=$(date +%s%N)

# Generate transaction commands
for i in $(seq 1 $NUM_TRANSACTIONS); do
    amount=$(echo "scale=2; 1 + ($RANDOM % 100) / 10" | bc)
    emissions=$(echo "scale=3; $RANDOM / 10000" | bc)
    elapsed=$((3600 + RANDOM % 3600))
    
    BACKLOG_DATA="{\"AmountMWh\":\"$amount\",\"Emissions\":\"$emissions\",\"ElectricityProductionMethod\":\"solar_pv\",\"ElapsedSeconds\":\"$elapsed\"}"
    FUNC_CALL="{\"function\":\"backlog:AddToBacklogElectricity\",\"Args\":[]}"
    TRANSIENT="{\"eBacklog\":\"$(echo -n "$BACKLOG_DATA" | base64 -w0)\"}"
    
    # Execute in background with worker limit
    while [ $(jobs -r | wc -l) -ge $NUM_WORKERS ]; do
        sleep 0.01
    done
    
    execute_transaction "$FUNC_CALL" "eproducer1" "$TRANSIENT" "$i" "$BACKLOG_RESULT" &
done

# Wait for all background jobs
wait

end_time=$(date +%s%N)

# Calculate metrics
total_time_ns=$((end_time - start_time))
total_time_ms=$((total_time_ns / 1000000))
success_count=$(grep "SUCCESS" "$BACKLOG_RESULT" | wc -l)
fail_count=$(grep "FAIL" "$BACKLOG_RESULT" | wc -l)

if [ $success_count -gt 0 ]; then
    total_latency=$(grep "SUCCESS" "$BACKLOG_RESULT" | awk '{sum+=$2} END {print sum}')
    avg_latency=$((total_latency / success_count))
    min_latency=$(grep "SUCCESS" "$BACKLOG_RESULT" | awk '{print $2}' | sort -n | head -1)
    max_latency=$(grep "SUCCESS" "$BACKLOG_RESULT" | awk '{print $2}' | sort -n | tail -1)
    tps=$(echo "scale=2; $success_count * 1000 / $total_time_ms" | bc)
else
    avg_latency=0
    min_latency=0
    max_latency=0
    tps=0
fi

echo -e "${GREEN}✓${NC} Completed: $success_count/$NUM_TRANSACTIONS successful"
echo "  TPS: $tps | Avg: ${avg_latency}ms | Min: ${min_latency}ms | Max: ${max_latency}ms"
echo ""

# Store results
BACKLOG_TPS=$tps
BACKLOG_AVG=$avg_latency
BACKLOG_MIN=$min_latency
BACKLOG_MAX=$max_latency
BACKLOG_SUCC=$success_count
BACKLOG_FAIL=$fail_count

# Test 2: CreateElectricityGO (producer with 3-org endorsement)
echo -e "${BLUE}Test 2: CreateElectricityGO (producer + SBE policy)${NC}"
echo "========================================"

GO_RESULT="$TEMP_DIR/go_results.txt"
> "$GO_RESULT"

start_time=$(date +%s%N)

# CreateElectricityGO requires endorsement from all 3 orgs (SBE policy)
for i in $(seq 1 $NUM_TRANSACTIONS); do
    amount=$(echo "scale=2; 1 + ($RANDOM % 100) / 10" | bc)
    emissions=$(echo "scale=3; $RANDOM / 10000" | bc)
    elapsed=$((3600 + RANDOM % 3600))
    
    GO_DATA="{\"AmountMWh\":\"$amount\",\"Emissions\":\"$emissions\",\"ElectricityProductionMethod\":\"solar_pv\",\"ElapsedSeconds\":\"$elapsed\"}"
    FUNC_CALL="{\"function\":\"issuance:CreateElectricityGO\",\"Args\":[]}"
    TRANSIENT="{\"eGO\":\"$(echo -n "$GO_DATA" | base64 -w0)\"}"
    
    # Execute in background
    while [ $(jobs -r | wc -l) -ge $NUM_WORKERS ]; do
        sleep 0.01
    done
    
    (
        # Set producer identity
        export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config
        export CORE_PEER_TLS_ENABLED=true
        export CORE_PEER_LOCALMSPID="eproducer1MSP"
        export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
        export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/users/Admin@eproducer1.go-platform.com/msp
        export CORE_PEER_ADDRESS=localhost:9051
        
        ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem
        
        tx_start=$(date +%s%N)
        
        # SBE requires peerAddresses from all 3 orgs
        if /root/hlf-go/repo/fabric-bin/bin/peer chaincode invoke \
            -o localhost:7050 \
            --ordererTLSHostnameOverride orderer1.go-platform.com \
            --tls --cafile "$ORDERER_CA" \
            -C electricity-de -n golifecycle \
            --peerAddresses localhost:7051 \
            --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
            --peerAddresses localhost:9051 \
            --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt \
            --peerAddresses localhost:13051 \
            --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt \
            -c "$FUNC_CALL" \
            --transient "$TRANSIENT" \
            > /dev/null 2>&1; then
            
            tx_end=$(date +%s%N)
            latency_ms=$(( (tx_end - tx_start) / 1000000 ))
            echo "SUCCESS $latency_ms" >> "$GO_RESULT"
        else
            echo "FAIL 0" >> "$GO_RESULT"
        fi
    ) &
done

wait

end_time=$(date +%s%N)

# Calculate metrics
total_time_ns=$((end_time - start_time))
total_time_ms=$((total_time_ns / 1000000))
success_count=$(grep "SUCCESS" "$GO_RESULT" | wc -l)
fail_count=$(grep "FAIL" "$GO_RESULT" | wc -l)

if [ $success_count -gt 0 ]; then
    total_latency=$(grep "SUCCESS" "$GO_RESULT" | awk '{sum+=$2} END {print sum}')
    avg_latency=$((total_latency / success_count))
    min_latency=$(grep "SUCCESS" "$GO_RESULT" | awk '{print $2}' | sort -n | head -1)
    max_latency=$(grep "SUCCESS" "$GO_RESULT" | awk '{print $2}' | sort -n | tail -1)
    tps=$(echo "scale=2; $success_count * 1000 / $total_time_ms" | bc)
else
    avg_latency=0
    min_latency=0
    max_latency=0
    tps=0
fi

echo -e "${GREEN}✓${NC} Completed: $success_count/$NUM_TRANSACTIONS successful"
echo "  TPS: $tps | Avg: ${avg_latency}ms | Min: ${min_latency}ms | Max: ${max_latency}ms"
echo ""

GO_TPS=$tps
GO_AVG=$avg_latency
GO_MIN=$min_latency
GO_MAX=$max_latency
GO_SUCC=$success_count
GO_FAIL=$fail_count

# Test 3: PublishOracleData (for comparison with multi-threading)
echo -e "${BLUE}Test 3: PublishOracleData (issuer, multi-threaded)${NC}"
echo "========================================"

ORACLE_RESULT="$TEMP_DIR/oracle_results.txt"
> "$ORACLE_RESULT"

start_time=$(date +%s%N)

for i in $(seq 1 $NUM_TRANSACTIONS); do
    now=$(date +%s)
    quantity=$((500 + RANDOM % 500))
    
    ORACLE_DATA="{\"CarrierType\":\"electricity\",\"Zone\":\"DE-LU\",\"PeriodStart\":$((now - 3600)),\"PeriodEnd\":$now,\"ProductionMethod\":\"solar_pv\",\"EnergyUnit\":\"MWh\",\"Quantity\":$quantity,\"EmissionFactor\":0,\"DataSource\":\"ENTSO-E-TP\",\"Attributes\":{}}"
    FUNC_CALL="{\"function\":\"oracle:PublishOracleData\",\"Args\":[]}"
    TRANSIENT="{\"OracleData\":\"$(echo -n "$ORACLE_DATA" | base64 -w0)\"}"
    
    while [ $(jobs -r | wc -l) -ge $NUM_WORKERS ]; do
        sleep 0.01
    done
    
    execute_transaction "$FUNC_CALL" "eissuer" "$TRANSIENT" "$i" "$ORACLE_RESULT" &
done

wait

end_time=$(date +%s%N)

total_time_ns=$((end_time - start_time))
total_time_ms=$((total_time_ns / 1000000))
success_count=$(grep "SUCCESS" "$ORACLE_RESULT" | wc -l)
fail_count=$(grep "FAIL" "$ORACLE_RESULT" | wc -l)

if [ $success_count -gt 0 ]; then
    total_latency=$(grep "SUCCESS" "$ORACLE_RESULT" | awk '{sum+=$2} END {print sum}')
    avg_latency=$((total_latency / success_count))
    min_latency=$(grep "SUCCESS" "$ORACLE_RESULT" | awk '{print $2}' | sort -n | head -1)
    max_latency=$(grep "SUCCESS" "$ORACLE_RESULT" | awk '{print $2}' | sort -n | tail -1)
    tps=$(echo "scale=2; $success_count * 1000 / $total_time_ms" | bc)
else
    avg_latency=0
    min_latency=0
    max_latency=0
    tps=0
fi

echo -e "${GREEN}✓${NC} Completed: $success_count/$NUM_TRANSACTIONS successful"
echo "  TPS: $tps | Avg: ${avg_latency}ms | Min: ${min_latency}ms | Max: ${max_latency}ms"
echo ""

ORACLE_TPS=$tps
ORACLE_AVG=$avg_latency
ORACLE_MIN=$min_latency
ORACLE_MAX=$max_latency
ORACLE_SUCC=$success_count
ORACLE_FAIL=$fail_count

# Summary table
echo "=========================================="
echo "Performance Results Summary"
echo "=========================================="
echo ""
printf "| %-30s | %-4s | %-4s | %-8s | %-10s | %-10s | %-10s |\n" \
    "Function" "Succ" "Fail" "TPS" "Avg (ms)" "Min (ms)" "Max (ms)"
echo "|--------------------------------|------|------|----------|------------|------------|------------|"

printf "| %-30s | %-4d | %-4d | %-8s | %-10s | %-10s | %-10s |\n" \
    "backlog:AddToBacklogElectricity" "$BACKLOG_SUCC" "$BACKLOG_FAIL" "$BACKLOG_TPS" "$BACKLOG_AVG" "$BACKLOG_MIN" "$BACKLOG_MAX"

printf "| %-30s | %-4d | %-4d | %-8s | %-10s | %-10s | %-10s |\n" \
    "issuance:CreateElectricityGO" "$GO_SUCC" "$GO_FAIL" "$GO_TPS" "$GO_AVG" "$GO_MIN" "$GO_MAX"

printf "| %-30s | %-4d | %-4d | %-8s | %-10s | %-10s | %-10s |\n" \
    "oracle:PublishOracleData (MT)" "$ORACLE_SUCC" "$ORACLE_FAIL" "$ORACLE_TPS" "$ORACLE_AVG" "$ORACLE_MIN" "$ORACLE_MAX"

echo ""
echo "=========================================="
echo "Configuration"
echo "=========================================="
echo "Transactions per function: $NUM_TRANSACTIONS"
echo "Concurrent workers: $NUM_WORKERS"
echo "Execution mode: Multi-threaded (background jobs)"
echo ""

# Cleanup
rm -rf "$TEMP_DIR"

echo -e "${GREEN}Multi-threaded performance benchmark completed!${NC}"
