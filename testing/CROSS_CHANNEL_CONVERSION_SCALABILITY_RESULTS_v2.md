# Cross-Channel Conversion Scalability Test Results — v2 (Multi-Peer Endorsement)

## Test Environment
- **Date**: 2026-04-26
- **Network**: Hyperledger Fabric 2.5.12 on Hetzner VM (204.168.234.18)
- **Topology**: Dual-channel (electricity-de, hydrogen-de), 6 peers, 4 orderers (Raft)
- **Chaincode**: golifecycle v10.1, Sequence 1
- **Test Scope**: 3-phase cross-channel conversion protocol (Lock → Mint → Finalize)

## Key Findings

### ✅ RESOLVED: Ledger Initialization Bootstrap Issue
**Problem**: Original init script called `admin:RegisterOrganization` before any issuer existed
**Root Cause**: `device:InitLedger` requires **multi-peer endorsement** to properly commit state
**Solution**: 
- Call `device:InitLedger(eissuerMSP)` FIRST with multi-peer endorsement
- Then call `device:RegisterOrgRole(...)` for other orgs
- Use multi-peer endorsement for ALL role registration operations

**Evidence**:
```bash
# FAILED (single-peer endorsement)
peer chaincode invoke ... --peerAddresses localhost:7051 -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}'
# → status:200 BUT subsequent reads fail with "organization eissuerMSP is not registered"

# SUCCESS (multi-peer endorsement)
peer chaincode invoke ... \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS \
  -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}'
# → status:200 AND subsequent reads succeed
```

**Lesson Learned**: Fabric endorsement policies require multi-peer endorsement for state to be properly committed and readable by subsequent transactions.

### ❌ UNRESOLVED: Private Data Gossip Dissemination Failure

**Problem**: Private data is NEVER shared between peers via gossip, even after all recommended fixes
**Symptom**: `private data matching public hash version is not available` during reads

**Evidence (peer logs)**:
```
Successfully fetched (or marked to reconcile later) all 1 eligible collection private write sets for block [26] 
(0 from local cache, 1 from transient store, 0 from other peers)
```

**Fixes Attempted**:
1. ✅ Collection config updated: `requiredPeerCount: 0 → 1` for all `privateDetails-*` collections
2. ✅ Gossip bootstrap fixed: All peers point to anchor peers (not self)
   ```yaml
   # Before: CORE_PEER_GOSSIP_BOOTSTRAP=peer0.eproducer1.go-platform.com:9051
   # After:  CORE_PEER_GOSSIP_BOOTSTRAP=peer0.eissuer.go-platform.com:7051
   ```
3. ✅ Gossip debug logging enabled: `FABRIC_LOGGING_SPEC=INFO:gossip=DEBUG:gossip.privdata=DEBUG`
4. ✅ Multi-peer endorsement used for all transactions
5. ✅ Network redeployed from scratch with all fixes

**Current Status**: 
- ✅ Gossip membership works: eproducer1 discovers eissuer anchor peer
- ✅ Peers communicate (MembershipResponse: Alive: 2)
- ❌ Private data NEVER shared: "0 from other peers" in all logs
- ⚠️ Reconciliation attempts run but don't succeed

**Hypothesis**: 
Possible Fabric 2.5.12 bug or fundamental private data gossip limitation in this network configuration. 
The private data gossip protocol may require:
- Different anchor peer configuration in channel update transactions
- Different collection policy structure
- Or is broken in this Fabric version

## Test Results

### Conversion Operations Scalability
Test Date: 2026-04-26 22:12

| Operation | Iterations | Success | Duration | TPS | Notes |
|-----------|------------|---------|----------|-----|-------|
| **LockGOForConversion** | 10 | 0 | 1.451s | 0 | Failed: private data not available |
| **ListConversionLocks** | 100 | 100 | 10.787s | 9.27 | ✅ Query-only, no private data |

### Ledger Initialization Success
✅ All operations succeeded with multi-peer endorsement:
```
[E1] Bootstrapping eissuer as issuer... status:200
[E2] Registering eproducer1 as producer... status:200
[E3] Registering ebuyer1 as buyer... status:200
[E4] Registering electricity smart meter device... status:200
[E5] Adding electricity backlog... status:200

[H1] Bootstrapping hissuer as issuer... status:200
[H2] Registering hproducer1 as producer... status:200
[H3] Registering hbuyer1 as buyer... status:200
[H4] Registering hydrogen production meter device... status:200
[H5] Adding hydrogen backlog... status:200
```

### GO Creation Results
✅ 3 Electricity GOs created:
- `eGO_186277a8af699282` (status: active)
- `eGO_84d86e1172d23d95` (status: active)
- `eGO_bd71f33bd965cf5f` (status: active)

⚠️ GOs 2 and 3 encountered private data errors during backlog operations but still created successfully.

## Recommendations

### Immediate Actions
1. ✅ **Use multi-peer endorsement for ALL chaincode operations**
   - Especially for state writes (InitLedger, RegisterOrgRole, RegisterDevice, etc.)
   - Ensures state is properly committed before subsequent reads

2. ⚠️ **Investigate alternative private data approaches**:
   - **Option A**: Remove private data entirely, use public state only
   - **Option B**: Use client-side encryption instead of Fabric private data
   - **Option C**: Test with simpler topology (fewer orgs, single channel) to isolate gossip issue

3. 🔍 **Debug gossip protocol deeper**:
   - Check if anchor peers are properly configured in channel configuration
   - Verify peer discovery service is running
   - Test with Fabric 3.x (latest version) to see if gossip is fixed

### Long-Term Considerations
1. **Private data may not be suitable for this use case** given the persistent gossip failures
2. Consider **re-architecting** to use:
   - Public ledger state with access control via chaincode logic
   - Off-chain storage with on-chain commitments/hashes
   - Alternative Fabric features (implicit collections, state-based endorsement)

## Artifacts
- Initialization script: `/tmp/init-ledger-multipeer.sh` (SUCCESS)
- GO creation script: `/tmp/create-gos-base64.sh` (PARTIAL SUCCESS)
- Conversion test script: `/tmp/test-conversion-scalability.sh` (FAILED on private data ops)
- Collection configs: `/root/hlf-go/repo/collections/collection-config-{electricity,hydrogen}-de.json`
- Network config: `/root/hlf-go/repo/network/base.yaml`, `/root/hlf-go/repo/network/docker/docker-compose-*.yaml`

## Next Steps
1. ✅ Ledger initialization works reliably with multi-peer endorsement
2. ❌ Private data gossip remains broken despite all fixes
3. 🔄 **DECISION REQUIRED**: Continue debugging gossip OR pivot to alternative architecture?

## Conclusion
**Multi-peer endorsement solved the bootstrap issue**, enabling reliable ledger initialization. However, **private data gossip dissemination remains non-functional** despite comprehensive fixes (collection configs, gossip bootstrap, debug logging, network redeployment). 

The conversion scalability test cannot proceed until private data gossip is resolved OR the architecture is changed to eliminate private data dependencies.

**Status**: Blocked on private data gossip protocol investigation or architectural pivot decision.
