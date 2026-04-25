#!/bin/bash
# Performance benchmarking for write operations using Fabric Peer CLI
# Measures TPS, average latency, and max latency for chaincode invocations

set -e

export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
NUM_TRANSACTIONS=50
ORACLE_FUNCTION="oracle:PublishOracleData"

echo "========================================"
echo "Write Operation Performance Benchmark"
echo "========================================"
echo "Function: $ORACLE_FUNCTION"
echo "Transactions: $NUM_TRANSACTIONS"
echo "Identity: eissuerMSP"
echo ""

# Setup issuer identity
export CORE_PEER_LOCALMSPID="eissuerMSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/users/Admin@eissuer.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

# Arrays to store timing data
declare -a latencies
success_count=0
fail_count=0
min_latency=999999
max_latency=0
total_latency=0

echo -e "${BLUE}Starting performance test...${NC}"
echo ""

# Record start time (nanoseconds)
start_time=$(date +%s%N)

# Execute transactions
for i in $(seq 1 $NUM_TRANSACTIONS); do
    # Generate unique oracle data
    now=$(date +%s)
    quantity=$((500 + RANDOM % 500))
    
    ORACLE_DATA="{\"CarrierType\":\"electricity\",\"Zone\":\"DE-LU\",\"PeriodStart\":$((now - 3600)),\"PeriodEnd\":$now,\"ProductionMethod\":\"solar_pv\",\"EnergyUnit\":\"MWh\",\"Quantity\":$quantity,\"EmissionFactor\":0,\"DataSource\":\"ENTSO-E-TP\",\"Attributes\":{}}"
    
    # Measure transaction latency
    tx_start=$(date +%s%N)
    
    # Execute transaction (suppress output for cleaner metrics)
    if /root/hlf-go/repo/fabric-bin/bin/peer chaincode invoke \
        -o localhost:7050 \
        --ordererTLSHostnameOverride orderer1.go-platform.com \
        --tls --cafile "$ORDERER_CA" \
        -C electricity-de -n golifecycle \
        -c "{\"function\":\"$ORACLE_FUNCTION\",\"Args\":[]}" \
        --transient "{\"OracleData\":\"$(echo -n "$ORACLE_DATA" | base64 -w0)\"}" \
        > /dev/null 2>&1; then
        
        tx_end=$(date +%s%N)
        latency_ns=$((tx_end - tx_start))
        latency_ms=$((latency_ns / 1000000))
        
        latencies+=($latency_ms)
        total_latency=$((total_latency + latency_ms))
        success_count=$((success_count + 1))
        
        # Update min/max
        if [ $latency_ms -lt $min_latency ]; then
            min_latency=$latency_ms
        fi
        if [ $latency_ms -gt $max_latency ]; then
            max_latency=$latency_ms
        fi
        
        # Progress indicator
        if [ $((i % 10)) -eq 0 ]; then
            echo -e "${GREEN}✓${NC} Completed $i/$NUM_TRANSACTIONS transactions (last: ${latency_ms}ms)"
        fi
    else
        fail_count=$((fail_count + 1))
        echo -e "${YELLOW}✗${NC} Transaction $i failed"
    fi
done

# Record end time
end_time=$(date +%s%N)

# Calculate metrics
total_time_ns=$((end_time - start_time))
total_time_s=$((total_time_ns / 1000000000))
total_time_ms=$((total_time_ns / 1000000))

# TPS calculation
if [ $total_time_s -gt 0 ]; then
    tps=$((success_count / total_time_s))
else
    # For very fast execution, use milliseconds
    tps=$(echo "scale=2; $success_count * 1000 / $total_time_ms" | bc)
fi

# Average latency
if [ $success_count -gt 0 ]; then
    avg_latency=$((total_latency / success_count))
else
    avg_latency=0
fi

# Success rate
success_rate=$(echo "scale=2; $success_count * 100 / $NUM_TRANSACTIONS" | bc)

echo ""
echo "========================================"
echo "Performance Results"
echo "========================================"
echo ""
echo "Transaction Summary:"
echo "  Total Transactions:  $NUM_TRANSACTIONS"
echo "  Successful:          $success_count (${success_rate}%)"
echo "  Failed:              $fail_count"
echo "  Total Time:          ${total_time_ms}ms (${total_time_s}s)"
echo ""
echo "Performance Metrics:"
echo "  Throughput (TPS):    $tps"
echo "  Avg Latency:         ${avg_latency}ms"
echo "  Min Latency:         ${min_latency}ms"
echo "  Max Latency:         ${max_latency}ms"
echo ""

# Create results table
echo "========================================"
echo "Results Table (Caliper-compatible format)"
echo "========================================"
echo ""
printf "| %-25s | %-4s | %-4s | %-15s | %-17s | %-17s | %-17s | %-16s |\n" \
    "Name" "Succ" "Fail" "Send Rate (TPS)" "Max Latency (s)" "Min Latency (s)" "Avg Latency (s)" "Throughput (TPS)"
echo "|---------------------------|------|------|-----------------|------------------|------------------|------------------|------------------|"

printf "| %-25s | %-4d | %-4d | %-15s | %-17s | %-17s | %-17s | %-16s |\n" \
    "$ORACLE_FUNCTION" \
    $success_count \
    $fail_count \
    "$tps" \
    "$(echo "scale=3; $max_latency / 1000" | bc)" \
    "$(echo "scale=3; $min_latency / 1000" | bc)" \
    "$(echo "scale=3; $avg_latency / 1000" | bc)" \
    "$tps"

echo ""
echo "========================================"
echo "Benchmark Complete"
echo "========================================"

# Verification query
echo ""
echo "Running verification query..."
RESULT=$(/root/hlf-go/repo/fabric-bin/bin/peer chaincode query \
    -C electricity-de -n golifecycle \
    -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null | head -20)

if [ -n "$RESULT" ]; then
    echo -e "${GREEN}✓${NC} Ledger query successful"
    echo "Sample records created (showing first 20 chars): ${RESULT:0:100}..."
else
    echo -e "${YELLOW}⚠${NC} Ledger query returned empty (expected if no GOs exist)"
fi

echo ""
echo -e "${GREEN}Performance benchmark completed successfully!${NC}"
